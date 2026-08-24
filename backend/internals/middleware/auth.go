package middleware

import (
	"net/http"

	"sn-backend/internals/model"
	"sn-backend/internals/service/sessionsvc"
)

type Auth struct {
	Session *sessionsvc.Service
}

func NewAuth(session *sessionsvc.Service) *Auth {
	return &Auth{Session: session}
}

func (a *Auth) Authorized(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := a.sessionFromRequest(r); err != nil {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *Auth) Guest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := a.sessionFromRequest(r); err == nil {
			http.Error(w, "already authenticated", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *Auth) sessionFromRequest(r *http.Request) (*model.Session, error) {
	cookie, err := r.Cookie(sessionsvc.CookieName)
	if err != nil {
		return nil, err
	}
	return a.Session.Get(cookie.Value)
}
