package upload

import (
	"context"
	"encoding/base64"
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

func tinyPNG(t *testing.T) []byte {
	t.Helper()

	data, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII=")
	if err != nil {
		t.Fatalf("decode png fixture: %v", err)
	}

	return data
}

func TestUploadImage(t *testing.T) {
	repo := &fakeImageRepository{}
	storage := &fakeImageStorage{}
	service := NewService(repo, storage)
	file := tinyPNG(t)

	result, err := service.UploadImage(t.Context(), UploadImageInput{
		File:     file,
		FileName: "avatar.png",
		MimeType: "image/png",
		Size:     int64(len(file)),
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
	file := tinyPNG(t)

	_, err := service.UploadImage(t.Context(), UploadImageInput{
		File:     file,
		FileName: "avatar.png",
		MimeType: "image/png",
		Size:     int64(len(file)),
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
	file := tinyPNG(t)

	result, err := service.UpdateImage(t.Context(), UpdateImageInput{
		ImageID:  10,
		File:     file,
		FileName: "avatar.png",
		MimeType: "image/png",
		Size:     int64(len(file)),
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
	file := tinyPNG(t)

	_, err := service.UpdateImage(t.Context(), UpdateImageInput{
		ImageID:  10,
		File:     file,
		FileName: "avatar.png",
		MimeType: "image/png",
		Size:     int64(len(file)),
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

func TestCheckImageRejectsInvalidImageContent(t *testing.T) {
	service := NewService(nil, nil)

	_, err := service.CheckImage(t.Context(), CheckImageInput{
		File:     []byte("not an image"),
		MimeType: "image/png",
		Size:     int64(len("not an image")),
	})
	if !errors.Is(err, ErrInvalidImageContent) {
		t.Fatalf("expected invalid image content, got %v", err)
	}
}

func TestCheckImageReturnsDimensions(t *testing.T) {
	service := NewService(nil, nil)
	file := tinyPNG(t)

	result, err := service.CheckImage(t.Context(), CheckImageInput{
		File:     file,
		MimeType: "image/png",
		Size:     int64(len(file)),
	})
	if err != nil {
		t.Fatalf("check image: %v", err)
	}

	if !result.Valid || result.Width != 1 || result.Height != 1 {
		t.Fatalf("expected valid 1x1 image, got %+v", result)
	}
}
