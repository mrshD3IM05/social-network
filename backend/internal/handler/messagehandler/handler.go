package messagehandler

import (
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"sn-backend/internal/handler/common"
	"sn-backend/internal/model"
	"sn-backend/internal/service/filesvc"
	"sn-backend/internal/service/messagesvc"
	"sn-backend/internal/service/sessionsvc"
)

// Publisher is the websocket hub: a sent message is pushed to both sides.
type Publisher interface {
	PublishMessage(*model.Message)
}

type Handler struct {
	Service   *messagesvc.Service
	Files     *filesvc.Service
	Session   *sessionsvc.Service
	WebSocket Publisher
}

func New(service *messagesvc.Service, files *filesvc.Service, session *sessionsvc.Service, webSocket Publisher) *Handler {
	return &Handler{Service: service, Files: files, Session: session, WebSocket: webSocket}
}

// History answers with the stored conversation with one user.
// Messages were always saved; until now nothing could read them back.
func (h *Handler) History(w http.ResponseWriter, r *http.Request) {
	viewerID, ok := h.caller(w, r)
	if !ok {
		return
	}
	otherID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || otherID < 1 {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	messages, err := h.Service.History(viewerID, otherID, limit)
	if err != nil {
		writeError(w, err, "could not load the conversation")
		return
	}
	common.WriteJSON(w, http.StatusOK, messages)
}

// Send saves a private message and pushes it to both users.
//
// Sending over HTTP (and not over the websocket) is what lets a message carry
// images: the message row has to exist before an upload can point at it, and
// both steps finish before anybody is told about the message.
func (h *Handler) Send(w http.ResponseWriter, r *http.Request) {
	fromID, ok := h.caller(w, r)
	if !ok {
		return
	}

	headers, err := readForm(w, r)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if len(headers) > filesvc.MaxImages {
		http.Error(w, filesvc.ErrTooManyImages.Error(), http.StatusBadRequest)
		return
	}

	toID, err := strconv.ParseInt(r.FormValue("to_user_id"), 10, 64)
	if err != nil || toID < 1 {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	message, err := h.Service.Send(fromID, toID, r.FormValue("content"), len(headers) > 0)
	if err != nil {
		writeError(w, err, "could not send the message")
		return
	}

	if len(headers) > 0 {
		if _, err := h.Files.UploadMany(fromID, headers, nil, &message.ID, nil); err != nil {
			http.Error(w, err.Error(), uploadStatus(err))
			return
		}
		if err := h.Service.LoadImages(message); err != nil {
			http.Error(w, "could not load the images", http.StatusInternalServerError)
			return
		}
	}

	h.WebSocket.PublishMessage(message)
	common.WriteJSON(w, http.StatusCreated, message)
}

func (h *Handler) caller(w http.ResponseWriter, r *http.Request) (int64, bool) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return 0, false
	}
	return userID, true
}

// readForm accepts a plain form, or a multipart one when images are attached.
func readForm(w http.ResponseWriter, r *http.Request) ([]*multipart.FileHeader, error) {
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
	case errors.Is(err, messagesvc.ErrNotAllowed):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, messagesvc.ErrEmpty), errors.Is(err, messagesvc.ErrTooLong):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, fallback, http.StatusInternalServerError)
	}
}
