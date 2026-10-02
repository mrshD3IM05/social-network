package websocket

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"sn-backend/internal/model"
	"sn-backend/internal/repository"
	"sn-backend/internal/service/sessionsvc"

	"github.com/gorilla/websocket"
)

var ErrInvalidMessage = errors.New("websocket: invalid message")

// same limit as messagesvc.MaxContentLength (the HTTP send endpoint)
const maxContentLength = 1000

type Hub struct {
	mu       sync.RWMutex
	clients  map[int64]map[*Client]struct{}
	repo     *repository.Repository
	sessions *sessionsvc.Service
}

func NewHub(repo *repository.Repository, sessions *sessionsvc.Service) *Hub {
	return &Hub{clients: make(map[int64]map[*Client]struct{}), repo: repo, sessions: sessions}
}

// ServeHTTP upgrades GET /ws to a websocket. The Hub is a plain http.Handler, so
// it is routed like any other endpoint. The connection is tied to the session
// that opened it, which is how logging out closes the socket straight away.
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionsvc.CookieName)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	session, err := h.sessions.Get(cookie.Value)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	connection, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := &Client{hub: h, connection: connection, userID: session.UserID, send: make(chan []byte, 16)}
	h.add(client)
	trackClient(cookie.Value, client)
	go client.writePump()
	client.readPump()
}

// Notify saves the notification and sends it right away to the user's open
// pages. A notification that cannot be saved is only logged: it never makes
// the action that caused it (a follow, an invite...) fail.
func (h *Hub) Notify(notification *model.Notification) {
	if err := h.repo.CreateNotification(notification); err != nil {
		log.Printf("could not create notification: %v", err)
		return
	}
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
	Type      string `json:"type"`
	ToUser    *int64 `json:"to_user_id,omitempty"`
	GroupID   *int64 `json:"group_id,omitempty"`
	Content   string `json:"content"`
	HasImages bool   `json:"has_images,omitempty"`
	ClientID  string `json:"client_id,omitempty"`
}

func (c *Client) readPump() {
	defer func() { c.hub.remove(c); untrackClient(c); c.connection.Close() }()
	c.connection.SetReadLimit(64 << 10)
	_ = c.connection.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.connection.SetPongHandler(func(string) error { return c.connection.SetReadDeadline(time.Now().Add(60 * time.Second)) })
	for {
		var input incomingMessage
		if err := c.connection.ReadJSON(&input); err != nil {
			return
		}
		// "someone is writing" is passed on and not stored
		if input.Type == "typing" {
			c.hub.relayTyping(c.userID, input.ToUser, input.GroupID)
			continue
		}
		input.Content = strings.TrimSpace(input.Content)
		if input.Type != "message" || (input.Content == "" && !input.HasImages) || utf8.RuneCountInString(input.Content) > maxContentLength || len(input.ClientID) > 64 || (input.ToUser == nil) == (input.GroupID == nil) {
			c.sendError(ErrInvalidMessage.Error())
			continue
		}
		allowed, err := c.hub.repo.CanMessage(c.userID, input.ToUser, input.GroupID)
		if err != nil || !allowed {
			c.sendError("message is not permitted")
			continue
		}
		message := &model.Message{FromUserID: c.userID, ToUserID: input.ToUser, GroupID: input.GroupID, Content: input.Content, Images: []string{}}
		if err := c.hub.repo.CreateMessage(message); err != nil {
			c.sendError("could not save message")
			continue
		}
		if input.HasImages {
			payload, _ := json.Marshal(map[string]any{"type": "message_created", "client_id": input.ClientID, "message_id": message.ID})
			select {
			case c.send <- payload:
			default:
			}
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
