package usersvc

import (
	"errors"
	"sn-backend/internal/model"
	"sn-backend/internal/repository"
)

type Service struct{ users *repository.Repository }

func New(users *repository.Repository) *Service          { return &Service{users: users} }
func (s *Service) GetUser(id int64) (*model.User, error) { return s.users.GetUserByID(id) }

// ListUsers is the directory behind GET /users: every registered user except
// the viewer. Private profiles stay in the list — CanViewProfile still gates
// the profile itself.
func (s *Service) ListUsers(viewerID int64) ([]*model.User, error) {
	return s.users.ListUsers(viewerID)
}
func (s *Service) CanViewProfile(viewerID int64, user *model.User) (bool, error) {
	if !user.Private || viewerID == user.ID {
		return true, nil
	}
	if viewerID == 0 {
		return false, nil
	}
	return s.users.IsFollowing(viewerID, user.ID)
}

// SetPrivacy turns the caller's own profile public or private and answers with
// the stored user, so the client never has to guess what was saved.
func (s *Service) SetPrivacy(userID int64, private bool) (*model.User, error) {
	if err := s.users.SetUserPrivate(userID, private); err != nil {
		return nil, err
	}
	return s.users.GetUserByID(userID)
}

// Notifications are the latest notifications of the user, newest first.
func (s *Service) Notifications(userID int64) ([]*model.Notification, error) {
	return s.users.ListNotifications(userID)
}

func (s *Service) MarkNotificationsRead(userID int64) error {
	return s.users.MarkNotificationsRead(userID)
}

func IsNotFound(err error) bool { return errors.Is(err, repository.ErrNotFound) }
