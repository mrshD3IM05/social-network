package sessionsvc

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"sn-backend/internal/model"
)

const (
	CookieName = "session"
	DefaultTTL = 30 * 24 * time.Hour
)

var (
	ErrInvalid = errors.New("session: invalid")
	ErrExpired = errors.New("session: expired")
)

type Repository interface {
	CreateSession(*model.Session) error
	GetSession(string) (*model.Session, error)
	DeleteSession(string) error
}

type Service struct {
	repo Repository
	ttl  time.Duration
}

func New(repo Repository) *Service { return &Service{repo: repo, ttl: DefaultTTL} }

func (s *Service) Create(userID int64) (*model.Session, error) {
	var token [32]byte
	if _, err := rand.Read(token[:]); err != nil {
		return nil, err
	}
	session := &model.Session{ID: hex.EncodeToString(token[:]), UserID: userID, ExpiresAt: time.Now().Add(s.ttl)}
	if err := s.repo.CreateSession(session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *Service) Get(token string) (*model.Session, error) {
	if token == "" {
		return nil, ErrInvalid
	}
	session, err := s.repo.GetSession(token)
	if err != nil {
		return nil, ErrInvalid
	}
	if time.Now().After(session.ExpiresAt) {
		_ = s.repo.DeleteSession(token)
		return nil, ErrExpired
	}
	return session, nil
}

func (s *Service) Delete(token string) error { return s.repo.DeleteSession(token) }

func (s *Service) SetCookie(w http.ResponseWriter, session *model.Session) {
	http.SetCookie(w, &http.Cookie{Name: CookieName, Value: session.ID, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Expires: session.ExpiresAt})
}

func (s *Service) ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: CookieName, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: -1})
}
