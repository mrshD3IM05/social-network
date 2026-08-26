package filehandler

import (
	"net/http"
	"os"
	"sn-backend/internal/handler/common"
	"sn-backend/internal/service/filesvc"
	"sn-backend/internal/service/sessionsvc"
	"strconv"
)

type Handler struct {
	Service *filesvc.Service
	Session *sessionsvc.Service
}

func New(service *filesvc.Service, session *sessionsvc.Service) *Handler {
	return &Handler{Service: service, Session: session}
}
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	ownerID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, filesvc.MaxImageSize+1<<20)
	if err := r.ParseMultipartForm(filesvc.MaxImageSize + 1<<20); err != nil {
		http.Error(w, "upload is too large or invalid", http.StatusBadRequest)
		return
	}
	headers := r.MultipartForm.File["files"]
	if len(headers) == 0 {
		headers = r.MultipartForm.File["file"]
	}
	if len(headers) == 0 {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}
	if len(headers) > filesvc.MaxImages {
		http.Error(w, filesvc.ErrTooManyImages.Error(), http.StatusBadRequest)
		return
	}
	var postID *int64
	if value := r.FormValue("post_id"); value != "" {
		parsed, parseErr := strconv.ParseInt(value, 10, 64)
		if parseErr != nil || parsed < 1 {
			http.Error(w, "invalid post id", http.StatusBadRequest)
			return
		}
		postID = &parsed
	}
	stored, err := h.Service.UploadMany(ownerID, headers, postID)
	if err != nil {
		if err == filesvc.ErrInvalidImage || err == filesvc.ErrFileTooLarge || err == filesvc.ErrTooManyImages {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "could not store file", http.StatusInternalServerError)
		}
		return
	}
	common.WriteJSON(w, http.StatusCreated, stored)
}
func (h *Handler) SetAvatar(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, filesvc.MaxImageSize+1<<20)
	if err := r.ParseMultipartForm(filesvc.MaxImageSize + 1<<20); err != nil {
		http.Error(w, "upload is too large or invalid", http.StatusBadRequest)
		return
	}
	headers := r.MultipartForm.File["avatar"]
	if len(headers) != 1 {
		http.Error(w, "exactly one avatar image is required", http.StatusBadRequest)
		return
	}
	user, err := h.Service.SetAvatar(userID, headers[0])
	if err != nil {
		if err == filesvc.ErrInvalidImage || err == filesvc.ErrFileTooLarge {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "could not set avatar", http.StatusInternalServerError)
		}
		return
	}
	common.WriteJSON(w, http.StatusOK, common.PrivateUser(user))
}
func (h *Handler) Thumbnail(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	id := r.PathValue("id")
	file, err := h.Service.Get(id)
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	visible, err := h.Service.CanView(userID, id)
	if err != nil || !visible {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	path := h.Service.ThumbnailPath(id)
	if _, err := os.Stat(path); err != nil {
		http.Error(w, "thumbnail not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", file.MIMEType)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeFile(w, r, path)
}

func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	ownerID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	id := r.PathValue("id")
	file, err := h.Service.Get(id)
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	visible, err := h.Service.CanView(ownerID, id)
	if err != nil || !visible {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", file.MIMEType)
	http.ServeFile(w, r, file.StoragePath)
}
