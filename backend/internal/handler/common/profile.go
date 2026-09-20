package common

import (
	"net/http"
	"strconv"

	"sn-backend/internal/model"
)

// Profile projects a user for the profile page. The subject asks for every
// registration field except the password, so email and date of birth are added
// once the viewer is the owner or an accepted follower; a stranger still only
// sees the public fields.
func Profile(user *model.User, viewerID int64, follower bool) map[string]any {
	profile := PublicUser(user)
	if viewerID == user.ID || follower {
		profile["email"] = user.Email
		profile["date_of_birth"] = user.DateOfBirth
	}
	return profile
}

// PublicUsers projects a list of users, never exposing password or email.
func PublicUsers(users []*model.User) []map[string]any {
	projected := make([]map[string]any, 0, len(users))
	for _, user := range users {
		projected = append(projected, PublicUser(user))
	}
	return projected
}

// PathID reads a positive integer path segment such as {id}.
func PathID(r *http.Request, name string) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id < 1 {
		return 0, strconv.ErrSyntax
	}
	return id, nil
}

// QueryInt reads an optional positive integer query parameter.
func QueryInt(r *http.Request, name string, fallback int) int {
	value, err := strconv.Atoi(r.URL.Query().Get(name))
	if err != nil || value < 1 {
		return fallback
	}
	return value
}

// FormIDs turns repeated form values into a de-duplicated list of ids.
func FormIDs(values []string) []int64 {
	seen := make(map[int64]struct{}, len(values))
	ids := make([]int64, 0, len(values))
	for _, value := range values {
		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil || id < 1 {
			continue
		}
		if _, duplicate := seen[id]; duplicate {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids
}

// FormBool reads a checkbox-style form value.
func FormBool(value string) (bool, bool) {
	switch value {
	case "true", "1", "on", "yes":
		return true, true
	case "false", "0", "off", "no":
		return false, true
	}
	return false, false
}
