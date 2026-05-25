package upload

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
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

type CheckImageInput struct {
	File     []byte
	MimeType string
	Size     int64
}

type UploadImageResult struct {
	ID        int64  `json:"id"`
	URL       string `json:"url"`
	ObjectKey string `json:"object_key"`
}

type CheckImageResult struct {
	Valid    bool   `json:"valid"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

type UpdateImageInput struct {
	ImageID  int64
	File     []byte
	FileName string
	MimeType string
	Size     int64
	OwnerID  string
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
	if _, err := s.validateImageFile(input.File, input.MimeType, input.Size); err != nil {
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
		return UploadImageResult{}, storageUploadError(err)
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
	if strings.TrimSpace(input.OwnerID) == "" {
		return UpdateImageResult{}, ErrOwnerIDRequired
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
	if _, err := s.validateImageFile(input.File, input.MimeType, input.Size); err != nil {
		return UpdateImageResult{}, err
	}

	image, err := s.images.FindByID(ctx, input.ImageID)
	if err != nil {
		return UpdateImageResult{}, err
	}
	if image.OwnerID != strings.TrimSpace(input.OwnerID) {
		return UpdateImageResult{}, ErrImageNotFound
	}

	oldObjectKey := image.ObjectKey
	newObjectKey := imageObjectKey(image.OwnerID, input.FileName, input.MimeType)
	newURL, err := s.storage.Upload(ctx, newObjectKey, bytes.NewReader(input.File), input.MimeType)
	if err != nil {
		return UpdateImageResult{}, storageUploadError(err)
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

func (s *Service) CheckImage(ctx context.Context, input CheckImageInput) (CheckImageResult, error) {
	_ = ctx
	if len(input.File) == 0 {
		return CheckImageResult{}, ErrFileRequired
	}

	dimensions, err := s.validateImageFile(input.File, input.MimeType, input.Size)
	if err != nil {
		return CheckImageResult{}, err
	}

	return CheckImageResult{
		Valid:    true,
		MimeType: strings.ToLower(strings.TrimSpace(input.MimeType)),
		Size:     input.Size,
		Width:    dimensions.width,
		Height:   dimensions.height,
	}, nil
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

type imageDimensions struct {
	width  int
	height int
}

func (s *Service) validateImageFile(file []byte, mimeType string, size int64) (imageDimensions, error) {
	if err := s.validateFile(mimeType, size); err != nil {
		return imageDimensions{}, err
	}

	dimensions, err := readImageDimensions(file, mimeType)
	if err != nil {
		return imageDimensions{}, err
	}

	if dimensions.width <= 0 || dimensions.height <= 0 {
		return imageDimensions{}, ErrInvalidImageContent
	}

	return dimensions, nil
}

func readImageDimensions(file []byte, mimeType string) (imageDimensions, error) {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "image/jpeg", "image/png":
		config, _, err := image.DecodeConfig(bytes.NewReader(file))
		if err != nil {
			return imageDimensions{}, ErrInvalidImageContent
		}

		return imageDimensions{width: config.Width, height: config.Height}, nil
	case "image/webp":
		return readWebPDimensions(file)
	default:
		return imageDimensions{}, ErrInvalidMimeType
	}
}

func readWebPDimensions(file []byte) (imageDimensions, error) {
	if len(file) < 30 || string(file[0:4]) != "RIFF" || string(file[8:12]) != "WEBP" {
		return imageDimensions{}, ErrInvalidImageContent
	}

	chunkType := string(file[12:16])
	switch chunkType {
	case "VP8 ":
		if len(file) < 30 || file[23] != 0x9d || file[24] != 0x01 || file[25] != 0x2a {
			return imageDimensions{}, ErrInvalidImageContent
		}
		width := int(uint16(file[26])|uint16(file[27])<<8) & 0x3fff
		height := int(uint16(file[28])|uint16(file[29])<<8) & 0x3fff
		return imageDimensions{width: width, height: height}, nil
	case "VP8L":
		if len(file) < 25 || file[20] != 0x2f {
			return imageDimensions{}, ErrInvalidImageContent
		}
		b0 := uint32(file[21])
		b1 := uint32(file[22])
		b2 := uint32(file[23])
		b3 := uint32(file[24])
		width := int(1 + (((b1 & 0x3f) << 8) | b0))
		height := int(1 + ((b3 << 6) | (b2 << 2) | ((b1 & 0xc0) >> 6)))
		return imageDimensions{width: width, height: height}, nil
	case "VP8X":
		if len(file) < 30 {
			return imageDimensions{}, ErrInvalidImageContent
		}
		width := 1 + int(uint32(file[24])|uint32(file[25])<<8|uint32(file[26])<<16)
		height := 1 + int(uint32(file[27])|uint32(file[28])<<8|uint32(file[29])<<16)
		return imageDimensions{width: width, height: height}, nil
	default:
		return imageDimensions{}, ErrInvalidImageContent
	}
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

func storageUploadError(err error) error {
	if errors.Is(err, ErrStorageFailed) || errors.Is(err, ErrDependencyUnavailable) {
		return err
	}
	return fmt.Errorf("%w: %v", ErrStorageFailed, err)
}
