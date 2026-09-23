package commentsvc

import (
	"errors"
	"strings"

	"sn-backend/internal/model"
	"sn-backend/internal/repository"
	ws "sn-backend/internal/websocket"
)

const maxContentLen = 2000

var (
	ErrInvalidContent = errors.New("comment: content is required (max 2000 chars)")
	ErrNotFound       = errors.New("comment: not found")
	ErrNoAccess       = errors.New("comment: no access to this post")
)

// Repository is the subset of repository.Repository the comment service needs.
type Repository interface {
	CreateComment(*model.Comment) error
	ListPostComments(int64) ([]*model.Comment, error)
	CanViewPost(int64, int64) (bool, error)
	GetPost(int64) (*model.Post, error)
	CreateNotification(*model.Notification) error
}

// Service reuses the existing notification plumbing (persist +
// hub fan-out), the same pattern as groupsvc.notify.
type Service struct {
	repo Repository
	hub  *ws.Hub
}

func New(repo Repository, hub *ws.Hub) *Service {
	return &Service{repo: repo, hub: hub}
}

// List returns the comments of a post the viewer can see.
func (s *Service) List(viewerID, postID int64) ([]*model.Comment, error) {
	visible, err := s.repo.CanViewPost(viewerID, postID)
	if err != nil {
		return nil, err
	}
	if !visible {
		return nil, ErrNoAccess
	}
	return s.repo.ListPostComments(postID)
}

// Create adds a comment to a post the viewer can see. Authorization goes
// through CanViewPost: post privacy for normal posts, group membership for
// group posts — a non-member cannot comment on a group post.
func (s *Service) Create(authorID, postID int64, content string) (*model.Comment, error) {
	content = strings.TrimSpace(content)
	if content == "" || len(content) > maxContentLen {
		return nil, ErrInvalidContent
	}
	visible, err := s.repo.CanViewPost(authorID, postID)
	if err != nil {
		return nil, err
	}
	if !visible {
		return nil, ErrNoAccess
	}

	comment := &model.Comment{PostID: postID, AuthorID: authorID, Content: content}
	if err := s.repo.CreateComment(comment); err != nil {
		return nil, err
	}

	// Best-effort notification to the post author (not when they comment
	// on their own post). Persistence failures never fail the comment.
	if post, err := s.repo.GetPost(postID); err == nil && post.AuthorID != authorID {
		notification := &model.Notification{
			UserID:  post.AuthorID,
			Type:    model.NotificationCommentPost,
			ActorID: authorID,
			Content: "commented on your post",
		}
		s.notify(notification)
	}

	created, err := s.reload(comment.ID, postID)
	if err != nil {
		return nil, err
	}
	return created, nil
}

// reload re-reads the freshly inserted comment with its author fields and
// images, mirroring how the list endpoint renders comments.
func (s *Service) reload(commentID, postID int64) (*model.Comment, error) {
	comments, err := s.repo.ListPostComments(postID)
	if err != nil {
		return nil, err
	}
	for _, comment := range comments {
		if comment.ID == commentID {
			return comment, nil
		}
	}
	return nil, repository.ErrNotFound
}

// notify persists a notification and pushes it to the user's websockets via
// the existing hub (same contract as groupsvc.notify).
func (s *Service) notify(n *model.Notification) {
	if err := s.repo.CreateNotification(n); err != nil {
		return
	}
	if s.hub != nil {
		s.hub.PublishNotification(n)
	}
}
