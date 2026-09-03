package postsvc

import (
	"errors"
	"sn-backend/internal/model"
	"sn-backend/internal/repository"
)

var (
	ErrInvalidPrivacy  = errors.New("post: invalid privacy")
	ErrInvalidReaction = errors.New("post: invalid reaction")
	ErrNotFound        = errors.New("post: not found")
)

type Repository interface {
	CreatePost(*model.Post) error
	GetPost(int64) (*model.Post, error)
	UpdatePostOwned(*model.Post, int64) error
	DeletePostOwned(int64, int64) error
	ListVisiblePosts(int64) ([]*model.Post, error)
	CanViewPost(int64, int64) (bool, error)
	LoadPostReactions(*model.Post, int64) error
	GetReaction(string, int64, int64) (*model.Reaction, error)
	SetReaction(string, int64, int64, string) error
	DeleteReaction(string, int64, int64) error
	GetReactionSummary(string, int64, int64) (*model.ReactionSummary, error)
}
type Service struct{ repo Repository }

func New(repo Repository) *Service { return &Service{repo: repo} }
func validPrivacy(privacy string) bool {
	return privacy == model.PostPublic || privacy == model.PostFollowersOnly || privacy == model.PostSelected
}
func validReaction(reaction string) bool {
	return reaction == model.ReactionLike || reaction == model.ReactionDislike
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
	updated, err := s.repo.GetPost(postID)
	if err != nil {
		return nil, err
	}
	if err := s.repo.LoadPostReactions(updated, ownerID); err != nil {
		return nil, err
	}
	return updated, nil
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

// React sets the viewer's like or dislike on a post and returns the updated
// totals. Sending the reaction the viewer already has removes it, so the same
// endpoint toggles a button off; sending the other one switches sides.
func (s *Service) React(viewerID, postID int64, reaction string) (*model.ReactionSummary, error) {
	if !validReaction(reaction) {
		return nil, ErrInvalidReaction
	}
	visible, err := s.repo.CanViewPost(viewerID, postID)
	if err != nil {
		return nil, err
	}
	if !visible {
		return nil, ErrNotFound
	}
	existing, err := s.repo.GetReaction(model.ReactionTargetPost, postID, viewerID)
	switch {
	case err == nil && existing.Reaction == reaction:
		err = s.repo.DeleteReaction(model.ReactionTargetPost, postID, viewerID)
	case err == nil, errors.Is(err, repository.ErrNotFound):
		err = s.repo.SetReaction(model.ReactionTargetPost, postID, viewerID, reaction)
	}
	if err != nil {
		return nil, err
	}
	return s.repo.GetReactionSummary(model.ReactionTargetPost, postID, viewerID)
}

// Unreact removes the viewer's reaction from a post, if they had one.
func (s *Service) Unreact(viewerID, postID int64) (*model.ReactionSummary, error) {
	visible, err := s.repo.CanViewPost(viewerID, postID)
	if err != nil {
		return nil, err
	}
	if !visible {
		return nil, ErrNotFound
	}
	if err := s.repo.DeleteReaction(model.ReactionTargetPost, postID, viewerID); err != nil {
		return nil, err
	}
	return s.repo.GetReactionSummary(model.ReactionTargetPost, postID, viewerID)
}
