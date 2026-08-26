package followsvc

import (
	"errors"
	"sn-backend/internal/model"
	"sn-backend/internal/repository"
)

var (
	ErrCannotFollowSelf = errors.New("follow: cannot follow yourself")
	ErrExists           = errors.New("follow: relationship already exists")
	ErrNotFound         = errors.New("follow: relationship not found")
	ErrNotRecipient     = errors.New("follow: user is not the recipient")
)

type Repository interface {
	GetUserByID(int64) (*model.User, error)
	GetFollowRequest(int64, int64) (*model.FollowRequest, error)
	GetFollowRequestByID(int64) (*model.FollowRequest, error)
	CreateFollowRequest(int64, int64, string) (*model.FollowRequest, error)
	UpdateFollowStatus(int64, string) error
	DeleteFollow(int64, int64) error
}
type Service struct{ repo Repository }

func New(repo Repository) *Service { return &Service{repo: repo} }
func (s *Service) Follow(from, to int64) (*model.FollowRequest, error) {
	if from == to {
		return nil, ErrCannotFollowSelf
	}
	target, err := s.repo.GetUserByID(to)
	if err != nil {
		return nil, err
	}
	existing, err := s.repo.GetFollowRequest(from, to)
	status := model.FollowAccepted
	if target.Private {
		status = model.FollowPending
	}
	if err == nil {
		if existing.Status == model.FollowAccepted || existing.Status == model.FollowPending {
			return nil, ErrExists
		}
		if err := s.repo.UpdateFollowStatus(existing.ID, status); err != nil {
			return nil, err
		}
		existing.Status = status
		return existing, nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}
	return s.repo.CreateFollowRequest(from, to, status)
}
func (s *Service) Unfollow(from, to int64) error { return s.repo.DeleteFollow(from, to) }
func (s *Service) Respond(recipient, requestID int64, status string) error {
	follow, err := s.repo.GetFollowRequestByID(requestID)
	if err != nil {
		return ErrNotFound
	}
	if follow.ToUserID != recipient {
		return ErrNotRecipient
	}
	if follow.Status != model.FollowPending {
		return ErrExists
	}
	return s.repo.UpdateFollowStatus(requestID, status)
}
