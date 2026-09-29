package websocket

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/gorilla/websocket"
	"sn-backend/internal/model"
	"sn-backend/internal/repository"
	"sn-backend/internal/service/sessionsvc"
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
	connection, err := (&websocket.Upgrader{CheckOrigin: checkOrigin}).Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := &Client{hub: h, connection: connection, userID: session.UserID, send: make(chan []byte, 16)}
	h.add(client)
	trackClient(cookie.Value, client)
	go client.writePump()
	client.readPump()
}

// allowedOrigins are the pages on another host that may open the socket.
// ALLOWED_ORIGINS (comma separated) replaces the default list. Any port on
// this machine is accepted too, see checkOrigin.
var allowedOrigins = loadAllowedOrigins()

func loadAllowedOrigins() []string {
	if value := os.Getenv("ALLOWED_ORIGINS"); value != "" {
		return strings.Split(value, ",")
	}
	return []string{"http://localhost:3000", "http://127.0.0.1:3000"}
}

// checkOrigin stops another website from opening a socket with the visitor's
// cookie: the page must come from this host or from an allowed origin.
func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true // not a browser, so no cookie of someone else to borrow
	}
	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}
	if strings.EqualFold(parsed.Host, r.Host) {
		return true
	}
	// the frontend in dev runs on this machine, on whichever port is free;
	// a page that attacks a visitor is never served from their own computer
	switch parsed.Hostname() {
	case "localhost", "127.0.0.1", "::1":
		return true
	}
	for _, allowed := range allowedOrigins {
		if strings.EqualFold(strings.TrimSpace(allowed), origin) {
			return true
		}
	}
	return false
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
	Type    string `json:"type"`
	ToUser  *int64 `json:"to_user_id,omitempty"`
	GroupID *int64 `json:"group_id,omitempty"`
	Content string `json:"content"`
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
		if input.Type != "message" || input.Content == "" || utf8.RuneCountInString(input.Content) > maxContentLength || (input.ToUser == nil) == (input.GroupID == nil) {
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
