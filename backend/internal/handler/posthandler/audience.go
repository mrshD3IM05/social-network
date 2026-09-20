package posthandler

import (
	"errors"
	"net/http"

	"sn-backend/internal/handler/common"
	"sn-backend/internal/service/postsvc"
	"sn-backend/internal/service/sessionsvc"
)

// AudienceHandler exposes the chosen-follower list of a "private" post. Until
// this existed the post_visibility table was never written, so a private post
// reached nobody but its author.
type AudienceHandler struct {
	Service *postsvc.AudienceService
	Session *sessionsvc.Service
}

func NewAudience(service *postsvc.AudienceService, session *sessionsvc.Service) *AudienceHandler {
	return &AudienceHandler{Service: service, Session: session}
}

func (h *AudienceHandler) Set(w http.ResponseWriter, r *http.Request) {
	authorID, ok := h.caller(w, r)
	if !ok {
		return
	}
	postID, err := common.PathID(r, "id")
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	audience, err := h.Service.Set(authorID, postID, common.FormIDs(r.Form["user_ids"]))
	if err != nil {
		writeAudienceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, map[string]any{"user_ids": audience})
}

func (h *AudienceHandler) Get(w http.ResponseWriter, r *http.Request) {
	authorID, ok := h.caller(w, r)
	if !ok {
		return
	}
	postID, err := common.PathID(r, "id")
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}
	audience, err := h.Service.Get(authorID, postID)
	if err != nil {
		writeAudienceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, map[string]any{"user_ids": audience})
}

func (h *AudienceHandler) caller(w http.ResponseWriter, r *http.Request) (int64, bool) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return 0, false
	}
	return userID, true
}

func writeAudienceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, postsvc.ErrNotFound):
		http.Error(w, "post not found", http.StatusNotFound)
	case errors.Is(err, postsvc.ErrNotOwner):
		http.Error(w, "not the author of this post", http.StatusForbidden)
	case errors.Is(err, postsvc.ErrInvalidPrivacy):
		http.Error(w, "only private posts have a chosen audience", http.StatusBadRequest)
	default:
		http.Error(w, "could not update the audience", http.StatusInternalServerError)
	}
}
