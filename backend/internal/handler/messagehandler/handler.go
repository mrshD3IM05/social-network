package messagehandler

import (
	"errors"
	"net/http"
	"strconv"

	"sn-backend/internal/handler/common"
	"sn-backend/internal/service/filesvc"
	"sn-backend/internal/service/messagesvc"
	"sn-backend/internal/service/sessionsvc"
	ws "sn-backend/internal/websocket"
)

type Handler struct {
	Service   *messagesvc.Service
	Session   *sessionsvc.Service
	WebSocket *ws.Hub
}

func New(service *messagesvc.Service, session *sessionsvc.Service, webSocket *ws.Hub) *Handler {
	return &Handler{Service: service, Session: session, WebSocket: webSocket}
}

// History answers with the stored conversation with one user.
// Messages were always saved; until now nothing could read them back.
func (h *Handler) History(w http.ResponseWriter, r *http.Request) {
	viewerID, ok := h.caller(w, r)
	if !ok {
		return
	}
	otherID, err := common.PathID(r, "id")
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	messages, err := h.Service.History(viewerID, otherID, limit, common.LastID(r))
	if err != nil {
		writeError(w, err, "could not load the conversation")
		return
	}
	common.WriteJSON(w, http.StatusOK, messages)
}

// Send saves a private message and pushes it to both users.
//
// Sending over HTTP (and not over the websocket) is what lets a message carry
// images: the bytes are written first, then the message and its file rows are
// committed together, before anybody is told about the message.
func (h *Handler) Send(w http.ResponseWriter, r *http.Request) {
	fromID, ok := h.caller(w, r)
	if !ok {
		return
	}

	headers, err := common.ReadFormWithFiles(w, r, filesvc.MaxRequestSize, filesvc.MaxMemory)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// the message goes either to one person or to a group chat
	toUserID, err := optionalID(r.FormValue("to_user_id"))
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	groupID, err := optionalID(r.FormValue("group_id"))
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}

	message, err := h.Service.Send(fromID, toUserID, groupID, r.FormValue("content"), headers)
	if err != nil {
		writeError(w, err, "could not send the message")
		return
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

// optionalID reads a form field that holds an id, or nil when it is not there.
func optionalID(value string) (*int64, error) {
	if value == "" {
		return nil, nil
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id < 1 {
		return nil, strconv.ErrSyntax
	}
	return &id, nil
}

func writeError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, messagesvc.ErrNotAllowed):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, messagesvc.ErrEmpty), errors.Is(err, messagesvc.ErrTooLong), filesvc.IsBadImage(err):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, fallback, http.StatusInternalServerError)
	}
}
