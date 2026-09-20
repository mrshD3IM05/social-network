package common

import (
	"net/http/httptest"
	"testing"

	"sn-backend/internal/model"
)

func TestPageFallsBackToTheFirstPage(t *testing.T) {
	limit, offset := Page(httptest.NewRequest("GET", "/users", nil))

	if limit != DefaultPageSize || offset != 0 {
		t.Fatalf("want %d/0, got %d/%d", DefaultPageSize, limit, offset)
	}
}

func TestPageReadsTheQueryString(t *testing.T) {
	limit, offset := Page(httptest.NewRequest("GET", "/users?limit=5&offset=10", nil))

	if limit != 5 || offset != 10 {
		t.Fatalf("want 5/10, got %d/%d", limit, offset)
	}
}

func TestPageCapsTheLimit(t *testing.T) {
	limit, _ := Page(httptest.NewRequest("GET", "/users?limit=5000", nil))

	if limit != MaxPageSize {
		t.Fatalf("want the limit capped at %d, got %d", MaxPageSize, limit)
	}
}

func TestFormIDsSkipsJunkAndDuplicates(t *testing.T) {
	ids := FormIDs([]string{"3", "abc", "3", "0", "-1", "7"})

	if len(ids) != 2 || ids[0] != 3 || ids[1] != 7 {
		t.Fatalf("want [3 7], got %v", ids)
	}
}

func TestProfileHidesPrivateFieldsFromStrangers(t *testing.T) {
	user := &model.User{ID: 1, Email: "alice@example.com", DateOfBirth: "1995-01-01"}

	stranger := Profile(user, 2, false)
	if _, found := stranger["email"]; found {
		t.Error("a stranger must not see the email")
	}
	if _, found := stranger["date_of_birth"]; found {
		t.Error("a stranger must not see the date of birth")
	}

	owner := Profile(user, 1, false)
	if owner["email"] != "alice@example.com" {
		t.Error("the owner should see their own email")
	}

	follower := Profile(user, 2, true)
	if follower["date_of_birth"] != "1995-01-01" {
		t.Error("an accepted follower should see the date of birth")
	}
}

func TestProfileNeverLeaksThePassword(t *testing.T) {
	user := &model.User{ID: 1, Password: "hashed-secret"}

	for _, profile := range []map[string]any{Profile(user, 1, true), PublicUser(user), PrivateUser(user)} {
		if _, found := profile["password"]; found {
			t.Fatal("no projection may carry the password")
		}
	}
}
