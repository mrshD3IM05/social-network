package commenthandler

import (
	"errors"
	"net/http"
	"strconv"

	"sn-backend/internal/handler/common"
	"sn-backend/internal/repository"
	"sn-backend/internal/service/commentsvc"
	"sn-backend/internal/service/sessionsvc"
)

type Handler struct {
	Service *commentsvc.Service
	Session *sessionsvc.Service
}

func New(service *commentsvc.Service, session *sessionsvc.Service) *Handler {
	return &Handler{Service: service, Session: session}
}

// ListComments handles GET /posts/{id}/comments. The viewer must be able to
// see the post (post privacy, or group membership for group posts).
func (h *Handler) ListComments(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	postID, err := parseID(r)
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}
	comments, err := h.Service.List(userID, postID)
	if err != nil {
		writeError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, comments)
}

// CreateComment handles POST /posts/{id}/comments with a `content` form
// field. Authorization goes through the post visibility rules in the service.
func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	postID, err := parseID(r)
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	comment, err := h.Service.Create(userID, postID, r.FormValue("content"))
	if err != nil {
		writeError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusCreated, comment)
}

// writeError maps comment errors like the other handlers. Posts the viewer
// cannot see answer 404 — the same convention as reacting to an invisible
// post — so the API does not reveal hidden posts exist.
func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, commentsvc.ErrInvalidContent):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, commentsvc.ErrNoAccess),
		errors.Is(err, commentsvc.ErrNotFound),
		errors.Is(err, repository.ErrNotFound):
		http.Error(w, "post not found", http.StatusNotFound)
	default:
		http.Error(w, "could not process comment", http.StatusInternalServerError)
	}
}

func parseID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, strconv.ErrSyntax
	}
	return id, nil
}
