package userhandler

import (
	"net/http"
	"sn-backend/internals/handler/common"
	"sn-backend/internals/model"
	"sn-backend/internals/service/followsvc"
	"sn-backend/internals/service/sessionsvc"
	"sn-backend/internals/service/usersvc"
	"strconv"
	"strings"
)

type Handler struct {
	Service *usersvc.Service
	Session *sessionsvc.Service
	Follow  *followsvc.Service
}

func New(service *usersvc.Service, session *sessionsvc.Service, follow *followsvc.Service) *Handler {
	return &Handler{Service: service, Session: session, Follow: follow}
}
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	user, err := h.Service.GetUser(id)
	if err != nil {
		if usersvc.IsNotFound(err) {
			http.Error(w, "user not found", http.StatusNotFound)
		} else {
			http.Error(w, "could not get user", http.StatusInternalServerError)
		}
		return
	}
	viewerID := int64(0)
	if cookie, cookieErr := r.Cookie(sessionsvc.CookieName); cookieErr == nil {
		if session, sessionErr := h.Session.Get(cookie.Value); sessionErr == nil {
			viewerID = session.UserID
		}
	}
	visible, err := h.Service.CanViewProfile(viewerID, user)
	if err != nil {
		http.Error(w, "could not check profile access", http.StatusInternalServerError)
		return
	}
	if !visible {
		http.Error(w, "profile is private", http.StatusForbidden)
		return
	}
	common.WriteJSON(w, http.StatusOK, common.PublicUser(user))
}
func (h *Handler) FollowUser(w http.ResponseWriter, r *http.Request) {
	viewerID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	targetID, err := parseID(r, "id")
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	follow, err := h.Follow.Follow(viewerID, targetID)
	if err != nil {
		if err == followsvc.ErrCannotFollowSelf || err == followsvc.ErrExists {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, "could not follow user", http.StatusInternalServerError)
		}
		return
	}
	common.WriteJSON(w, http.StatusCreated, follow)
}
func (h *Handler) UnfollowUser(w http.ResponseWriter, r *http.Request) {
	viewerID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	targetID, err := parseID(r, "id")
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	if err := h.Follow.Unfollow(viewerID, targetID); err != nil {
		http.Error(w, "could not unfollow user", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h *Handler) RespondFollow(w http.ResponseWriter, r *http.Request) {
	viewerID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	requestID, err := parseID(r, "id")
	if err != nil {
		http.Error(w, "invalid follow request id", http.StatusBadRequest)
		return
	}
	status := model.FollowAccepted
	if strings.HasSuffix(r.URL.Path, "/decline") {
		status = model.FollowDeclined
	}
	if err := h.Follow.Respond(viewerID, requestID, status); err != nil {
		if err == followsvc.ErrNotRecipient {
			http.Error(w, "not the follow request recipient", http.StatusForbidden)
		} else {
			http.Error(w, "could not respond to follow request", http.StatusConflict)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "user listing is not implemented", http.StatusNotImplemented)
}
func parseID(r *http.Request, name string) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id < 1 {
		return 0, strconv.ErrSyntax
	}
	return id, nil
}
