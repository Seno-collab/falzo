package upload

import (
	"context"
	"errors"
	"io"
	"slices"
	"testing"
)

type fakeImageRepository struct {
	image     *Image
	saveErr   error
	findErr   error
	updateErr error
}

func (f *fakeImageRepository) Save(ctx context.Context, image *Image) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	image.ID = 1
	f.image = image
	return nil
}

func (f *fakeImageRepository) FindByID(ctx context.Context, id int64) (*Image, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	if f.image == nil {
		return nil, ErrImageNotFound
	}
	copy := *f.image
	return &copy, nil
}

func (f *fakeImageRepository) Update(ctx context.Context, image *Image) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.image = image
	return nil
}

type fakeImageStorage struct {
	uploadErr error
	uploaded  []string
	deleted   []string
}

func (f *fakeImageStorage) Upload(ctx context.Context, objectKey string, data io.Reader, contentType string) (string, error) {
	if f.uploadErr != nil {
		return "", f.uploadErr
	}
	f.uploaded = append(f.uploaded, objectKey)
	return "https://cdn.example.com/" + objectKey, nil
}

func (f *fakeImageStorage) Delete(ctx context.Context, objectKey string) error {
	f.deleted = append(f.deleted, objectKey)
	return nil
}

func TestUploadImage(t *testing.T) {
	repo := &fakeImageRepository{}
	storage := &fakeImageStorage{}
	service := NewService(repo, storage)

	result, err := service.UploadImage(t.Context(), UploadImageInput{
		File:     []byte("png"),
		FileName: "avatar.png",
		MimeType: "image/png",
		Size:     3,
		OwnerID:  "7",
	})
	if err != nil {
		t.Fatalf("upload image: %v", err)
	}

	if result.ID != 1 || result.URL == "" || result.ObjectKey == "" {
		t.Fatalf("expected uploaded image result, got %+v", result)
	}
	if repo.image == nil || repo.image.Status != "ready" {
		t.Fatalf("expected ready image saved, got %+v", repo.image)
	}
}

func TestUploadImageDeletesNewObjectWhenSaveFails(t *testing.T) {
	repo := &fakeImageRepository{saveErr: errors.New("db down")}
	storage := &fakeImageStorage{}
	service := NewService(repo, storage)

	_, err := service.UploadImage(t.Context(), UploadImageInput{
		File:     []byte("png"),
		FileName: "avatar.png",
		MimeType: "image/png",
		Size:     3,
		OwnerID:  "7",
	})
	if err == nil {
		t.Fatal("expected upload error")
	}
	if len(storage.uploaded) != 1 || !slices.Equal(storage.uploaded, storage.deleted) {
		t.Fatalf("expected uploaded object rollback, uploaded=%v deleted=%v", storage.uploaded, storage.deleted)
	}
}

func TestUpdateImageReplacesObjectAndDeletesOldObject(t *testing.T) {
	repo := &fakeImageRepository{image: &Image{
		ID:        10,
		OwnerID:   "7",
		ObjectKey: "7/old.png",
		URL:       "https://cdn.example.com/7/old.png",
		MimeType:  "image/png",
		Size:      3,
		Status:    "ready",
	}}
	storage := &fakeImageStorage{}
	service := NewService(repo, storage)

	result, err := service.UpdateImage(t.Context(), UpdateImageInput{
		ImageID:  10,
		File:     []byte("new"),
		FileName: "avatar.webp",
		MimeType: "image/webp",
		Size:     3,
	})
	if err != nil {
		t.Fatalf("update image: %v", err)
	}

	if result.ID != 10 || result.URL == "" || result.ObjectKey == "7/old.png" {
		t.Fatalf("expected replacement result, got %+v", result)
	}
	if !slices.Contains(storage.deleted, "7/old.png") {
		t.Fatalf("expected old object delete, deleted=%v", storage.deleted)
	}
}

func TestUpdateImageDeletesNewObjectWhenUpdateFails(t *testing.T) {
	repo := &fakeImageRepository{
		image: &Image{
			ID:        10,
			OwnerID:   "7",
			ObjectKey: "7/old.png",
			URL:       "https://cdn.example.com/7/old.png",
			MimeType:  "image/png",
			Size:      3,
			Status:    "ready",
		},
		updateErr: errors.New("db down"),
	}
	storage := &fakeImageStorage{}
	service := NewService(repo, storage)

	_, err := service.UpdateImage(t.Context(), UpdateImageInput{
		ImageID:  10,
		File:     []byte("new"),
		FileName: "avatar.webp",
		MimeType: "image/webp",
		Size:     3,
	})
	if err == nil {
		t.Fatal("expected update error")
	}
	if len(storage.uploaded) != 1 || !slices.Contains(storage.deleted, storage.uploaded[0]) {
		t.Fatalf("expected new object rollback, uploaded=%v deleted=%v", storage.uploaded, storage.deleted)
	}
	if slices.Contains(storage.deleted, "7/old.png") {
		t.Fatalf("old object must not be deleted when DB update fails, deleted=%v", storage.deleted)
	}
}
