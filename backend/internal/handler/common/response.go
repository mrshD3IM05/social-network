package common

import (
	"encoding/json"
	"net/http"
	"sn-backend/internal/model"
	"sn-backend/internal/service/sessionsvc"
	"strconv"
)

// PublicUser is the profile as anyone may see it: never the email, the date
// of birth or the password, plus the follow counts and the relation the viewer
// has with the user it was read for.
func PublicUser(user *model.User) map[string]any {
	return map[string]any{
		"id":           user.ID,
		"first_name":   user.FirstName,
		"last_name":    user.LastName,
		"avatar":       user.Avatar,
		"nickname":     user.Nickname,
		"about_me":     user.AboutMe,
		"private":      user.Private,
		"created_at":   user.CreatedAt,
		"is_followed":  user.IsFollowed,
		"is_following": user.IsFollowing,
		"followers":    user.Followers,
		"following":    user.Following,
		"post_count":   user.PostCount,
	}
}

// PrivateUser is the profile with the contact details on it, for the endpoints
// where the reader is the subject: /me, /register, /login, /me/privacy, /avatar.
func PrivateUser(user *model.User) map[string]any {
	profile := PublicUser(user)
	profile["email"] = user.Email
	profile["date_of_birth"] = user.DateOfBirth

	return profile
}

// Profile is the profile of user as viewerID may see it. A private profile only
// shows its contact details to the subject and to the users following them, so
// a stranger gets the public profile alone — never a 403. A follow still
// waiting for an answer is not enough to see them.
//
// The post count rides along on the user, so it is the same for everyone and
// counts every privacy level: the posts tab behind it may show fewer.
func Profile(user *model.User, viewerID int64) map[string]any {
	var profile map[string]any
	if user.Private && viewerID != user.ID && user.IsFollowing != model.FollowStateActive {
		profile = PublicUser(user)
	} else {
		profile = PrivateUser(user)
	}
	if viewerID == user.ID {
		// remove is followed and is_following from the profile of the subject, so it never sees itself as following or followed by anyone
		delete(profile, "is_followed")
		delete(profile, "is_following")
	}
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

// LastID reads ?last=, the id of the last item the client already has: the
// next page starts right after it. 0 (or nothing) asks for the first page.
func LastID(r *http.Request) int64 {
	last, err := strconv.ParseInt(r.URL.Query().Get("last"), 10, 64)
	if err != nil || last < 0 {
		return 0
	}
	return last
}

// PathID reads a positive id from the url path, like {id} in /posts/{id}.
func PathID(r *http.Request, name string) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id < 1 {
		return 0, strconv.ErrSyntax
	}
	return id, nil
}
