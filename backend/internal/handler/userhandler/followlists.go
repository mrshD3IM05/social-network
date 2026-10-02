package userhandler

import (
	"net/http"
	"sn-backend/internal/handler/common"
)

// Followers handles GET /users/{id}/followers and Following GET
// /users/{id}/following: the two lists a profile shows. They sit behind the same
// privacy gate as the profile itself, so a private one stays hidden.
func (h *Handler) Followers(w http.ResponseWriter, r *http.Request) {
	user, ok := h.visibleUser(w, r)
	if !ok {
		return
	}
	users, err := h.Follow.Followers(user.ID, common.LastID(r))
	if err != nil {
		http.Error(w, "could not list followers", http.StatusInternalServerError)
		return
	}
	writePeople(w, users)
}

func (h *Handler) Following(w http.ResponseWriter, r *http.Request) {
	user, ok := h.visibleUser(w, r)
	if !ok {
		return
	}
	users, err := h.Follow.Following(user.ID, common.LastID(r))
	if err != nil {
		http.Error(w, "could not list following", http.StatusInternalServerError)
		return
	}
	writePeople(w, users)
}
