package infra

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"falzo-be/internal/upload"
	"falzo-be/pkg/config"
)

type SeaweedFSStorage struct {
	baseURL    string
	httpClient *http.Client
}

func NewSeaweedFSStorage(cfg config.UploadConfig) *SeaweedFSStorage {
	return &SeaweedFSStorage{
		baseURL: strings.TrimRight(strings.TrimSpace(cfg.SeaweedFSBaseURL), "/"),
		httpClient: &http.Client{
			Timeout: cfg.SeaweedFSTimeout,
		},
	}
}

func (s *SeaweedFSStorage) Upload(ctx context.Context, objectKey string, data io.Reader, contentType string) (string, error) {
	if s == nil || s.baseURL == "" || s.httpClient == nil {
		return "", upload.ErrDependencyUnavailable
	}

	targetURL := s.objectURL(objectKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, targetURL, data)
	if err != nil {
		return "", upload.ErrStorageFailed
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", upload.ErrStorageFailed
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", upload.ErrStorageFailed
	}

	return targetURL, nil
}

func (s *SeaweedFSStorage) Delete(ctx context.Context, objectKey string) error {
	if s == nil || s.baseURL == "" || s.httpClient == nil {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, s.objectURL(objectKey), nil)
	if err != nil {
		return nil
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	return nil
}

func (s *SeaweedFSStorage) objectURL(objectKey string) string {
	parts := strings.Split(strings.Trim(objectKey, "/"), "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return fmt.Sprintf("%s/%s", s.baseURL, strings.Join(parts, "/"))
}

var _ upload.ImageStorage = (*SeaweedFSStorage)(nil)
