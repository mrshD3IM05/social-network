package userhandler

import (
	"net/http"
	"sn-backend/internal/handler/common"
	"sn-backend/internal/model"
	"sn-backend/internal/service/followsvc"
	"sn-backend/internal/service/postsvc"
	"sn-backend/internal/service/sessionsvc"
	"sn-backend/internal/service/usersvc"
	"strconv"
)

type Handler struct {
	Service *usersvc.Service
	Session *sessionsvc.Service
	Follow  *followsvc.Service
	Post    *postsvc.Service
}

func New(service *usersvc.Service, session *sessionsvc.Service, follow *followsvc.Service, post *postsvc.Service) *Handler {
	return &Handler{Service: service, Session: session, Follow: follow, Post: post}
}

// ListUsers handles GET /users?q=&last=: one page of the people directory
// every "pick a person" screen reads from (People, Messages, group invites),
// searched by name or nickname. It never includes the caller and only exposes
// the public profile fields.
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	viewerID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	users, err := h.Service.ListUsers(viewerID, r.URL.Query().Get("q"), common.LastID(r))
	if err != nil {
		http.Error(w, "could not list users", http.StatusInternalServerError)
		return
	}
	writePeople(w, users)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	user, ok := h.visibleUser(w, r)
	if !ok {
		return
	}
	viewerID, _ := common.CurrentUserID(r, h.Session)
	posts, followers, following, err := h.Service.ProfileCounts(viewerID, user.ID)
	if err != nil {
		http.Error(w, "could not count profile activity", http.StatusInternalServerError)
		return
	}
	// the subject wants every register field on the profile (never the
	// password), and visibleUser already checked the caller may see it
	profile := common.PrivateUser(user)
	profile["post_count"] = posts
	profile["follower_count"] = followers
	profile["following_count"] = following
	common.WriteJSON(w, http.StatusOK, profile)
}

// UserPosts handles GET /users/{id}/posts?last=: one page of the posts on a
// profile, behind the same privacy gate as the profile itself.
func (h *Handler) UserPosts(w http.ResponseWriter, r *http.Request) {
	user, ok := h.visibleUser(w, r)
	if !ok {
		return
	}
	viewerID, _ := common.CurrentUserID(r, h.Session)
	posts, err := h.Post.UserPosts(viewerID, user.ID, common.LastID(r))
	if err != nil {
		http.Error(w, "could not list posts", http.StatusInternalServerError)
		return
	}
	common.WriteJSON(w, http.StatusOK, posts)
}

// SetPrivacy handles PUT /me/privacy: the switch on your own profile that turns
// it public or private. It always acts on the caller, so one user can never
// change another user's privacy.
func (h *Handler) SetPrivacy(w http.ResponseWriter, r *http.Request) {
	viewerID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	private, err := strconv.ParseBool(r.FormValue("private"))
	if err != nil {
		http.Error(w, "private must be true or false", http.StatusBadRequest)
		return
	}
	user, err := h.Service.SetPrivacy(viewerID, private)
	if err != nil {
		http.Error(w, "could not update profile privacy", http.StatusInternalServerError)
		return
	}
	// a public profile is followed without asking, so the waiting requests
	// are accepted at once
	if !private {
		if err := h.Follow.AcceptAllPending(viewerID); err != nil {
			http.Error(w, "could not accept pending follow requests", http.StatusInternalServerError)
			return
		}
	}
	common.WriteJSON(w, http.StatusOK, common.PrivateUser(user))
}

// Notifications handles GET /notifications: the caller's latest notifications.
func (h *Handler) Notifications(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	notifications, err := h.Service.Notifications(userID, common.LastID(r))
	if err != nil {
		http.Error(w, "could not list notifications", http.StatusInternalServerError)
		return
	}
	common.WriteJSON(w, http.StatusOK, notifications)
}

// visibleUser resolves the {id} in the path and checks the caller may see that
// profile: a private one only opens up to its followers. It writes the error
// itself and answers false once the caller should stop.
func (h *Handler) visibleUser(w http.ResponseWriter, r *http.Request) (*model.User, bool) {
	id, err := common.PathID(r, "id")
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return nil, false
	}
	user, err := h.Service.GetUser(id)
	if err != nil {
		if usersvc.IsNotFound(err) {
			http.Error(w, "user not found", http.StatusNotFound)
		} else {
			http.Error(w, "could not get user", http.StatusInternalServerError)
		}
		return nil, false
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
		return nil, false
	}
	if !visible {
		http.Error(w, "profile is private", http.StatusForbidden)
		return nil, false
	}
	return user, true
}

// writePeople answers with the public profile of every user in the list.
func writePeople(w http.ResponseWriter, users []*model.User) {
	people := make([]map[string]any, 0, len(users))
	for _, user := range users {
		people = append(people, common.PublicUser(user))
	}
	common.WriteJSON(w, http.StatusOK, people)
}
