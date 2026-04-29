package upload

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"
)

var (
	ErrInvalidFileSize   = errors.New("invalid file size")
	ErrInvalidMimeType   = errors.New("invalid mime type")
	ErrImageNotFound     = errors.New("image not found")
	ErrStorageFailed     = errors.New("failed to upload image to storage")
	ErrFileSizeTooLarge  = errors.New("file size exceeds the maximum allowed limit")
	ErrInvalidImageURL   = errors.New("invalid image URL")
	ErrFileRequired      = errors.New("file is required")
	ErrOwnerIDRequired   = errors.New("owner id is required")
	ErrMissingRepository = errors.New("image repository is unavailable")
)

type ImageRepository interface {
	Save(ctx context.Context, image *Image) error
}

type ImageStorage interface {
	Upload(ctx context.Context, objectKey string, data io.Reader, contentType string) (url string, err error)
}

type Image struct {
	ID        int64
	OwnerID   string
	ObjectKey string
	URL       string
	MimeType  string
	Size      int64
	Status    string
	CreatedAt time.Time
}

func NewImage(ownerID string, objectKey string, mimeType string, size int64) (*Image, error) {
	if strings.TrimSpace(ownerID) == "" {
		return nil, ErrOwnerIDRequired
	}
	if _, err := NewFileSize(size); err != nil {
		return nil, err
	}
	if strings.TrimSpace(mimeType) == "" {
		return nil, ErrInvalidMimeType
	}

	return &Image{
		OwnerID:   strings.TrimSpace(ownerID),
		ObjectKey: objectKey,
		MimeType:  mimeType,
		Size:      size,
		Status:    "pending",
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (i *Image) MarkProcessing() {
	i.Status = "processing"
}

func (i *Image) MarkReady(url string) error {
	imageURL, err := NewImageURL(url)
	if err != nil {
		return err
	}

	i.Status = "ready"
	i.URL = imageURL.String()
	return nil
}

func (i *Image) MarkFailed() {
	i.Status = "failed"
}

type FileSize int64

func NewFileSize(value int64) (FileSize, error) {
	if value <= 0 {
		return 0, ErrInvalidFileSize
	}
	if value > 10*1024*1024 {
		return 0, ErrFileSizeTooLarge
	}
	return FileSize(value), nil
}

func (fs FileSize) Value() int64 { return int64(fs) }

type ImageURL string

func NewImageURL(value string) (ImageURL, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ErrInvalidImageURL
	}
	return ImageURL(value), nil
}

func (iu ImageURL) String() string { return string(iu) }
