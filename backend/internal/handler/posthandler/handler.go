package posthandler

import (
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
	viewers, err := formIDs(r, "viewers")
	if err != nil {
		http.Error(w, "invalid viewers", http.StatusBadRequest)
		return
	}
	post, err := h.Service.Create(userID, r.FormValue("content"), r.FormValue("privacy"), viewers)
	if err != nil {
		if err == postsvc.ErrInvalidPrivacy || err == postsvc.ErrInvalidContent || err == postsvc.ErrInvalidViewers {
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
	posts, err := h.Service.ListVisible(viewerID, common.LastID(r))
	if err != nil {
		http.Error(w, "could not list posts", http.StatusInternalServerError)
		return
	}
	common.WriteJSON(w, http.StatusOK, posts)
}

// GetPost handles GET /posts/{id}. Visibility follows CanViewPost: privacy
// rules for normal posts, group membership for group posts — invisible posts
// answer 404 like the reaction endpoints.
func (h *Handler) GetPost(w http.ResponseWriter, r *http.Request) {
	viewerID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	id, err := common.PathID(r, "id")
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}
	post, err := h.Service.Get(viewerID, id)
	if err != nil {
		if err == postsvc.ErrNotFound {
			http.Error(w, "post not found", http.StatusNotFound)
		} else {
			http.Error(w, "could not get post", http.StatusInternalServerError)
		}
		return
	}
	common.WriteJSON(w, http.StatusOK, post)
}
func (h *Handler) ReactionPost(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	id, err := common.PathID(r, "id")
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	summary, err := h.Service.React(userID, id, r.FormValue("reaction"))
	if err != nil {
		writeReactionError(w, err, "could not react to post")
		return
	}
	common.WriteJSON(w, http.StatusOK, summary)
}
func (h *Handler) DeleteReaction(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	id, err := common.PathID(r, "id")
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}
	summary, err := h.Service.Unreact(userID, id)
	if err != nil {
		writeReactionError(w, err, "could not remove reaction")
		return
	}
	common.WriteJSON(w, http.StatusOK, summary)
}

func writeReactionError(w http.ResponseWriter, err error, fallback string) {
	if err == postsvc.ErrInvalidReaction {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else if err == postsvc.ErrNotFound {
		http.Error(w, "post not found", http.StatusNotFound)
	} else {
		http.Error(w, fallback, http.StatusInternalServerError)
	}
}

// ListViewers handles GET /posts/{id}/viewers: the followers a private post was
// shared with. Only the author gets an answer.
func (h *Handler) ListViewers(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	id, err := common.PathID(r, "id")
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}
	viewers, err := h.Service.Viewers(userID, id)
	if err != nil {
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}
	common.WriteJSON(w, http.StatusOK, viewers)
}

func (h *Handler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	id, err := common.PathID(r, "id")
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	viewers, err := formIDs(r, "viewers")
	if err != nil {
		http.Error(w, "invalid viewers", http.StatusBadRequest)
		return
	}
	post, err := h.Service.Update(userID, id, r.FormValue("content"), r.FormValue("privacy"), viewers)
	if err != nil {
		if err == postsvc.ErrInvalidPrivacy || err == postsvc.ErrInvalidContent || err == postsvc.ErrInvalidViewers {
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
	id, err := common.PathID(r, "id")
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

// formIDs reads a list of user ids sent as the same form field repeated
// (viewers=2&viewers=5).
func formIDs(r *http.Request, name string) ([]int64, error) {
	ids := []int64{}
	for _, value := range r.Form[name] {
		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil || id < 1 {
			return nil, strconv.ErrSyntax
		}
		ids = append(ids, id)
	}
	return ids, nil
}
