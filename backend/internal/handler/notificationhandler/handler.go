package notificationhandler

import (
	"errors"
	"net/http"

	"sn-backend/internal/handler/common"
	"sn-backend/internal/service/notificationsvc"
	"sn-backend/internal/service/sessionsvc"
)

type Handler struct {
	Service *notificationsvc.Service
	Session *sessionsvc.Service
}

func New(service *notificationsvc.Service, session *sessionsvc.Service) *Handler {
	return &Handler{Service: service, Session: session}
}

// List answers with the stored history so a notification that arrived while the
// user was on another page is not lost.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.caller(w, r)
	if !ok {
		return
	}
	notifications, err := h.Service.List(userID, common.QueryInt(r, "limit", 50))
	if err != nil {
		http.Error(w, "could not list notifications", http.StatusInternalServerError)
		return
	}
	unread, err := h.Service.UnreadCount(userID)
	if err != nil {
		http.Error(w, "could not count notifications", http.StatusInternalServerError)
		return
	}
	common.WriteJSON(w, http.StatusOK, map[string]any{"notifications": notifications, "unread": unread})
}

func (h *Handler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.caller(w, r)
	if !ok {
		return
	}
	unread, err := h.Service.UnreadCount(userID)
	if err != nil {
		http.Error(w, "could not count notifications", http.StatusInternalServerError)
		return
	}
	common.WriteJSON(w, http.StatusOK, map[string]any{"unread": unread})
}

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.caller(w, r)
	if !ok {
		return
	}
	notificationID, err := common.PathID(r, "id")
	if err != nil {
		http.Error(w, "invalid notification id", http.StatusBadRequest)
		return
	}
	if err := h.Service.MarkRead(userID, notificationID); err != nil {
		if errors.Is(err, notificationsvc.ErrNotFound) {
			http.Error(w, "notification not found", http.StatusNotFound)
		} else {
			http.Error(w, "could not update notification", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.caller(w, r)
	if !ok {
		return
	}
	if err := h.Service.MarkAllRead(userID); err != nil {
		http.Error(w, "could not update notifications", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) caller(w http.ResponseWriter, r *http.Request) (int64, bool) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return 0, false
	}
	return userID, true
}
