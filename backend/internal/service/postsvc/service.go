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
	ErrNotGroupMember  = errors.New("post: only group members can do that")
	ErrForbidden       = errors.New("post: only the author or the group creator can delete a group post")
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
	IsGroupMember(int64, int64) (bool, error)
	ListGroupPosts(groupID, viewerID int64) ([]*model.Post, error)
	GetGroupIDForPost(int64) (int64, error)
	GetPostAuthor(int64) (int64, error)
	IsGroupCreator(int64, int64) (bool, error)
	DeletePost(int64) error
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
	post := &model.Post{AuthorID: authorID, Content: content, Privacy: privacy}
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

// Get returns one post the viewer is allowed to see (privacy rules for
// normal posts, group membership for group posts — both in CanViewPost).
func (s *Service) Get(viewerID, postID int64) (*model.Post, error) {
	visible, err := s.repo.CanViewPost(viewerID, postID)
	if err != nil {
		return nil, err
	}
	if !visible {
		return nil, ErrNotFound
	}
	post, err := s.repo.GetPost(postID)
	if err != nil {
		return nil, err
	}
	if err := s.repo.LoadPostReactions(post, viewerID); err != nil {
		return nil, err
	}
	return post, nil
}

// CreateGroupPost creates a post inside a group. Membership is verified
// server-side; the visibility of group posts is membership, so the privacy
// column is pinned to public (never used by the group branch of the
// visibility rules) and client-supplied privacy values are ignored.
func (s *Service) CreateGroupPost(authorID, groupID int64, content, privacy string) (*model.Post, error) {
	member, err := s.repo.IsGroupMember(groupID, authorID)
	if err != nil {
		return nil, err
	}
	if !member {
		return nil, ErrNotGroupMember
	}
	if !validPrivacy(privacy) {
		privacy = model.PostPublic
	}
	post := &model.Post{AuthorID: authorID, Content: content, Privacy: privacy, GroupID: &groupID}
	if err := s.repo.CreatePost(post); err != nil {
		return nil, err
	}
	created, err := s.repo.GetPost(post.ID)
	if err != nil {
		return nil, err
	}
	if err := s.repo.LoadPostReactions(created, authorID); err != nil {
		return nil, err
	}
	return created, nil
}

// GroupPosts lists the posts of one group. Members only.
func (s *Service) GroupPosts(viewerID, groupID int64) ([]*model.Post, error) {
	member, err := s.repo.IsGroupMember(groupID, viewerID)
	if err != nil {
		return nil, err
	}
	if !member {
		return nil, ErrNotGroupMember
	}
	return s.repo.ListGroupPosts(groupID, viewerID)
}

// DeleteGroupPost removes one group post. The current user is always taken
// from the session (the handler passes it, never trusting client-sent ids or
// admin flags): the post must exist and belong to the group in the URL, the
// caller must be a member of that group, and only the post's author or the
// group's creator (the project's group-admin role) may delete it.
func (s *Service) DeleteGroupPost(userID, groupID, postID int64) error {
	// The post must exist and be a group post. A post from another group (or
	// a non-group post) is invisible in this group context: both answer
	// ErrNotFound so ids cannot be swapped to reach other groups' posts.
	postGroupID, err := s.repo.GetGroupIDForPost(postID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	if postGroupID != groupID {
		return ErrNotFound
	}

	// The group's creator (the project's group-admin role) may delete any
	// post in their group, whoever wrote it.
	isCreator, err := s.repo.IsGroupCreator(groupID, userID)
	if err != nil {
		return err
	}
	if isCreator {
		return s.deletePostRow(postID)
	}

	// Everyone else must be a member of the group — checked before anything
	// else, so outsiders get the same "not allowed" answer no matter which
	// post id they guess.
	member, err := s.repo.IsGroupMember(groupID, userID)
	if err != nil {
		return err
	}
	if !member {
		return ErrNotGroupMember
	}

	// …and the author of the post. Anything else is forbidden.
	authorID, err := s.repo.GetPostAuthor(postID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	if authorID != userID {
		return ErrForbidden
	}
	return s.deletePostRow(postID)
}

// deletePostRow removes the post row itself. ErrNotFound when the post
// vanished between the authorization checks and the delete.
func (s *Service) deletePostRow(postID int64) error {
	if err := s.repo.DeletePost(postID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

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
