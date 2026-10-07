package messagesvc

import (
	"errors"
	"mime/multipart"
	"strings"
	"unicode/utf8"

	"sn-backend/internal/model"
	"sn-backend/internal/repository"
	"sn-backend/internal/service/filesvc"
)

const (
	MaxContentLength = 1000
	DefaultLimit     = repository.MessagePageSize
	MaxLimit         = 200
)

var (
	ErrNotAllowed = errors.New("message: you cannot write here")
	ErrEmpty      = errors.New("message: write something or add an image")
	ErrTooLong    = errors.New("message: content is too long")
)

type Service struct {
	repos *repository.Repositories
	files *filesvc.Service
}

func New(repos *repository.Repositories, files *filesvc.Service) *Service {
	return &Service{repos: repos, files: files}
}

// History returns the stored conversation with one user. The same rule the
// websocket applies before accepting a message guards it, so history cannot be
// read by someone who could not have taken part in it.
func (s *Service) History(viewerID, otherID int64, limit int, lastID int64) ([]*model.Message, error) {
	allowed, err := s.repos.Messages.CanMessage(viewerID, &otherID, nil)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrNotAllowed
	}
	if limit < 1 || limit > MaxLimit {
		limit = DefaultLimit
	}
	return s.repos.Messages.ListMessages(viewerID, otherID, limit, lastID)
}

// Send saves a message, either to one person or a group chat. Any images are
// written to disk first, then the message and its file rows are committed in
// one transaction, so a failure leaves nothing half-attached behind.
func (s *Service) Send(fromID int64, toUserID, groupID *int64, content string, headers []*multipart.FileHeader) (*model.Message, error) {
	allowed, err := s.repos.Messages.CanMessage(fromID, toUserID, groupID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrNotAllowed
	}

	content = strings.TrimSpace(content)
	if content == "" && len(headers) == 0 {
		return nil, ErrEmpty
	}
	if utf8.RuneCountInString(content) > MaxContentLength {
		return nil, ErrTooLong
	}

	message := &model.Message{
		FromUserID: fromID,
		ToUserID:   toUserID,
		GroupID:    groupID,
		Content:    content,
		Images:     []string{},
	}
	if len(headers) == 0 {
		if err := s.repos.Messages.CreateMessage(message); err != nil {
			return nil, err
		}
		return message, nil
	}

	staged, err := s.files.Stage(fromID, headers)
	if err != nil {
		return nil, err
	}
	err = s.repos.WithinTx(func(tx *repository.Repositories) error {
		if err := tx.Messages.CreateMessage(message); err != nil {
			return err
		}
		for _, file := range staged {
			file.MessageID = &message.ID
			if err := tx.Files.CreateFile(file); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		s.files.Discard(staged)
		return nil, err
	}
	for _, file := range staged {
		message.Images = append(message.Images, file.ID)
	}
	return message, nil
}
