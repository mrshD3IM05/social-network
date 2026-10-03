package requesthandler

import (
	"net/http"

	"sn-backend/internal/handler/common"
	"sn-backend/internal/service/followsvc"
	"sn-backend/internal/service/groupsvc"
	"sn-backend/internal/service/sessionsvc"
)

// Handler serves everything waiting for the caller to accept or decline, so
// the notifications page loads it in one request instead of three.
type Handler struct {
	Follow  *followsvc.Service
	Group   *groupsvc.Service
	Session *sessionsvc.Service
}

func New(follow *followsvc.Service, group *groupsvc.Service, session *sessionsvc.Service) *Handler {
	return &Handler{Follow: follow, Group: group, Session: session}
}

// Pending handles GET /requests: the follow requests sent to the caller, the
// group invitations they received, and the join requests to the groups they
// created. Each list is always present, empty when there is nothing.
func (h *Handler) Pending(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	follows, err := h.Follow.PendingRequests(userID, userID)
	if err != nil {
		http.Error(w, "could not list follow requests", http.StatusInternalServerError)
		return
	}
	invitations, err := h.Group.PendingInvitations(userID)
	if err != nil {
		http.Error(w, "could not list group invitations", http.StatusInternalServerError)
		return
	}
	joins, err := h.Group.MyJoinRequests(userID)
	if err != nil {
		http.Error(w, "could not list join requests", http.StatusInternalServerError)
		return
	}

	followList := make([]map[string]any, 0, len(follows))
	for _, request := range follows {
		followList = append(followList, map[string]any{"id": request.ID, "created_at": request.CreatedAt, "user": common.PublicUser(request.From)})
	}
	common.WriteJSON(w, http.StatusOK, map[string]any{
		"follow_requests":     followList,
		"group_invitations":   invitations,
		"group_join_requests": joins,
	})
}
