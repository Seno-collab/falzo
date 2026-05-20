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
	baseURL       string
	publicBaseURL string
	httpClient    *http.Client
}

func NewSeaweedFSStorage(cfg config.UploadConfig) *SeaweedFSStorage {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.SeaweedFSBaseURL), "/")
	publicBaseURL := strings.TrimRight(strings.TrimSpace(cfg.SeaweedFSPublicURL), "/")
	if publicBaseURL == "" {
		publicBaseURL = baseURL
	}

	return &SeaweedFSStorage{
		baseURL:       baseURL,
		publicBaseURL: publicBaseURL,
		httpClient: &http.Client{
			Timeout: cfg.SeaweedFSTimeout,
		},
	}
}

func (s *SeaweedFSStorage) Upload(ctx context.Context, objectKey string, data io.Reader, contentType string) (string, error) {
	if s == nil || s.baseURL == "" || s.httpClient == nil {
		return "", upload.ErrDependencyUnavailable
	}

	targetURL := objectURL(s.baseURL, objectKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, targetURL, data)
	if err != nil {
		return "", fmt.Errorf("%w: create seaweedfs PUT request: %w", upload.ErrStorageFailed, err)
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: seaweedfs PUT %s: %w", upload.ErrStorageFailed, targetURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		detail := strings.TrimSpace(string(body))
		if detail == "" {
			detail = resp.Status
		}
		return "", fmt.Errorf("%w: seaweedfs PUT %s returned %s: %s", upload.ErrStorageFailed, targetURL, resp.Status, detail)
	}

	return objectURL(s.publicBaseURL, objectKey), nil
}

func (s *SeaweedFSStorage) Delete(ctx context.Context, objectKey string) error {
	if s == nil || s.baseURL == "" || s.httpClient == nil {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, objectURL(s.baseURL, objectKey), nil)
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

func objectURL(baseURL string, objectKey string) string {
	parts := strings.Split(strings.Trim(objectKey, "/"), "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return fmt.Sprintf("%s/%s", baseURL, strings.Join(parts, "/"))
}

var _ upload.ImageStorage = (*SeaweedFSStorage)(nil)
