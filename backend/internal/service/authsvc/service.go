package authsvc

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"sn-backend/internal/model"
	"sn-backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailTaken         = errors.New("auth: email already registered")
	ErrNicknameTaken      = errors.New("auth: nickname already taken")
	ErrInvalidCredentials = errors.New("auth: invalid credentials")
	ErrInvalidInput       = errors.New("auth: invalid registration input")
)

// Same limits as the frontend (frontend/lib/validate.js).
const (
	maxNameLen    = 50
	maxEmailLen   = 254
	maxAboutMeLen = 500
	minPassword   = 8
	maxPassword   = 72 // bcrypt only reads the first 72 bytes
)

// dummyHash is compared against when the account does not exist, so a failed
// login takes as long whether or not the email is registered.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy password"), bcrypt.DefaultCost)

type Service struct{ users *repository.Repository }

func New(users *repository.Repository) *Service { return &Service{users: users} }

func (s *Service) UserByID(id int64) (*model.User, error) { return s.users.GetUserByID(id) }

func (s *Service) Login(identifier, password string) (*model.User, error) {
	var user *model.User
	var err error

	identifier = strings.ToLower(strings.TrimSpace(identifier))
	if isValidEmail(identifier) {
		user, err = s.users.GetUserByEmail(identifier)
	} else if isValidNickname(identifier) {
		user, err = s.users.GetUserByNickname(identifier)
	} else {
		return nil, ErrInvalidCredentials
	}

	if err != nil || user == nil {
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		return nil, ErrInvalidCredentials
	}

	if bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	) != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

type RegisterInput struct {
	Email, Password, FirstName, LastName, DateOfBirth string
	Nickname, AboutMe                                 string
}

func (s *Service) Register(input RegisterInput) (*model.User, error) {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.Nickname = strings.ToLower(strings.TrimSpace(input.Nickname))
	input.FirstName = strings.TrimSpace(input.FirstName)
	input.LastName = strings.TrimSpace(input.LastName)
	input.AboutMe = strings.TrimSpace(input.AboutMe)
	if err := validateRegisterInput(input); err != nil {
		return nil, err
	}
	if _, err := s.users.GetUserByEmail(input.Email); err == nil {
		return nil, ErrEmailTaken
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}
	// login accepts a nickname too, so two accounts cannot share one
	if _, err := s.users.GetUserByNickname(input.Nickname); err == nil {
		return nil, ErrNicknameTaken
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &model.User{Email: input.Email, Password: string(hash), FirstName: input.FirstName, LastName: input.LastName, DateOfBirth: input.DateOfBirth, Nickname: input.Nickname, AboutMe: input.AboutMe}
	if err := s.users.CreateUser(user); err != nil {
		return nil, err
	}
	return user, nil
}

func validateRegisterInput(input RegisterInput) error {
	if strings.TrimSpace(input.Email) == "" {
		return fmt.Errorf("%w: email is required", ErrInvalidInput)
	}
	if len(input.Email) > maxEmailLen || !isValidEmail(input.Email) {
		return fmt.Errorf("%w: invalid email address", ErrInvalidInput)
	}

	if strings.TrimSpace(input.Nickname) == "" {
		return fmt.Errorf("%w: nickname is required", ErrInvalidInput)
	}
	if !isValidNickname(input.Nickname) {
		return fmt.Errorf("%w: invalid nickname", ErrInvalidInput)
	}

	if strings.TrimSpace(input.Password) == "" {
		return fmt.Errorf("%w: password is required", ErrInvalidInput)
	}
	if len(input.Password) < minPassword || len(input.Password) > maxPassword {
		return fmt.Errorf("%w: password must be %d to %d characters", ErrInvalidInput, minPassword, maxPassword)
	}

	if input.FirstName == "" || input.LastName == "" {
		return fmt.Errorf("%w: first and last name are required", ErrInvalidInput)
	}
	if utf8.RuneCountInString(input.FirstName) > maxNameLen || utf8.RuneCountInString(input.LastName) > maxNameLen {
		return fmt.Errorf("%w: names must be %d characters or less", ErrInvalidInput, maxNameLen)
	}
	if utf8.RuneCountInString(input.AboutMe) > maxAboutMeLen {
		return fmt.Errorf("%w: about me must be %d characters or less", ErrInvalidInput, maxAboutMeLen)
	}

	birthDate, err := time.Parse("2006-01-02", input.DateOfBirth)
	if err != nil {
		return fmt.Errorf("%w: date of birth must be YYYY-MM-DD", ErrInvalidInput)
	}

	today := time.Now()

	if birthDate.After(today) {
		return fmt.Errorf("%w: date of birth cannot be in the future", ErrInvalidInput)
	}

	age := today.Year() - birthDate.Year()

	if today.Month() < birthDate.Month() ||
		(today.Month() == birthDate.Month() && today.Day() < birthDate.Day()) {
		age--
	}

	if age < 13 {
		return fmt.Errorf("%w: must be at least 13 years old", ErrInvalidInput)
	}

	return nil
}

var (
	emailRegex    = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	nicknameRegex = regexp.MustCompile(`^[a-z0-9]{4,15}$`)
	letterRegex   = regexp.MustCompile(`[a-z]`)
)

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func isValidNickname(nickname string) bool {
	return nicknameRegex.MatchString(nickname) &&
		letterRegex.MatchString(nickname)
}
