package messagesvc

import (
	"errors"
	"strings"

	"sn-backend/internal/model"
)

const (
	MaxContentLength = 1000
	DefaultLimit     = 50
	MaxLimit         = 200
)

var (
	ErrNotAllowed = errors.New("message: you cannot write here")
	ErrEmpty      = errors.New("message: write something or add an image")
	ErrTooLong    = errors.New("message: content is too long")
)

type Repository interface {
	CanMessage(int64, *int64, *int64) (bool, error)
	CreateMessage(*model.Message) error
	ListMessages(int64, int64, int) ([]*model.Message, error)
	ListMessageFileIDs(int64) ([]string, error)
}

type Service struct{ repo Repository }

func New(repo Repository) *Service { return &Service{repo: repo} }

// History returns the stored conversation with one user. The same rule the
// websocket applies before accepting a message guards it, so history cannot be
// read by someone who could not have taken part in it.
func (s *Service) History(viewerID, otherID int64, limit int) ([]*model.Message, error) {
	allowed, err := s.repo.CanMessage(viewerID, &otherID, nil)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrNotAllowed
	}
	if limit < 1 || limit > MaxLimit {
		limit = DefaultLimit
	}
	return s.repo.ListMessages(viewerID, otherID, limit)
}

// Send saves a message, either to one person or to a group chat.
// withImages says whether pictures will be attached afterwards, which is what
// allows a message with no text.
func (s *Service) Send(fromID int64, toUserID, groupID *int64, content string, withImages bool) (*model.Message, error) {
	allowed, err := s.repo.CanMessage(fromID, toUserID, groupID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrNotAllowed
	}

	content = strings.TrimSpace(content)
	if content == "" && !withImages {
		return nil, ErrEmpty
	}
	if len(content) > MaxContentLength {
		return nil, ErrTooLong
	}

	message := &model.Message{
		FromUserID: fromID,
		ToUserID:   toUserID,
		GroupID:    groupID,
		Content:    content,
		Images:     []string{},
	}
	if err := s.repo.CreateMessage(message); err != nil {
		return nil, err
	}
	return message, nil
}

// LoadImages fills in the pictures of a message once they are uploaded.
func (s *Service) LoadImages(message *model.Message) error {
	images, err := s.repo.ListMessageFileIDs(message.ID)
	if err != nil {
		return err
	}
	message.Images = images
	return nil
}
