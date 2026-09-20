package socialsvc

import (
	"errors"
	"regexp"
	"strings"

	"sn-backend/internal/model"
	"sn-backend/internal/repository"
)

const (
	MaxNameLength    = 50
	MaxAboutMeLength = 500
)

var (
	ErrNotFound      = errors.New("social: not found")
	ErrForbidden     = errors.New("social: not allowed")
	ErrInvalidInput  = errors.New("social: invalid input")
	ErrNicknameTaken = errors.New("social: nickname already taken")
)

var nicknameRegex = regexp.MustCompile(`^[a-z0-9]{4,15}$`)

type Repository interface {
	GetUserByID(int64) (*model.User, error)
	IsFollowing(int64, int64) (bool, error)
	ListFollowers(int64) ([]*model.User, error)
	ListFollowing(int64) ([]*model.User, error)
	ListAllUsers(int64) ([]*model.User, error)
	ListPendingFollowRequests(int64) ([]*repository.PendingFollowRequest, error)
	FollowState(int64, int64) (string, error)
	ListUserPosts(int64, int64) ([]*model.Post, error)
	SetUserPrivacy(int64, bool) error
	UpdateProfileFields(*model.User) error
	NicknameTaken(string, int64) (bool, error)
}

type Service struct{ repo Repository }

func New(repo Repository) *Service { return &Service{repo: repo} }

// canSee applies the same rule as the profile endpoint: a private profile is
// only readable by its owner and by accepted followers.
func (s *Service) canSee(viewerID, targetID int64) (bool, error) {
	if viewerID == targetID {
		return true, nil
	}
	target, err := s.repo.GetUserByID(targetID)
	if err != nil {
		return false, ErrNotFound
	}
	if !target.Private {
		return true, nil
	}
	return s.repo.IsFollowing(viewerID, targetID)
}

func (s *Service) Followers(viewerID, targetID int64) ([]*model.User, error) {
	if err := s.guard(viewerID, targetID); err != nil {
		return nil, err
	}
	return s.repo.ListFollowers(targetID)
}

func (s *Service) Following(viewerID, targetID int64) ([]*model.User, error) {
	if err := s.guard(viewerID, targetID); err != nil {
		return nil, err
	}
	return s.repo.ListFollowing(targetID)
}

func (s *Service) UserPosts(viewerID, targetID int64) ([]*model.Post, error) {
	if err := s.guard(viewerID, targetID); err != nil {
		return nil, err
	}
	return s.repo.ListUserPosts(targetID, viewerID)
}

func (s *Service) guard(viewerID, targetID int64) error {
	visible, err := s.canSee(viewerID, targetID)
	if err != nil {
		return err
	}
	if !visible {
		return ErrForbidden
	}
	return nil
}

func (s *Service) AllUsers(viewerID int64) ([]*model.User, error) {
	return s.repo.ListAllUsers(viewerID)
}

func (s *Service) PendingRequests(userID int64) ([]*repository.PendingFollowRequest, error) {
	return s.repo.ListPendingFollowRequests(userID)
}

func (s *Service) Relationship(viewerID, targetID int64) (string, error) {
	if viewerID == targetID {
		return "self", nil
	}
	return s.repo.FollowState(viewerID, targetID)
}

func (s *Service) SetPrivacy(userID int64, private bool) (*model.User, error) {
	if err := s.repo.SetUserPrivacy(userID, private); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return s.repo.GetUserByID(userID)
}

// ProfileUpdate carries only the fields a user may change about themselves.
// A nil pointer means "leave this field as it is".
type ProfileUpdate struct {
	FirstName *string
	LastName  *string
	Nickname  *string
	AboutMe   *string
	Private   *bool
}

func (s *Service) UpdateProfile(userID int64, update ProfileUpdate) (*model.User, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, ErrNotFound
	}
	if update.FirstName != nil {
		name := strings.TrimSpace(*update.FirstName)
		if name == "" || len(name) > MaxNameLength {
			return nil, ErrInvalidInput
		}
		user.FirstName = name
	}
	if update.LastName != nil {
		name := strings.TrimSpace(*update.LastName)
		if name == "" || len(name) > MaxNameLength {
			return nil, ErrInvalidInput
		}
		user.LastName = name
	}
	if update.AboutMe != nil {
		about := strings.TrimSpace(*update.AboutMe)
		if len(about) > MaxAboutMeLength {
			return nil, ErrInvalidInput
		}
		user.AboutMe = about
	}
	if update.Nickname != nil {
		nickname := strings.ToLower(strings.TrimSpace(*update.Nickname))
		if nickname != "" && !nicknameRegex.MatchString(nickname) {
			return nil, ErrInvalidInput
		}
		if nickname != user.Nickname {
			taken, err := s.repo.NicknameTaken(nickname, userID)
			if err != nil {
				return nil, err
			}
			if taken {
				return nil, ErrNicknameTaken
			}
		}
		user.Nickname = nickname
	}
	if update.Private != nil {
		user.Private = *update.Private
	}
	if err := s.repo.UpdateProfileFields(user); err != nil {
		return nil, err
	}
	return s.repo.GetUserByID(userID)
}
