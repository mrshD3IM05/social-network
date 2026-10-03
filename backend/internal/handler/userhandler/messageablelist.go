package userhandler

import (
	"net/http"
	"sn-backend/internal/handler/common"
)

// Contacts handles GET /contacts: the people the caller can start a private
// conversation with, which is what the Messages list shows. The rule is the one
// /messages and /ws apply before accepting a message (CanMessage): at least one
// of the two follows the other, accepted. Anyone outside it would only get a 403
// from /messages/{id}, so they are left out here.
func (h *Handler) Contacts(w http.ResponseWriter, r *http.Request) {
	viewerID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	users, err := h.Follow.Messageable(viewerID, common.LastID(r))
	if err != nil {
		http.Error(w, "could not list contacts", http.StatusInternalServerError)
		return
	}
	writePeople(w, users)
}

// Suggestions handles GET /users/suggestions: a few people the caller has no
// accepted follow with either way, for the "People you may know" panel.
func (h *Handler) Suggestions(w http.ResponseWriter, r *http.Request) {
	viewerID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	users, err := h.Follow.Suggestions(viewerID)
	if err != nil {
		http.Error(w, "could not list suggestions", http.StatusInternalServerError)
		return
	}
	writePeople(w, users)
}
