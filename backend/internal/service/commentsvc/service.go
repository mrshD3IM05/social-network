package commentsvc

import (
	"errors"
	"strings"

	"sn-backend/internal/model"
	"sn-backend/internal/repository"
)

const MaxCommentLength = 1000

var (
	ErrEmptyContent    = errors.New("comment: content is required")
	ErrTooLong         = errors.New("comment: content is too long")
	ErrNotFound        = errors.New("comment: not found")
	ErrInvalidReaction = errors.New("comment: invalid reaction")
)

type Repository interface {
	CreateComment(*model.Comment) error
	GetComment(int64) (*model.Comment, error)
	ListPostComments(int64, int64) ([]*model.Comment, error)
	DeleteCommentOwned(int64, int64) ([]string, error)
	CommentPostID(int64) (int64, error)
	AttachFileToComment(string, int64, int64) error
	CanViewPost(int64, int64) (bool, error)
	GetReaction(string, int64, int64) (*model.Reaction, error)
	SetReaction(string, int64, int64, string) error
	DeleteReaction(string, int64, int64) error
	GetReactionSummary(string, int64, int64) (*model.ReactionSummary, error)
}

type Service struct{ repo Repository }

func New(repo Repository) *Service { return &Service{repo: repo} }

// Create adds a comment to a post the author is allowed to see. Images are
// attached afterwards with AttachImage, so an upload failure cannot leave the
// comment itself unsaved.
func (s *Service) Create(authorID, postID int64, content string) (*model.Comment, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, ErrEmptyContent
	}
	if len(content) > MaxCommentLength {
		return nil, ErrTooLong
	}
	visible, err := s.repo.CanViewPost(authorID, postID)
	if err != nil {
		return nil, err
	}
	if !visible {
		return nil, ErrNotFound
	}
	comment := &model.Comment{PostID: postID, AuthorID: authorID, Content: content}
	if err := s.repo.CreateComment(comment); err != nil {
		return nil, err
	}
	return s.repo.GetComment(comment.ID)
}

func (s *Service) List(viewerID, postID int64) ([]*model.Comment, error) {
	visible, err := s.repo.CanViewPost(viewerID, postID)
	if err != nil {
		return nil, err
	}
	if !visible {
		return nil, ErrNotFound
	}
	return s.repo.ListPostComments(postID, viewerID)
}

func (s *Service) Get(viewerID, commentID int64) (*model.Comment, error) {
	comment, err := s.repo.GetComment(commentID)
	if err != nil {
		return nil, ErrNotFound
	}
	visible, err := s.repo.CanViewPost(viewerID, comment.PostID)
	if err != nil {
		return nil, err
	}
	if !visible {
		return nil, ErrNotFound
	}
	return comment, nil
}

// Delete removes a comment written by the caller, or any comment sitting on a
// post the caller owns. It answers with the ids of the files to unlink.
func (s *Service) Delete(userID, commentID int64) ([]string, error) {
	fileIDs, err := s.repo.DeleteCommentOwned(commentID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return fileIDs, nil
}

// AttachImage binds an upload the caller already owns to their comment.
func (s *Service) AttachImage(ownerID, commentID int64, fileID string) error {
	comment, err := s.repo.GetComment(commentID)
	if err != nil {
		return ErrNotFound
	}
	if comment.AuthorID != ownerID {
		return ErrNotFound
	}
	if err := s.repo.AttachFileToComment(fileID, commentID, ownerID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *Service) React(viewerID, commentID int64, reaction string) (*model.ReactionSummary, error) {
	if reaction != model.ReactionLike && reaction != model.ReactionDislike {
		return nil, ErrInvalidReaction
	}
	if _, err := s.Get(viewerID, commentID); err != nil {
		return nil, err
	}
	existing, err := s.repo.GetReaction(model.ReactionTargetComment, commentID, viewerID)
	switch {
	case err == nil && existing.Reaction == reaction:
		err = s.repo.DeleteReaction(model.ReactionTargetComment, commentID, viewerID)
	case err == nil, errors.Is(err, repository.ErrNotFound):
		err = s.repo.SetReaction(model.ReactionTargetComment, commentID, viewerID, reaction)
	}
	if err != nil {
		return nil, err
	}
	return s.repo.GetReactionSummary(model.ReactionTargetComment, commentID, viewerID)
}

func (s *Service) Unreact(viewerID, commentID int64) (*model.ReactionSummary, error) {
	if _, err := s.Get(viewerID, commentID); err != nil {
		return nil, err
	}
	if err := s.repo.DeleteReaction(model.ReactionTargetComment, commentID, viewerID); err != nil {
		return nil, err
	}
	return s.repo.GetReactionSummary(model.ReactionTargetComment, commentID, viewerID)
}
