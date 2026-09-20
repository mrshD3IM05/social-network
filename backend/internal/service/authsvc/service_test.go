package authsvc

import (
	"testing"
	"time"
)

// validInput is a registration that should always be accepted.
func validInput() RegisterInput {
	return RegisterInput{
		Email:       "alice@example.com",
		Password:    "password123",
		FirstName:   "Alice",
		LastName:    "Smith",
		DateOfBirth: "1995-01-01",
	}
}

func TestNicknameIsOptional(t *testing.T) {
	if err := validateRegisterInput(validInput()); err != nil {
		t.Fatalf("an empty nickname should be accepted, got %v", err)
	}
}

func TestNicknameIsCheckedWhenGiven(t *testing.T) {
	input := validInput()
	input.Nickname = "no"

	if err := validateRegisterInput(input); err == nil {
		t.Fatal("a nickname shorter than four characters should be rejected")
	}
}

func TestEmailMustLookLikeAnEmail(t *testing.T) {
	input := validInput()
	input.Email = "not-an-email"

	if err := validateRegisterInput(input); err == nil {
		t.Fatal("an address without a domain should be rejected")
	}
}

func TestNameIsRequired(t *testing.T) {
	input := validInput()
	input.FirstName = "   "

	if err := validateRegisterInput(input); err == nil {
		t.Fatal("a blank first name should be rejected")
	}
}

func TestUserMustBeThirteen(t *testing.T) {
	input := validInput()
	input.DateOfBirth = time.Now().AddDate(-12, 0, 0).Format("2006-01-02")

	if err := validateRegisterInput(input); err == nil {
		t.Fatal("a twelve year old should be rejected")
	}
}

func TestDateOfBirthCannotBeInTheFuture(t *testing.T) {
	input := validInput()
	input.DateOfBirth = time.Now().AddDate(1, 0, 0).Format("2006-01-02")

	if err := validateRegisterInput(input); err == nil {
		t.Fatal("a date in the future should be rejected")
	}
}
