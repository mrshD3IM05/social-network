package authhandler

import (
	"errors"
	"net/http"
	"sn-backend/internal/handler/common"
	"sn-backend/internal/service/authsvc"
	"sn-backend/internal/service/sessionsvc"
	ws "sn-backend/internal/websocket"
)

type Handler struct {
	Service   *authsvc.Service
	Session   *sessionsvc.Service
	WebSocket *ws.Hub
}

func New(service *authsvc.Service, session *sessionsvc.Service, webSocket *ws.Hub) *Handler {
	return &Handler{Service: service, Session: session, WebSocket: webSocket}
}
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	user, err := h.Service.Register(authsvc.RegisterInput{Email: r.FormValue("email"), Password: r.FormValue("password"), FirstName: r.FormValue("first_name"), LastName: r.FormValue("last_name"), DateOfBirth: r.FormValue("date_of_birth"), Avatar: r.FormValue("avatar"), Nickname: r.FormValue("nickname"), AboutMe: r.FormValue("about_me")})
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, authsvc.ErrEmailTaken) {
			status = http.StatusConflict
		}
		http.Error(w, err.Error(), status)
		return
	}
	session, err := h.Session.Create(user.ID)
	if err != nil {
		http.Error(w, "could not create session", http.StatusInternalServerError)
		return
	}
	h.Session.SetCookie(w, session)
	common.WriteJSON(w, http.StatusCreated, common.PrivateUser(user))
}
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	user, err := h.Service.Login(r.FormValue("email"), r.FormValue("password"))
	if err != nil {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}
	session, err := h.Session.Create(user.ID)
	if err != nil {
		http.Error(w, "could not create session", http.StatusInternalServerError)
		return
	}
	h.Session.SetCookie(w, session)
	common.WriteJSON(w, http.StatusOK, common.PrivateUser(user))
}
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionsvc.CookieName); err == nil {
		_ = h.Session.Delete(cookie.Value)
		if h.WebSocket != nil {
			h.WebSocket.RevokeSessionClients(cookie.Value)
		}
	}
	h.Session.ClearCookie(w)
	common.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, err := common.CurrentUserID(r, h.Session)
	if err != nil {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	user, err := h.Service.UserByID(userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	common.WriteJSON(w, http.StatusOK, common.PrivateUser(user))
}
