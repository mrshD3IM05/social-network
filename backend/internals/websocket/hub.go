package websocket

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"sn-backend/internals/model"
	"sn-backend/internals/service/sessionsvc"
)

var ErrInvalidMessage = errors.New("websocket: invalid message")

type Repository interface {
	CreateMessage(*model.Message) error
	CanMessage(int64, *int64, *int64) (bool, error)
	GroupMemberIDs(int64) ([]int64, error)
}

type Hub struct {
	mu      sync.RWMutex
	clients map[int64]map[*Client]struct{}
	repo    Repository
}

func NewHub(repo Repository) *Hub {
	return &Hub{clients: make(map[int64]map[*Client]struct{}), repo: repo}
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request, sessions *sessionsvc.Service) {
	cookie, err := r.Cookie(sessionsvc.CookieName)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	session, err := sessions.Get(cookie.Value)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	connection, err := (&websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}).Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := &Client{hub: h, connection: connection, userID: session.UserID, send: make(chan []byte, 16)}
	h.add(client)
	go client.writePump()
	client.readPump()
}

func (h *Hub) PublishNotification(notification *model.Notification) {
	h.publish(notification.UserID, map[string]any{"type": "notification", "notification": notification})
}

func (h *Hub) add(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[client.userID] == nil {
		h.clients[client.userID] = make(map[*Client]struct{})
	}
	h.clients[client.userID][client] = struct{}{}
}

func (h *Hub) remove(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if clients := h.clients[client.userID]; clients != nil {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.clients, client.userID)
		}
	}
}

func (h *Hub) publish(userID int64, event any) {
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients[userID] {
		select {
		case client.send <- payload:
		default:
		}
	}
}

type Client struct {
	hub        *Hub
	connection *websocket.Conn
	userID     int64
	send       chan []byte
}

type incomingMessage struct {
	Type    string `json:"type"`
	ToUser  *int64 `json:"to_user_id,omitempty"`
	GroupID *int64 `json:"group_id,omitempty"`
	Content string `json:"content"`
}

func (c *Client) readPump() {
	defer func() { c.hub.remove(c); c.connection.Close() }()
	c.connection.SetReadLimit(64 << 10)
	_ = c.connection.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.connection.SetPongHandler(func(string) error { return c.connection.SetReadDeadline(time.Now().Add(60 * time.Second)) })
	for {
		var input incomingMessage
		if err := c.connection.ReadJSON(&input); err != nil {
			return
		}
		if input.Type != "message" || input.Content == "" || (input.ToUser == nil) == (input.GroupID == nil) {
			c.sendError(ErrInvalidMessage.Error())
			continue
		}
		allowed, err := c.hub.repo.CanMessage(c.userID, input.ToUser, input.GroupID)
		if err != nil || !allowed {
			c.sendError("message is not permitted")
			continue
		}
		message := &model.Message{FromUserID: c.userID, ToUserID: input.ToUser, GroupID: input.GroupID, Content: input.Content}
		if err := c.hub.repo.CreateMessage(message); err != nil {
			c.sendError("could not save message")
			continue
		}
		event := map[string]any{"type": "message", "message": message}
		if input.ToUser != nil {
			c.hub.publish(*input.ToUser, event)
			c.hub.publish(c.userID, event)
		} else if members, err := c.hub.repo.GroupMemberIDs(*input.GroupID); err == nil {
			for _, memberID := range members {
				c.hub.publish(memberID, event)
			}
		}
	}
}

func (c *Client) sendError(message string) {
	payload, _ := json.Marshal(map[string]string{"type": "error", "error": message})
	select {
	case c.send <- payload:
	default:
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(45 * time.Second)
	defer func() { ticker.Stop(); c.connection.Close() }()
	for {
		select {
		case payload, ok := <-c.send:
			_ = c.connection.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok || c.connection.WriteMessage(websocket.TextMessage, payload) != nil {
				return
			}
		case <-ticker.C:
			_ = c.connection.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if c.connection.WriteMessage(websocket.PingMessage, nil) != nil {
				return
			}
		}
	}
}
