package upload

import (
	"context"
	"errors"
	"io"
	"net/url"
	"strings"
	"time"
)

var (
	ErrInvalidFileSize       = errors.New("invalid file size")
	ErrInvalidMimeType       = errors.New("invalid mime type")
	ErrImageNotFound         = errors.New("image not found")
	ErrStorageFailed         = errors.New("failed to upload image to storage")
	ErrDependencyUnavailable = errors.New("upload dependency unavailable")
	ErrInternal              = errors.New("upload internal error")
	ErrFileSizeTooLarge      = errors.New("file size exceeds the maximum allowed limit")
	ErrInvalidImageURL       = errors.New("invalid image URL")
	ErrInvalidImageContent   = errors.New("invalid image content")
	ErrFileRequired          = errors.New("file is required")
	ErrOwnerIDRequired       = errors.New("owner id is required")
	ErrMissingRepository     = errors.New("image repository is unavailable")
	ErrImageIDRequired       = errors.New("image id is required")
)

type ImageRepository interface {
	Save(ctx context.Context, image *Image) error
	FindByID(ctx context.Context, id int64) (*Image, error)
	Update(ctx context.Context, image *Image) error
}

type ImageStorage interface {
	Upload(ctx context.Context, objectKey string, data io.Reader, contentType string) (url string, err error)
	Delete(ctx context.Context, objectKey string) error
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
	UpdatedAt time.Time
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
	if strings.TrimSpace(objectKey) == "" {
		return nil, ErrInvalidImageURL
	}

	now := time.Now().UTC()
	return &Image{
		OwnerID:   strings.TrimSpace(ownerID),
		ObjectKey: strings.TrimSpace(objectKey),
		MimeType:  strings.TrimSpace(mimeType),
		Size:      size,
		Status:    "pending",
		CreatedAt: now,
		UpdatedAt: now,
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
	i.UpdatedAt = time.Now().UTC()
	return nil
}

func (i *Image) MarkFailed() {
	i.Status = "failed"
	i.UpdatedAt = time.Now().UTC()
}

func (i *Image) Replace(objectKey string, url string, mimeType string, size int64) error {
	if strings.TrimSpace(objectKey) == "" {
		return ErrInvalidImageURL
	}
	if _, err := NewFileSize(size); err != nil {
		return err
	}
	if strings.TrimSpace(mimeType) == "" {
		return ErrInvalidMimeType
	}
	imageURL, err := NewImageURL(url)
	if err != nil {
		return err
	}

	i.ObjectKey = strings.TrimSpace(objectKey)
	i.URL = imageURL.String()
	i.MimeType = strings.TrimSpace(mimeType)
	i.Size = size
	i.Status = "ready"
	i.UpdatedAt = time.Now().UTC()
	return nil
}

type FileSize int64

func NewFileSize(value int64) (FileSize, error) {
	if value <= 0 {
		return 0, ErrInvalidFileSize
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
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed == nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", ErrInvalidImageURL
	}
	return ImageURL(value), nil
}

func (iu ImageURL) String() string { return string(iu) }
