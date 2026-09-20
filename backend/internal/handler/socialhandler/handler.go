package socialhandler

import (
	"errors"
	"net/http"

	"sn-backend/internal/handler/common"
	"sn-backend/internal/model"
	"sn-backend/internal/service/sessionsvc"
	"sn-backend/internal/service/socialsvc"
)

type Handler struct {
	Service *socialsvc.Service
	Session *sessionsvc.Service
}

func New(service *socialsvc.Service, session *sessionsvc.Service) *Handler {
	return &Handler{Service: service, Session: session}
}

func (h *Handler) Followers(w http.ResponseWriter, r *http.Request) {
	h.userList(w, r, h.Service.Followers)
}

func (h *Handler) Following(w http.ResponseWriter, r *http.Request) {
	h.userList(w, r, h.Service.Following)
}

func (h *Handler) userList(w http.ResponseWriter, r *http.Request, list func(int64, int64, int, int) ([]*model.User, error)) {
	viewerID, ok := h.caller(w, r)
	if !ok {
		return
	}
	targetID, err := common.PathID(r, "id")
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	limit, offset := common.Page(r)
	users, err := list(viewerID, targetID, limit, offset)
	if err != nil {
		writeError(w, err, "could not list users")
		return
	}
	common.WriteJSON(w, http.StatusOK, common.PublicUsers(users))
}

// UserPosts replaces the client-side "filter the whole feed" workaround with a
// server-side, visibility-aware list of one user's posts.
func (h *Handler) UserPosts(w http.ResponseWriter, r *http.Request) {
	viewerID, ok := h.caller(w, r)
	if !ok {
		return
	}
	targetID, err := common.PathID(r, "id")
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	limit, offset := common.Page(r)
	posts, err := h.Service.UserPosts(viewerID, targetID, limit, offset)
	if err != nil {
		writeError(w, err, "could not list posts")
		return
	}
	common.WriteJSON(w, http.StatusOK, posts)
}

// Users lists every other account, so the directory and the private-post
// audience picker no longer have to guess people from the feed.
func (h *Handler) Users(w http.ResponseWriter, r *http.Request) {
	viewerID, ok := h.caller(w, r)
	if !ok {
		return
	}
	limit, offset := common.Page(r)
	users, err := h.Service.AllUsers(viewerID, limit, offset)
	if err != nil {
		http.Error(w, "could not list users", http.StatusInternalServerError)
		return
	}
	common.WriteJSON(w, http.StatusOK, common.PublicUsers(users))
}

// PendingFollowRequests makes the existing accept/decline endpoints reachable:
// without it no client could ever learn a follow request id.
func (h *Handler) PendingFollowRequests(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.caller(w, r)
	if !ok {
		return
	}
	requests, err := h.Service.PendingRequests(userID)
	if err != nil {
		http.Error(w, "could not list follow requests", http.StatusInternalServerError)
		return
	}
	payload := make([]map[string]any, 0, len(requests))
	for _, request := range requests {
		payload = append(payload, map[string]any{
			"id":         request.ID,
			"status":     request.Status,
			"created_at": request.CreatedAt,
			"from_user":  common.PublicUser(request.From),
		})
	}
	common.WriteJSON(w, http.StatusOK, payload)
}

// Relationship tells the caller how they relate to another profile, so the
// profile page can show the right button instead of both Follow and Unfollow.
func (h *Handler) Relationship(w http.ResponseWriter, r *http.Request) {
	viewerID, ok := h.caller(w, r)
	if !ok {
		return
	}
	targetID, err := common.PathID(r, "id")
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	state, err := h.Service.Relationship(viewerID, targetID)
	if err != nil {
		writeError(w, err, "could not read the relationship")
		return
	}
	common.WriteJSON(w, http.StatusOK, map[string]any{"status": state})
}

// UpdateMe changes the caller's own profile, including the public/private
// toggle the subject asks for.
func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.caller(w, r)
	if !ok {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	update := socialsvc.ProfileUpdate{}
	if r.Form.Has("first_name") {
		value := r.FormValue("first_name")
		update.FirstName = &value
	}
	if r.Form.Has("last_name") {
		value := r.FormValue("last_name")
		update.LastName = &value
	}
	if r.Form.Has("nickname") {
		value := r.FormValue("nickname")
		update.Nickname = &value
	}
	if r.Form.Has("about_me") {
		value := r.FormValue("about_me")
		update.AboutMe = &value
	}
	if r.Form.Has("private") {
		value, ok := common.FormBool(r.FormValue("private"))
		if !ok {
			http.Error(w, "private must be true or false", http.StatusBadRequest)
			return
		}
		update.Private = &value
	}
	user, err := h.Service.UpdateProfile(userID, update)
	if err != nil {
		writeError(w, err, "could not update the profile")
		return
	}
	common.WriteJSON(w, http.StatusOK, common.PrivateUser(user))
}

func (h *Handler) caller(w http.ResponseWriter, r *http.Request) (int64, bool) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return 0, false
	}
	return userID, true
}

func writeError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, socialsvc.ErrNotFound):
		http.Error(w, "not found", http.StatusNotFound)
	case errors.Is(err, socialsvc.ErrForbidden):
		http.Error(w, "profile is private", http.StatusForbidden)
	case errors.Is(err, socialsvc.ErrNicknameTaken):
		http.Error(w, "nickname already taken", http.StatusConflict)
	case errors.Is(err, socialsvc.ErrInvalidInput):
		http.Error(w, "invalid profile details", http.StatusBadRequest)
	default:
		http.Error(w, fallback, http.StatusInternalServerError)
	}
}
