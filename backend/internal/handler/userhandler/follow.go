package userhandler

import (
	"net/http"
	"sn-backend/internal/handler/common"
	"sn-backend/internal/model"
	"sn-backend/internal/service/followsvc"
	"strings"
)

func (h *Handler) RespondFollow(w http.ResponseWriter, r *http.Request) {
	viewerID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	requestID, err := common.PathID(r, "id")
	if err != nil {
		http.Error(w, "invalid follow request id", http.StatusBadRequest)
		return
	}
	status := model.FollowAccepted
	if strings.HasSuffix(r.URL.Path, "/decline") {
		status = model.FollowDeclined
	}
	if err := h.Follow.Respond(viewerID, requestID, status); err != nil {
		if err == followsvc.ErrNotRecipient {
			http.Error(w, "not the follow request recipient", http.StatusForbidden)
		} else {
			http.Error(w, "could not respond to follow request", http.StatusConflict)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) FollowUser(w http.ResponseWriter, r *http.Request) {
	viewerID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	targetID, err := common.PathID(r, "id")
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	follow, err := h.Follow.Follow(viewerID, targetID)
	if err != nil {
		if err == followsvc.ErrCannotFollowSelf || err == followsvc.ErrExists {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, "could not follow user", http.StatusInternalServerError)
		}
		return
	}
	common.WriteJSON(w, http.StatusCreated, follow)
}

func (h *Handler) UnfollowUser(w http.ResponseWriter, r *http.Request) {
	viewerID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	targetID, err := common.PathID(r, "id")
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	if err := h.Follow.Unfollow(viewerID, targetID); err != nil {
		http.Error(w, "could not unfollow user", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
