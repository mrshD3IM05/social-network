package posthandler

import (
	"errors"
	"net/http"
	"sn-backend/internal/handler/common"
	"sn-backend/internal/service/postsvc"
	"sn-backend/internal/service/sessionsvc"
	"strconv"
)

type Handler struct {
	Service *postsvc.Service
	Session *sessionsvc.Service
}

func New(service *postsvc.Service, session *sessionsvc.Service) *Handler {
	return &Handler{Service: service, Session: session}
}
func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	post, err := h.Service.Create(userID, r.FormValue("content"), r.FormValue("privacy"))
	if err != nil {
		if err == postsvc.ErrInvalidPrivacy {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "could not create post", http.StatusInternalServerError)
		}
		return
	}
	common.WriteJSON(w, http.StatusCreated, post)
}
func (h *Handler) ListPosts(w http.ResponseWriter, r *http.Request) {
	viewerID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	posts, err := h.Service.ListVisible(viewerID)
	if err != nil {
		http.Error(w, "could not list posts", http.StatusInternalServerError)
		return
	}
	common.WriteJSON(w, http.StatusOK, posts)
}
func (h *Handler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	post, err := h.Service.Update(userID, id, r.FormValue("content"), r.FormValue("privacy"))
	if err != nil {
		if err == postsvc.ErrInvalidPrivacy {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else if err == postsvc.ErrNotFound {
			http.Error(w, "post not found", http.StatusNotFound)
		} else {
			http.Error(w, "could not update post", http.StatusInternalServerError)
		}
		return
	}
	common.WriteJSON(w, http.StatusOK, post)
}
func (h *Handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	id, err := parseID(r)
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}
	if err := h.Service.Delete(userID, id); err != nil {
		if err == postsvc.ErrNotFound {
			http.Error(w, "post not found", http.StatusNotFound)
		} else {
			http.Error(w, "could not delete post", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func parseID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		if err == nil {
			err = errors.New("id must be positive")
		}
		return 0, err
	}
	return id, nil
}
