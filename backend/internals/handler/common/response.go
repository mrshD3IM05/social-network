package common

import (
	"encoding/json"
	"net/http"
	"sn-backend/internals/model"
	"sn-backend/internals/service/sessionsvc"
)

func PublicUser(user *model.User) map[string]any {
	return map[string]any{"id": user.ID, "first_name": user.FirstName, "last_name": user.LastName, "avatar": user.Avatar, "nickname": user.Nickname, "about_me": user.AboutMe, "private": user.Private, "created_at": user.CreatedAt}
}
func PrivateUser(user *model.User) map[string]any {
	profile := PublicUser(user)
	profile["email"] = user.Email
	profile["date_of_birth"] = user.DateOfBirth
	return profile
}
func WriteJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func CurrentUserID(r *http.Request, sessions *sessionsvc.Service) (int64, error) {
	cookie, err := r.Cookie(sessionsvc.CookieName)
	if err != nil {
		return 0, err
	}
	session, err := sessions.Get(cookie.Value)
	if err != nil {
		return 0, err
	}
	return session.UserID, nil
}
