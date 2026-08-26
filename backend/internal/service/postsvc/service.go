package postsvc

import (
	"errors"
	"sn-backend/internal/model"
	"sn-backend/internal/repository"
)

var (
	ErrInvalidPrivacy = errors.New("post: invalid privacy")
	ErrNotFound       = errors.New("post: not found")
)

type Repository interface {
	CreatePost(*model.Post) error
	GetPost(int64) (*model.Post, error)
	UpdatePostOwned(*model.Post, int64) error
	DeletePostOwned(int64, int64) error
	ListVisiblePosts(int64) ([]*model.Post, error)
}
type Service struct{ repo Repository }

func New(repo Repository) *Service { return &Service{repo: repo} }
func validPrivacy(privacy string) bool {
	return privacy == model.PostPublic || privacy == model.PostFollowersOnly || privacy == model.PostSelected
}
func (s *Service) Create(authorID int64, content, privacy string) (*model.Post, error) {
	if !validPrivacy(privacy) {
		return nil, ErrInvalidPrivacy
	}
	post := &model.Post{AuthorID: authorID, Content: content, Privacy: privacy, Type: "post"}
	if err := s.repo.CreatePost(post); err != nil {
		return nil, err
	}
	return post, nil
}
func (s *Service) Update(ownerID, postID int64, content, privacy string) (*model.Post, error) {
	if !validPrivacy(privacy) {
		return nil, ErrInvalidPrivacy
	}
	post := &model.Post{ID: postID, Content: content, Privacy: privacy}
	if err := s.repo.UpdatePostOwned(post, ownerID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return s.repo.GetPost(postID)
}
func (s *Service) Delete(ownerID, postID int64) error {
	if err := s.repo.DeletePostOwned(postID, ownerID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}
func (s *Service) ListVisible(viewerID int64) ([]*model.Post, error) {
	return s.repo.ListVisiblePosts(viewerID)
}
