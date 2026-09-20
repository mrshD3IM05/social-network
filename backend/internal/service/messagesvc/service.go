package messagesvc

import (
	"errors"
	"time"

	"sn-backend/internal/model"
	"sn-backend/internal/repository"
)

var ErrNotAllowed = errors.New("message: conversation is not permitted")

type Repository interface {
	CanMessage(int64, *int64, *int64) (bool, error)
	ListConversation(int64, int64, int, *time.Time) ([]*model.Message, error)
	ListConversations(int64) ([]*repository.Conversation, error)
}

type Service struct{ repo Repository }

func New(repo Repository) *Service { return &Service{repo: repo} }

// Conversation returns the stored history with one user, gated by the same
// follow-or-public rule the websocket applies before accepting a message.
func (s *Service) Conversation(viewerID, otherID int64, limit int, before *time.Time) ([]*model.Message, error) {
	allowed, err := s.repo.CanMessage(viewerID, &otherID, nil)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrNotAllowed
	}
	return s.repo.ListConversation(viewerID, otherID, limit, before)
}

func (s *Service) Inbox(viewerID int64) ([]*repository.Conversation, error) {
	return s.repo.ListConversations(viewerID)
}
