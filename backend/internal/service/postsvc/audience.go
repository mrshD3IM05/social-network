package postsvc

import (
	"errors"

	"sn-backend/internal/model"
	"sn-backend/internal/repository"
)

// MaxContentLength mirrors the limit the web client enforces, so the API no
// longer accepts arbitrarily long posts.
const MaxContentLength = 1000

var (
	ErrNotOwner       = errors.New("post: not the author")
	ErrContentTooLong = errors.New("post: content is too long")
)

// AudienceRepository is the slice of the repository needed to manage the
// chosen-follower list of a "private" post.
type AudienceRepository interface {
	GetPost(int64) (*model.Post, error)
	SetPostAudience(int64, int64, []int64) error
	ListPostAudience(int64) ([]int64, error)
}

// AudienceService fills post_visibility, which is what makes a "private" post
// reach the followers its author picked.
type AudienceService struct{ repo AudienceRepository }

func NewAudience(repo AudienceRepository) *AudienceService { return &AudienceService{repo: repo} }

func (s *AudienceService) Set(authorID, postID int64, userIDs []int64) ([]int64, error) {
	post, err := s.repo.GetPost(postID)
	if err != nil {
		return nil, ErrNotFound
	}
	if post.AuthorID != authorID {
		return nil, ErrNotOwner
	}
	if post.Privacy != model.PostSelected {
		return nil, ErrInvalidPrivacy
	}
	if err := s.repo.SetPostAudience(postID, authorID, userIDs); err != nil {
		if errors.Is(err, repository.ErrNotOwner) {
			return nil, ErrNotOwner
		}
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return s.repo.ListPostAudience(postID)
}

func (s *AudienceService) Get(authorID, postID int64) ([]int64, error) {
	post, err := s.repo.GetPost(postID)
	if err != nil {
		return nil, ErrNotFound
	}
	if post.AuthorID != authorID {
		return nil, ErrNotOwner
	}
	return s.repo.ListPostAudience(postID)
}
