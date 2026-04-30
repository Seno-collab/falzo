package upload

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

const defaultMaxImageSize = 10 << 20

type Service struct {
	images            ImageRepository
	storage           ImageStorage
	maxSize           int64
	allowedImageTypes map[string]struct{}
}

type ServiceOption func(*Service)

func NewService(images ImageRepository, storage ImageStorage, options ...ServiceOption) *Service {
	s := &Service{
		images:            images,
		storage:           storage,
		maxSize:           defaultMaxImageSize,
		allowedImageTypes: allowedImageTypes([]string{"image/jpeg", "image/png", "image/webp"}),
	}
	for _, option := range options {
		option(s)
	}
	return s
}

func WithMaxSize(maxSize int64) ServiceOption {
	return func(s *Service) {
		if maxSize > 0 {
			s.maxSize = maxSize
		}
	}
}

func WithAllowedImageTypes(types []string) ServiceOption {
	return func(s *Service) {
		allowed := allowedImageTypes(types)
		if len(allowed) > 0 {
			s.allowedImageTypes = allowed
		}
	}
}

type UploadImageInput struct {
	File     []byte
	FileName string
	MimeType string
	Size     int64
	OwnerID  string
}

type UploadImageResult struct {
	ID        int64  `json:"id"`
	URL       string `json:"url"`
	ObjectKey string `json:"object_key"`
}

type UpdateImageInput struct {
	ImageID  int64
	File     []byte
	FileName string
	MimeType string
	Size     int64
}

type UpdateImageResult struct {
	ID        int64  `json:"id"`
	URL       string `json:"url"`
	ObjectKey string `json:"object_key"`
}

func (s *Service) UploadImage(ctx context.Context, input UploadImageInput) (UploadImageResult, error) {
	if len(input.File) == 0 {
		return UploadImageResult{}, ErrFileRequired
	}
	if s.images == nil {
		return UploadImageResult{}, ErrMissingRepository
	}
	if s.storage == nil {
		return UploadImageResult{}, ErrDependencyUnavailable
	}
	if err := s.validateFile(input.MimeType, input.Size); err != nil {
		return UploadImageResult{}, err
	}

	objectKey := imageObjectKey(input.OwnerID, input.FileName, input.MimeType)
	image, err := NewImage(input.OwnerID, objectKey, input.MimeType, input.Size)
	if err != nil {
		return UploadImageResult{}, err
	}
	image.MarkProcessing()

	url, err := s.storage.Upload(ctx, objectKey, bytes.NewReader(input.File), input.MimeType)
	if err != nil {
		return UploadImageResult{}, ErrStorageFailed
	}
	if err := image.MarkReady(url); err != nil {
		_ = s.storage.Delete(ctx, objectKey)
		return UploadImageResult{}, err
	}

	if err := s.images.Save(ctx, image); err != nil {
		_ = s.storage.Delete(ctx, objectKey)
		return UploadImageResult{}, err
	}

	return UploadImageResult{ID: image.ID, URL: image.URL, ObjectKey: image.ObjectKey}, nil
}

func (s *Service) UpdateImage(ctx context.Context, input UpdateImageInput) (UpdateImageResult, error) {
	if input.ImageID <= 0 {
		return UpdateImageResult{}, ErrImageIDRequired
	}
	if len(input.File) == 0 {
		return UpdateImageResult{}, ErrFileRequired
	}
	if s.images == nil {
		return UpdateImageResult{}, ErrMissingRepository
	}
	if s.storage == nil {
		return UpdateImageResult{}, ErrDependencyUnavailable
	}
	if err := s.validateFile(input.MimeType, input.Size); err != nil {
		return UpdateImageResult{}, err
	}

	image, err := s.images.FindByID(ctx, input.ImageID)
	if err != nil {
		return UpdateImageResult{}, err
	}

	oldObjectKey := image.ObjectKey
	newObjectKey := imageObjectKey(image.OwnerID, input.FileName, input.MimeType)
	newURL, err := s.storage.Upload(ctx, newObjectKey, bytes.NewReader(input.File), input.MimeType)
	if err != nil {
		return UpdateImageResult{}, ErrStorageFailed
	}

	if err := image.Replace(newObjectKey, newURL, input.MimeType, input.Size); err != nil {
		_ = s.storage.Delete(ctx, newObjectKey)
		return UpdateImageResult{}, err
	}
	if err := s.images.Update(ctx, image); err != nil {
		_ = s.storage.Delete(ctx, newObjectKey)
		return UpdateImageResult{}, err
	}

	if oldObjectKey != "" && oldObjectKey != newObjectKey {
		_ = s.storage.Delete(ctx, oldObjectKey)
	}

	return UpdateImageResult{ID: image.ID, URL: image.URL, ObjectKey: image.ObjectKey}, nil
}

func (s *Service) validateFile(mimeType string, size int64) error {
	if _, err := NewFileSize(size); err != nil {
		return err
	}
	if size > s.maxSize {
		return ErrFileSizeTooLarge
	}
	if _, ok := s.allowedImageTypes[strings.ToLower(strings.TrimSpace(mimeType))]; !ok {
		return ErrInvalidMimeType
	}
	return nil
}

func imageObjectKey(ownerID string, fileName string, mimeType string) string {
	extension := imageExtension(mimeType)
	if extension == "" {
		extension = strings.ToLower(filepath.Ext(strings.TrimSpace(fileName)))
	}
	randomID, err := randomHex(16)
	if err != nil {
		randomID = strings.ReplaceAll(fmt.Sprintf("%p", &fileName), "0x", "")
	}
	owner := strings.TrimSpace(ownerID)
	if owner == "" {
		owner = "unknown"
	}
	return owner + "/" + randomID + extension
}

func imageExtension(mimeType string) string {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	default:
		return ""
	}
}

func allowedImageTypes(types []string) map[string]struct{} {
	allowed := make(map[string]struct{}, len(types))
	for _, value := range types {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		allowed[value] = struct{}{}
	}
	return allowed
}

func randomHex(size int) (string, error) {
	if size <= 0 {
		return "", errors.New("invalid random size")
	}
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
