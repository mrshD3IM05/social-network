package userhandler

import (
	"net/http"
	"sn-backend/internal/handler/common"
	"sn-backend/internal/model"
	"sn-backend/internal/service/followsvc"
	"sn-backend/internal/service/sessionsvc"
	"sn-backend/internal/service/usersvc"
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

// ListUsers handles GET /users: the people directory every "pick a person"
// screen reads from (People, Messages, group invites). It never includes the
// caller and only exposes the public profile fields.
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	viewerID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	users, err := h.Service.ListUsers(viewerID)
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
	common.WriteJSON(w, http.StatusOK, common.PublicUser(user))
}

// Followers handles GET /users/{id}/followers and Following GET
// /users/{id}/following: the two lists a profile shows. They sit behind the same
// privacy gate as the profile itself, so a private one stays hidden.
func (h *Handler) Followers(w http.ResponseWriter, r *http.Request) {
	user, ok := h.visibleUser(w, r)
	if !ok {
		return
	}
	users, err := h.Follow.Followers(user.ID)
	if err != nil {
		http.Error(w, "could not list followers", http.StatusInternalServerError)
		return
	}
	writePeople(w, users)
}
func (h *Handler) Following(w http.ResponseWriter, r *http.Request) {
	user, ok := h.visibleUser(w, r)
	if !ok {
		return
	}
	users, err := h.Follow.Following(user.ID)
	if err != nil {
		http.Error(w, "could not list following", http.StatusInternalServerError)
		return
	}
	writePeople(w, users)
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
	common.WriteJSON(w, http.StatusOK, common.PrivateUser(user))
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

// FollowStatus handles GET /users/{id}/follow: {"status": "accepted" | "pending" | ""}
// so the profile page knows which button to show.
func (h *Handler) FollowStatus(w http.ResponseWriter, r *http.Request) {
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
	status, err := h.Follow.Status(viewerID, targetID)
	if err != nil {
		http.Error(w, "could not get follow status", http.StatusInternalServerError)
		return
	}
	common.WriteJSON(w, http.StatusOK, map[string]string{"status": status})
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

// visibleUser resolves the {id} in the path and checks the caller may see that
// profile: a private one only opens up to its followers. It writes the error
// itself and answers false once the caller should stop.
func (h *Handler) visibleUser(w http.ResponseWriter, r *http.Request) (*model.User, bool) {
	id, err := parseID(r, "id")
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

func parseID(r *http.Request, name string) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id < 1 {
		return 0, strconv.ErrSyntax
	}
	return id, nil
}
