package infra

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"falzo-be/internal/upload"
	"falzo-be/pkg/config"
)

func TestSeaweedFSStorageUploadsToInternalURLAndReturnsPublicURL(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		gotPath = r.URL.EscapedPath()
		w.WriteHeader(http.StatusCreated)
	}))
	t.Cleanup(server.Close)

	storage := NewSeaweedFSStorage(config.UploadConfig{
		SeaweedFSBaseURL:   server.URL,
		SeaweedFSPublicURL: "https://cdn.example.com/assets",
		SeaweedFSTimeout:   time.Second,
	})

	gotURL, err := storage.Upload(t.Context(), "7/avatar one.png", strings.NewReader("image"), "image/png")
	if err != nil {
		t.Fatalf("upload: %v", err)
	}

	if gotPath != "/7/avatar%20one.png" {
		t.Fatalf("expected upload through internal seaweedfs path, got %q", gotPath)
	}
	if gotURL != "https://cdn.example.com/assets/7/avatar%20one.png" {
		t.Fatalf("expected public URL, got %q", gotURL)
	}
}

func TestSeaweedFSStorageWrapsFailedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "no volume", http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)

	storage := NewSeaweedFSStorage(config.UploadConfig{
		SeaweedFSBaseURL: server.URL,
		SeaweedFSTimeout: time.Second,
	})

	_, err := storage.Upload(t.Context(), "7/avatar.png", strings.NewReader("image"), "image/png")
	if !errors.Is(err, upload.ErrStorageFailed) {
		t.Fatalf("expected ErrStorageFailed, got %v", err)
	}
}
