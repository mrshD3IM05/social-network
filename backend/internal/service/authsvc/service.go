package authsvc

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"sn-backend/internal/model"
	"sn-backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailTaken         = errors.New("auth: email already registered")
	ErrInvalidCredentials = errors.New("auth: invalid credentials")
	ErrInvalidInput       = errors.New("auth: invalid registration input")
)

type Repository interface {
	CreateUser(*model.User) error
	GetUserByEmail(string) (*model.User, error)
	GetUserByNickname(string) (*model.User, error)
	GetUserByID(int64) (*model.User, error)
}

type Service struct{ 
	users Repository
 }

func New(users Repository) *Service { return &Service{users: users} }

func (s *Service) UserByID(id int64) (*model.User, error) { return s.users.GetUserByID(id) }

func (s *Service) Login(identifier, password string) (*model.User, error) {
	var user *model.User
	var err error

	if isValidEmail(identifier) {
		user, err = s.users.GetUserByEmail(identifier)
	} else if isValidNickname(identifier) {
		user, err = s.users.GetUserByNickname(identifier)
	} else {
		return nil, ErrInvalidCredentials
	}

	if err != nil || user == nil {
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
	Avatar, Nickname, AboutMe                         string
}

func (s *Service) Register(input RegisterInput) (*model.User, error) {
	if err := validateRegisterInput(input); err != nil {
		return nil, err
	}
	if _, err := s.users.GetUserByEmail(input.Email); err == nil {
		return nil, ErrEmailTaken
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &model.User{Email: input.Email, Password: string(hash), FirstName: input.FirstName, LastName: input.LastName, DateOfBirth: input.DateOfBirth, Avatar: input.Avatar, Nickname: input.Nickname, AboutMe: input.AboutMe}
	if err := s.users.CreateUser(user); err != nil {
		return nil, err
	}
	return user, nil
}

func validateRegisterInput(input RegisterInput) error {
	if strings.TrimSpace(input.Email) == "" {
		return fmt.Errorf("%w: email is required", ErrInvalidInput)
	}
	if !isValidEmail(input.Email) {
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

	if strings.TrimSpace(input.FirstName) == "" ||
		strings.TrimSpace(input.LastName) == "" {
		return fmt.Errorf("%w: first and last name are required", ErrInvalidInput)
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
