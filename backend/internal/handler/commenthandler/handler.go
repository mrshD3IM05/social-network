package commenthandler

import (
	"errors"
	"mime/multipart"
	"net/http"
	"strings"

	"sn-backend/internal/handler/common"
	"sn-backend/internal/service/commentsvc"
	"sn-backend/internal/service/filesvc"
	"sn-backend/internal/service/sessionsvc"
)

type Handler struct {
	Service *commentsvc.Service
	Files   *filesvc.Service
	Session *sessionsvc.Service
}

func New(service *commentsvc.Service, files *filesvc.Service, session *sessionsvc.Service) *Handler {
	return &Handler{Service: service, Files: files, Session: session}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	viewerID, ok := h.caller(w, r)
	if !ok {
		return
	}
	postID, err := common.PathID(r, "id")
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}
	comments, err := h.Service.List(viewerID, postID)
	if err != nil {
		writeError(w, err, "could not list comments")
		return
	}
	common.WriteJSON(w, http.StatusOK, comments)
}

// Create accepts either a plain form or a multipart form carrying up to three
// images, mirroring how a post is written.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	authorID, ok := h.caller(w, r)
	if !ok {
		return
	}
	postID, err := common.PathID(r, "id")
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}

	headers, err := parseBody(w, r)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if len(headers) > filesvc.MaxImages {
		http.Error(w, filesvc.ErrTooManyImages.Error(), http.StatusBadRequest)
		return
	}

	comment, err := h.Service.Create(authorID, postID, r.FormValue("content"))
	if err != nil {
		writeError(w, err, "could not create comment")
		return
	}

	if len(headers) > 0 {
		files, err := h.Files.UploadMany(authorID, headers, nil, nil)
		if err != nil {
			// The comment itself is saved; report the upload problem only.
			http.Error(w, err.Error(), uploadStatus(err))
			return
		}
		for _, file := range files {
			if err := h.Service.AttachImage(authorID, comment.ID, file.ID); err != nil {
				http.Error(w, "could not attach image", http.StatusInternalServerError)
				return
			}
		}
		if comment, err = h.Service.Get(authorID, comment.ID); err != nil {
			writeError(w, err, "could not load comment")
			return
		}
	}

	common.WriteJSON(w, http.StatusCreated, comment)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.caller(w, r)
	if !ok {
		return
	}
	commentID, err := common.PathID(r, "id")
	if err != nil {
		http.Error(w, "invalid comment id", http.StatusBadRequest)
		return
	}
	fileIDs, err := h.Service.Delete(userID, commentID)
	if err != nil {
		writeError(w, err, "could not delete comment")
		return
	}
	h.Files.RemoveStored(fileIDs)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) React(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.caller(w, r)
	if !ok {
		return
	}
	commentID, err := common.PathID(r, "id")
	if err != nil {
		http.Error(w, "invalid comment id", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	summary, err := h.Service.React(userID, commentID, r.FormValue("reaction"))
	if err != nil {
		writeError(w, err, "could not react to comment")
		return
	}
	common.WriteJSON(w, http.StatusOK, summary)
}

func (h *Handler) Unreact(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.caller(w, r)
	if !ok {
		return
	}
	commentID, err := common.PathID(r, "id")
	if err != nil {
		http.Error(w, "invalid comment id", http.StatusBadRequest)
		return
	}
	summary, err := h.Service.Unreact(userID, commentID)
	if err != nil {
		writeError(w, err, "could not remove reaction")
		return
	}
	common.WriteJSON(w, http.StatusOK, summary)
}

func (h *Handler) caller(w http.ResponseWriter, r *http.Request) (int64, bool) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return 0, false
	}
	return userID, true
}

func parseBody(w http.ResponseWriter, r *http.Request) ([]*multipart.FileHeader, error) {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		return nil, r.ParseForm()
	}
	limit := filesvc.MaxImageSize*filesvc.MaxImages + 1<<20
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	if err := r.ParseMultipartForm(limit); err != nil {
		return nil, err
	}
	if r.MultipartForm == nil {
		return nil, nil
	}
	return r.MultipartForm.File["files"], nil
}

func uploadStatus(err error) int {
	if errors.Is(err, filesvc.ErrInvalidImage) || errors.Is(err, filesvc.ErrFileTooLarge) || errors.Is(err, filesvc.ErrTooManyImages) {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

func writeError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, commentsvc.ErrNotFound):
		http.Error(w, "comment or post not found", http.StatusNotFound)
	case errors.Is(err, commentsvc.ErrEmptyContent), errors.Is(err, commentsvc.ErrTooLong), errors.Is(err, commentsvc.ErrInvalidReaction):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, fallback, http.StatusInternalServerError)
	}
}
