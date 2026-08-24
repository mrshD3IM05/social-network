package authsvc

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"sn-backend/internals/model"
	"sn-backend/internals/repository"

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
	GetUserByID(int64) (*model.User, error)
}

type Service struct{ users Repository }

func New(users Repository) *Service { return &Service{users: users} }

func (s *Service) UserByID(id int64) (*model.User, error) { return s.users.GetUserByID(id) }

func (s *Service) Login(email, password string) (*model.User, error) {
	user, err := s.users.GetUserByEmail(email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
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
	address, err := mail.ParseAddress(input.Email)
	if err != nil || strings.TrimSpace(address.Address) != strings.TrimSpace(input.Email) {
		return fmt.Errorf("%w: invalid email address", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Password) == "" {
		return fmt.Errorf("%w: password is required", ErrInvalidInput)
	}
	if strings.TrimSpace(input.FirstName) == "" || strings.TrimSpace(input.LastName) == "" {
		return fmt.Errorf("%w: first and last name are required", ErrInvalidInput)
	}
	if _, err := time.Parse("2006-01-02", input.DateOfBirth); err != nil {
		return fmt.Errorf("%w: date of birth must be YYYY-MM-DD", ErrInvalidInput)
	}
	return nil
}
