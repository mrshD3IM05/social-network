package filesvc

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sn-backend/internals/model"
	"sn-backend/internals/repository"
)

const (
	MaxImageSize int64 = 10 << 20
	MaxImages          = 5
)

var (
	ErrInvalidImage  = errors.New("file: only JPEG, PNG, and GIF images are allowed")
	ErrFileTooLarge  = errors.New("file: image exceeds the 10 MB limit")
	ErrTooManyImages = errors.New("file: a maximum of 5 images is allowed")
)

type Repository interface {
	CreateFile(*model.File) error
	GetFile(string) (*model.File, error)
	CanViewFile(int64, string) (bool, error)
	GetPost(int64) (*model.Post, error)
}
type Service struct {
	repo        Repository
	storagePath string
}

func New(repo Repository, storagePath string) *Service {
	return &Service{repo: repo, storagePath: storagePath}
}
func (s *Service) Upload(ownerID int64, header *multipart.FileHeader, postID *int64) (*model.File, error) {
	if header == nil || header.Size > MaxImageSize {
		return nil, ErrFileTooLarge
	}
	if postID != nil {
		post, err := s.repo.GetPost(*postID)
		if err != nil {
			return nil, err
		}
		if post.AuthorID != ownerID {
			return nil, repository.ErrNotFound
		}
	}
	source, err := header.Open()
	if err != nil {
		return nil, err
	}
	defer source.Close()
	contentType, err := detectImageType(source)
	if err != nil {
		return nil, err
	}
	if !allowedImageType(contentType) {
		return nil, ErrInvalidImage
	}
	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	id, err := randomID()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(s.storagePath, 0o750); err != nil {
		return nil, err
	}
	path := filepath.Join(s.storagePath, id)
	destination, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(destination, io.LimitReader(source, MaxImageSize+1)); err != nil {
		destination.Close()
		_ = os.Remove(path)
		return nil, err
	}
	if err := destination.Close(); err != nil {
		_ = os.Remove(path)
		return nil, err
	}
	file := &model.File{ID: id, StoragePath: path, OriginalName: filepath.Base(header.Filename), MIMEType: contentType, Size: header.Size, OwnerUserID: ownerID, PostID: postID}
	if err := s.repo.CreateFile(file); err != nil {
		_ = os.Remove(path)
		return nil, err
	}
	return file, nil
}
func (s *Service) UploadMany(ownerID int64, headers []*multipart.FileHeader, postID *int64) ([]*model.File, error) {
	if len(headers) == 0 {
		return nil, errors.New("file: at least one image is required")
	}
	if len(headers) > MaxImages {
		return nil, ErrTooManyImages
	}
	files := make([]*model.File, 0, len(headers))
	for _, header := range headers {
		file, err := s.Upload(ownerID, header, postID)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}
func (s *Service) Get(id string) (*model.File, error) { return s.repo.GetFile(id) }
func (s *Service) CanView(viewerID int64, id string) (bool, error) {
	return s.repo.CanViewFile(viewerID, id)
}
func detectImageType(source multipart.File) (string, error) {
	buffer := make([]byte, 512)
	count, err := source.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return http.DetectContentType(buffer[:count]), nil
}
func allowedImageType(contentType string) bool {
	return contentType == "image/jpeg" || contentType == "image/png" || contentType == "image/gif"
}
func randomID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes[:]), nil
}
