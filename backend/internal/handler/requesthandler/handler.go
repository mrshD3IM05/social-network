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

// Pending handles GET /requests: the first page (10) of each list waiting for
// the caller — follow requests sent to them, group invitations they received,
// join requests to the groups they created. Each list is always present.
//
// GET /requests?type=follow_requests|group_invitations|group_join_requests&last=<id>
// answers the next page of that one list only, as a plain array.
func (h *Handler) Pending(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	last := common.LastID(r)

	lists := map[string]func() (any, error){
		"follow_requests": func() (any, error) {
			follows, err := h.Follow.PendingRequests(userID, userID, last)
			if err != nil {
				return nil, err
			}
			list := make([]map[string]any, 0, len(follows))
			for _, request := range follows {
				list = append(list, map[string]any{"id": request.ID, "created_at": request.CreatedAt, "user": common.PublicUser(request.From)})
			}
			return list, nil
		},
		"group_invitations": func() (any, error) {
			return h.Group.PendingInvitations(userID, last)
		},
		"group_join_requests": func() (any, error) {
			return h.Group.MyJoinRequests(userID, last)
		},
	}

	if kind := r.URL.Query().Get("type"); kind != "" {
		load, ok := lists[kind]
		if !ok {
			http.Error(w, "unknown request type", http.StatusBadRequest)
			return
		}
		list, err := load()
		if err != nil {
			http.Error(w, "could not list requests", http.StatusInternalServerError)
			return
		}
		common.WriteJSON(w, http.StatusOK, list)
		return
	}

	all := make(map[string]any, len(lists))
	for kind, load := range lists {
		list, err := load()
		if err != nil {
			http.Error(w, "could not list requests", http.StatusInternalServerError)
			return
		}
		all[kind] = list
	}
	common.WriteJSON(w, http.StatusOK, all)
}
