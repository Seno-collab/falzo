package upload

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
)

type Service struct {
	images  ImageRepository
	storage ImageStorage
}

func NewService(images ImageRepository, storage ImageStorage) *Service {
	return &Service{images: images, storage: storage}
}

type UploadImageInput struct {
	File     []byte
	FileName string
	MimeType string
	Size     int64
	OwnerID  string
}

type UploadImageResult struct {
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

	objectKey := imageObjectKey(input.OwnerID, input.FileName)
	image, err := NewImage(input.OwnerID, objectKey, input.MimeType, input.Size)
	if err != nil {
		return UploadImageResult{}, err
	}
	image.MarkProcessing()

	if s.storage != nil {
		url, err := s.storage.Upload(ctx, objectKey, bytes.NewReader(input.File), input.MimeType)
		if err != nil {
			image.MarkFailed()
			_ = s.images.Save(ctx, image)
			return UploadImageResult{}, ErrStorageFailed
		}
		if err := image.MarkReady(url); err != nil {
			return UploadImageResult{}, err
		}
	}

	if err := s.images.Save(ctx, image); err != nil {
		return UploadImageResult{}, err
	}

	return UploadImageResult{URL: image.URL, ObjectKey: image.ObjectKey}, nil
}

func imageObjectKey(ownerID string, fileName string) string {
	name := strings.TrimSpace(filepath.Base(fileName))
	if name == "." || name == "/" || name == "" {
		name = "image"
	}
	return strings.TrimSpace(ownerID) + "/" + name
}
