package userhandler

import (
	"net/http"
	"sn-backend/internal/handler/common"
)

// Followers handles GET /users/{id}/followers and Following GET
// /users/{id}/following: the two lists a profile shows. They are not behind the
// privacy gate the posts are — their counts already sit on the profile of
// everyone — so every row answers the public profile plus the relation the
// caller has with it.
func (h *Handler) Followers(w http.ResponseWriter, r *http.Request) {
	user, viewerID, ok := h.resolveUser(w, r)
	if !ok {
		return
	}
	users, err := h.Follow.Followers(viewerID, user.ID, common.LastID(r))
	if err != nil {
		http.Error(w, "could not list followers", http.StatusInternalServerError)
		return
	}
	writePeople(w, users)
}

func (h *Handler) Following(w http.ResponseWriter, r *http.Request) {
	user, viewerID, ok := h.resolveUser(w, r)
	if !ok {
		return
	}
	users, err := h.Follow.Following(viewerID, user.ID, common.LastID(r))
	if err != nil {
		http.Error(w, "could not list following", http.StatusInternalServerError)
		return
	}
	writePeople(w, users)
}
