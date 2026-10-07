package filesvc

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sn-backend/internal/model"
	"sn-backend/internal/repository"
)

const (
	MaxImageSize int64 = 10 << 20
	MaxImages          = 3
	// MaxImageSide is the largest width or height accepted, in pixels. It keeps
	// "image bombs" (a tiny file that expands to a huge picture) out.
	MaxImageSide = 8000
	// MaxRequestSize is the biggest upload body: every image plus the form fields.
	MaxRequestSize = MaxImageSize*MaxImages + 1<<20
	// MaxMemory is how much of an upload is held in memory while it is read.
	// Anything past it goes to a temporary file, so a few big uploads at the
	// same time cannot fill the server's memory.
	MaxMemory = 1 << 20
)

var (
	ErrInvalidImage    = errors.New("file: only JPEG, PNG, and GIF images are allowed")
	ErrFileTooLarge    = errors.New("file: image exceeds the 10 MB limit")
	ErrTooManyImages   = errors.New("file: a maximum of 3 images is allowed")
	ErrImageDimensions = errors.New("file: image is larger than 8000x8000 pixels")
)

// IsBadImage says whether err is the client's fault (answer 400).
func IsBadImage(err error) bool {
	return errors.Is(err, ErrInvalidImage) || errors.Is(err, ErrFileTooLarge) ||
		errors.Is(err, ErrTooManyImages) || errors.Is(err, ErrImageDimensions)
}

type Service struct {
	repos       *repository.Repositories
	repo        *repository.FileRepository
	storagePath string
}

func New(repos *repository.Repositories, storagePath string) *Service {
	return &Service{
		repos:       repos,
		repo:        repos.Files,
		storagePath: storagePath,
	}
}

// CheckImage makes sure the upload really is a JPEG, PNG or GIF picture and
// answers with its type. The file name and the browser's type are never
// trusted: the first bytes decide the type, then the image header is decoded
// to prove it matches and to read the size in pixels.
func CheckImage(header *multipart.FileHeader) (string, error) {
	if header == nil || header.Size <= 0 {
		return "", ErrInvalidImage
	}
	if header.Size > MaxImageSize {
		return "", ErrFileTooLarge
	}
	source, err := header.Open()
	if err != nil {
		return "", err
	}
	defer source.Close()

	contentType, err := detectImageType(source)
	if err != nil {
		return "", err
	}
	if !allowedImageType(contentType) {
		return "", ErrInvalidImage
	}
	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	config, format, err := image.DecodeConfig(source)
	if err != nil || "image/"+format != contentType {
		return "", ErrInvalidImage
	}
	if config.Width < 1 || config.Height < 1 {
		return "", ErrInvalidImage
	}
	if config.Width > MaxImageSide || config.Height > MaxImageSide {
		return "", ErrImageDimensions
	}
	return contentType, nil
}

// Stage validates a batch of images and writes their bytes to disk without
// recording them, so a caller can attach them inside its own transaction. Every
// staged file must later be attached or removed with Discard.
func (s *Service) Stage(ownerID int64, headers []*multipart.FileHeader) ([]*model.File, error) {
	if len(headers) == 0 {
		return nil, errors.New("file: at least one image is required")
	}
	if len(headers) > MaxImages {
		return nil, ErrTooManyImages
	}
	staged := make([]*model.File, 0, len(headers))
	for _, header := range headers {
		contentType, err := CheckImage(header)
		if err != nil {
			s.Discard(staged)
			return nil, err
		}
		id, path, err := s.writeImage(header)
		if err != nil {
			s.Discard(staged)
			return nil, err
		}
		staged = append(staged, &model.File{ID: id, StoragePath: path, OriginalName: filepath.Base(header.Filename), MIMEType: contentType, Size: header.Size, OwnerUserID: &ownerID})
	}
	return staged, nil
}

// Discard removes staged files' bytes from disk. It is best effort: a file that
// is already gone is not an error.
func (s *Service) Discard(files []*model.File) {
	for _, file := range files {
		if file != nil && file.StoragePath != "" {
			_ = os.Remove(file.StoragePath)
		}
	}
}

// writeImage streams one image to a new file under the storage path and returns
// its id and path. The bytes are written before any database row exists, so a
// caller that fails to record the file can remove it with Discard.
func (s *Service) writeImage(header *multipart.FileHeader) (string, string, error) {
	source, err := header.Open()
	if err != nil {
		return "", "", err
	}
	defer source.Close()
	id, err := randomID()
	if err != nil {
		return "", "", err
	}
	if err := os.MkdirAll(s.storagePath, 0o750); err != nil {
		return "", "", err
	}
	path := filepath.Join(s.storagePath, id)
	destination, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return "", "", err
	}
	if _, err := io.Copy(destination, io.LimitReader(source, MaxImageSize+1)); err != nil {
		destination.Close()
		_ = os.Remove(path)
		return "", "", err
	}
	if err := destination.Close(); err != nil {
		_ = os.Remove(path)
		return "", "", err
	}
	return id, path, nil
}

func (s *Service) Get(id string) (*model.File, error) { return s.repo.GetFile(id) }

func (s *Service) CanView(viewerID int64, id string) (bool, error) {
	return s.repo.CanViewFile(viewerID, id)
}

// SetAvatar stages the picture, then records it and points the user at it in one
// transaction, so a failure never leaves a half-applied avatar.
func (s *Service) SetAvatar(ownerID int64, header *multipart.FileHeader) (*model.User, error) {
	staged, err := s.Stage(ownerID, []*multipart.FileHeader{header})
	if err != nil {
		return nil, err
	}
	file := staged[0]
	var user *model.User
	err = s.repos.WithinTx(func(tx *repository.Repositories) error {
		if err := tx.Files.CreateFile(file); err != nil {
			return err
		}
		u, err := tx.Users.GetUserByID(ownerID)
		if err != nil {
			return err
		}
		u.Avatar = file.ID
		if err := tx.Users.UpdateUser(u); err != nil {
			return err
		}
		user = u
		return nil
	})
	if err != nil {
		s.Discard(staged)
		return nil, err
	}
	return user, nil
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
