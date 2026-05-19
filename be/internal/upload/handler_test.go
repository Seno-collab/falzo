package upload

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"falzo-be/internal/auth"
)

type fakeUploadHandlerService struct{}

func (fakeUploadHandlerService) CheckImage(context.Context, CheckImageInput) (CheckImageResult, error) {
	return CheckImageResult{Valid: true, MimeType: "image/png", Size: 67, Width: 1, Height: 1}, nil
}

func (fakeUploadHandlerService) UploadImage(context.Context, UploadImageInput) (UploadImageResult, error) {
	return UploadImageResult{ID: 1, URL: "https://cdn.example.com/1.png", ObjectKey: "7/1.png"}, nil
}

func (fakeUploadHandlerService) UpdateImage(context.Context, UpdateImageInput) (UpdateImageResult, error) {
	return UpdateImageResult{ID: 1, URL: "https://cdn.example.com/1.png", ObjectKey: "7/1.png"}, nil
}

type fakeUploadAuthService struct{}

func (fakeUploadAuthService) Authenticate(context.Context, string) (*auth.AuthenticatedUser, error) {
	return &auth.AuthenticatedUser{UserID: 7, Username: "admin"}, nil
}

func TestUploadImageRejectsBodyOverLimit(t *testing.T) {
	body, contentType := multipartImageBody(t, strings.Repeat("a", 128))
	handler := NewHandler(fakeUploadHandlerService{}, fakeUploadAuthService{}, WithMaxBodyBytes(64))

	req := httptest.NewRequest(http.MethodPost, "/images/upload", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer signed-token")
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}

func multipartImageBody(t *testing.T, content string) ([]byte, string) {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "avatar.png")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	return body.Bytes(), writer.FormDataContentType()
}
