package commentsvc

import (
	"errors"
	"mime/multipart"
	"strings"

	"sn-backend/internal/model"
	"sn-backend/internal/repository"
	"sn-backend/internal/service/filesvc"
	"sn-backend/internal/service/notificationsvc"
)

const maxContentLen = 2000

var (
	ErrInvalidContent = errors.New("comment: content is required (max 2000 chars)")
	ErrNotFound       = errors.New("comment: not found")
	ErrNoAccess       = errors.New("comment: no access to this post")
)

func checkContent(content string, hasAttachments bool) (string, error) {
	content = strings.TrimSpace(content)
	if content == "" && hasAttachments {
		return "", nil
	}
	if content == "" || len(content) > maxContentLen {
		return "", ErrInvalidContent
	}
	return content, nil
}

// Service uses the hub to notify a post's author of new comments.
type Service struct {
	repos         *repository.Repositories
	repo          *repository.CommentRepository
	posts         *repository.PostRepository
	files         *filesvc.Service
	notifications notificationsvc.Notifier
}

func New(repos *repository.Repositories, files *filesvc.Service, notifications notificationsvc.Notifier) *Service {
	return &Service{repos: repos, repo: repos.Comments, posts: repos.Posts, files: files, notifications: notifications}
}

// List returns the comments of a post the viewer can see.
func (s *Service) List(viewerID, postID, lastID int64) ([]*model.Comment, error) {
	visible, err := s.posts.CanViewPost(viewerID, postID)
	if err != nil {
		return nil, err
	}
	if !visible {
		return nil, ErrNoAccess
	}
	return s.repo.ListPostComments(postID, lastID)
}

// Create adds a comment to a post the viewer can see. Authorization goes
// through CanViewPost: post privacy for normal posts, group membership for
// group posts — a non-member cannot comment on a group post.
func (s *Service) Create(authorID, postID int64, content string, headers []*multipart.FileHeader) (*model.Comment, error) {
	content, err := checkContent(content, len(headers) > 0)
	if err != nil {
		return nil, err
	}
	visible, err := s.posts.CanViewPost(authorID, postID)
	if err != nil {
		return nil, err
	}
	if !visible {
		return nil, ErrNoAccess
	}

	var staged []*model.File
	if len(headers) > 0 {
		staged, err = s.files.Stage(authorID, headers)
		if err != nil {
			return nil, err
		}
	}

	comment := &model.Comment{PostID: postID, AuthorID: authorID, Content: content}
	err = s.repos.WithinTx(func(tx *repository.Repositories) error {
		if err := tx.Comments.CreateComment(comment); err != nil {
			return err
		}
		for _, file := range staged {
			file.CommentID = &comment.ID
			if err := tx.Files.CreateFile(file); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		s.files.Discard(staged)
		return nil, err
	}

	created, err := s.reload(comment.ID, postID)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *Service) NotifyCreated(authorID, postID int64, created *model.Comment) {
	if post, err := s.posts.GetPost(postID); err == nil && post.AuthorID != authorID {
		s.notifications.Notify(&model.Notification{
			UserID:  post.AuthorID,
			Type:    model.NotificationCommentPost,
			ActorID: authorID,
			Content: created.AuthorFirstName + " " + created.AuthorLastName + " commented on your post",
		})
	}
}

// reload re-reads the freshly inserted comment with its author fields and
// images, mirroring how the list endpoint renders comments.
func (s *Service) reload(commentID, postID int64) (*model.Comment, error) {
	comment, err := s.repo.GetComment(commentID)
	if err != nil {
		return nil, err
	}
	if comment.PostID != postID {
		return nil, repository.ErrNotFound
	}
	return comment, nil
}

// Update changes the text of a comment the caller wrote.
func (s *Service) Update(authorID, commentID int64, content string) (*model.Comment, error) {
	content, err := checkContent(content, false)
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpdateCommentOwned(commentID, authorID, content); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return s.repo.GetComment(commentID)
}

// Delete removes a comment. Its author may delete it, and so may the author of
// the post it sits under.
func (s *Service) Delete(userID, commentID int64) error {
	if err := s.repo.DeleteCommentOwned(commentID, userID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}
