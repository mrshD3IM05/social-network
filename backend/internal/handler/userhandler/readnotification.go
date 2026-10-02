package userhandler

import (
	"net/http"
	"sn-backend/internal/handler/common"
)

// ReadNotifications handles POST /notifications/read: marks them all as seen.
func (h *Handler) ReadNotifications(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	if err := h.Service.MarkNotificationsRead(userID); err != nil {
		http.Error(w, "could not update notifications", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UnreadNotifications handles GET /notifications/unread: {"count": n}, the
// number on the bell in the sidebar.
func (h *Handler) UnreadNotifications(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	count, err := h.Service.UnreadNotifications(userID)
	if err != nil {
		http.Error(w, "could not count notifications", http.StatusInternalServerError)
		return
	}
	common.WriteJSON(w, http.StatusOK, map[string]int{"count": count})
}
