package messagehandler

import (
	"errors"
	"net/http"
	"time"

	"sn-backend/internal/handler/common"
	"sn-backend/internal/service/messagesvc"
	"sn-backend/internal/service/sessionsvc"
)

type Handler struct {
	Service *messagesvc.Service
	Session *sessionsvc.Service
}

func New(service *messagesvc.Service, session *sessionsvc.Service) *Handler {
	return &Handler{Service: service, Session: session}
}

// Conversation replays the stored history of a private chat. Messages were
// already persisted; until now nothing could read them back.
func (h *Handler) Conversation(w http.ResponseWriter, r *http.Request) {
	viewerID, ok := h.caller(w, r)
	if !ok {
		return
	}
	otherID, err := common.PathID(r, "id")
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	var before *time.Time
	if value := r.URL.Query().Get("before"); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			http.Error(w, "before must be an RFC3339 timestamp", http.StatusBadRequest)
			return
		}
		before = &parsed
	}
	messages, err := h.Service.Conversation(viewerID, otherID, common.QueryInt(r, "limit", 50), before)
	if err != nil {
		if errors.Is(err, messagesvc.ErrNotAllowed) {
			http.Error(w, "you cannot message this user", http.StatusForbidden)
		} else {
			http.Error(w, "could not load the conversation", http.StatusInternalServerError)
		}
		return
	}
	common.WriteJSON(w, http.StatusOK, messages)
}

// Inbox lists the people the caller already talked to, newest first.
func (h *Handler) Inbox(w http.ResponseWriter, r *http.Request) {
	viewerID, ok := h.caller(w, r)
	if !ok {
		return
	}
	conversations, err := h.Service.Inbox(viewerID)
	if err != nil {
		http.Error(w, "could not load conversations", http.StatusInternalServerError)
		return
	}
	payload := make([]map[string]any, 0, len(conversations))
	for _, conversation := range conversations {
		payload = append(payload, map[string]any{
			"user":         common.PublicUser(conversation.User),
			"last_message": conversation.LastMessage,
		})
	}
	common.WriteJSON(w, http.StatusOK, payload)
}

func (h *Handler) caller(w http.ResponseWriter, r *http.Request) (int64, bool) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return 0, false
	}
	return userID, true
}
