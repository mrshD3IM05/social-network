package usersvc

import (
	"errors"
	"sn-backend/internal/model"
	"sn-backend/internal/repository"
	"strings"
)

type Service struct{ users *repository.Repository }

func New(users *repository.Repository) *Service          { return &Service{users: users} }
func (s *Service) GetUser(id int64) (*model.User, error) { return s.users.GetUserByID(id) }

// ListUsers is the directory behind GET /users: every registered user except
// the viewer. Private profiles stay in the list — CanViewProfile still gates
// the profile itself.
func (s *Service) ListUsers(viewerID int64, search string, offset int) ([]*model.User, error) {
	return s.users.ListUsers(viewerID, strings.TrimSpace(search), offset)
}

// ProfileCounts are the numbers shown on a profile: the posts viewerID can
// see, the followers and the followed users.
func (s *Service) ProfileCounts(viewerID, userID int64) (posts, followers, following int, err error) {
	if posts, err = s.users.CountVisiblePosts(viewerID, userID); err != nil {
		return
	}
	if followers, err = s.users.CountFollowers(userID); err != nil {
		return
	}
	following, err = s.users.CountFollowing(userID)
	return
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

// Notifications is one page of the user's notifications, newest first.
func (s *Service) Notifications(userID int64, offset int) ([]*model.Notification, error) {
	return s.users.ListNotifications(userID, offset)
}

func (s *Service) UnreadNotifications(userID int64) (int, error) {
	return s.users.CountUnreadNotifications(userID)
}

func (s *Service) MarkNotificationsRead(userID int64) error {
	return s.users.MarkNotificationsRead(userID)
}

func IsNotFound(err error) bool { return errors.Is(err, repository.ErrNotFound) }
