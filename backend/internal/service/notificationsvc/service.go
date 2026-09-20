package notificationsvc

import (
	"errors"

	"sn-backend/internal/model"
	"sn-backend/internal/repository"
)

var ErrNotFound = errors.New("notification: not found")

type Repository interface {
	ListNotifications(int64, int) ([]*model.Notification, error)
	CountUnreadNotifications(int64) (int, error)
	MarkNotificationRead(int64, int64) error
	MarkAllNotificationsRead(int64) error
}

type Service struct{ repo Repository }

func New(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(userID int64, limit int) ([]*model.Notification, error) {
	return s.repo.ListNotifications(userID, limit)
}

func (s *Service) UnreadCount(userID int64) (int, error) {
	return s.repo.CountUnreadNotifications(userID)
}

func (s *Service) MarkRead(userID, notificationID int64) error {
	if err := s.repo.MarkNotificationRead(notificationID, userID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *Service) MarkAllRead(userID int64) error { return s.repo.MarkAllNotificationsRead(userID) }
