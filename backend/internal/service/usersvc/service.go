package usersvc

import (
	"errors"
	"sn-backend/internal/model"
	"sn-backend/internal/repository"
)

type Repository interface {
	GetUserByID(int64) (*model.User, error)
	ListUsers(int64) ([]*model.User, error)
	IsFollowing(int64, int64) (bool, error)
}

type Service struct{ users Repository }

func New(users Repository) *Service                      { return &Service{users: users} }
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
func IsNotFound(err error) bool { return errors.Is(err, repository.ErrNotFound) }
