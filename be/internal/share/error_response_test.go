package share

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestWriteErrorLogsMappedContextAndRootCause(t *testing.T) {
	var output bytes.Buffer
	previous := log.Logger
	t.Cleanup(func() {
		log.Logger = previous
	})
	log.Logger = zerolog.New(&output)

	root := errors.New("dial tcp seaweedfs:8888: connect: connection refused")
	err := fmt.Errorf("%w: seaweedfs PUT failed: %w", errors.New("failed to upload image to storage"), root)

	req := httptest.NewRequest(http.MethodPost, "/api/images/upload", nil)
	rec := httptest.NewRecorder()
	WriteError(rec, req, err, "upload_image", func(error) ApiError {
		return ServiceUnavailable("Image upload unavailable", "Image upload is temporarily unavailable")
	})

	var entry map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(output.Bytes()), &entry); err != nil {
		t.Fatalf("unmarshal log entry: %v", err)
	}

	if entry["operation"] != "upload_image" {
		t.Fatalf("expected operation in log, got %#v", entry["operation"])
	}
	if entry["module"] != "share" {
		t.Fatalf("expected module in log, got %#v", entry["module"])
	}
	if entry["source_file"] != "internal/share/error_response_test.go" {
		t.Fatalf("expected source file in log, got %#v", entry["source_file"])
	}
	if _, ok := entry["source_line"].(float64); !ok {
		t.Fatalf("expected source line in log, got %#v", entry["source_line"])
	}
	if entry["source_func"] != "falzo-be/internal/share.TestWriteErrorLogsMappedContextAndRootCause" {
		t.Fatalf("expected source func in log, got %#v", entry["source_func"])
	}
	if entry["api_code"] != SERVICE_UNAVAILABLE {
		t.Fatalf("expected api code in log, got %#v", entry["api_code"])
	}
	if entry["root_error"] != root.Error() {
		t.Fatalf("expected root cause in log, got %#v", entry["root_error"])
	}
	if got := int(entry["status"].(float64)); got != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, got)
	}
}

func TestWriteErrorLogsAppErrorInternalCause(t *testing.T) {
	var output bytes.Buffer
	previous := log.Logger
	t.Cleanup(func() {
		log.Logger = previous
	})
	log.Logger = zerolog.New(&output)

	publicErr := errors.New("upload dependency unavailable")
	root := errors.New("seaweedfs: connection refused")
	err := NewAppError("STORAGE_UNAVAILABLE", publicErr.Error(), publicErr, root, "images.upload", map[string]string{
		"service": "upload",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/images/upload", nil)
	rec := httptest.NewRecorder()
	WriteError(rec, req, err, "upload_image", func(error) ApiError {
		return ServiceUnavailable("Image upload unavailable", "Image upload is temporarily unavailable")
	})

	var entry map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(output.Bytes()), &entry); err != nil {
		t.Fatalf("unmarshal log entry: %v", err)
	}

	if entry["error"] != publicErr.Error() {
		t.Fatalf("expected safe public error in error field, got %#v", entry["error"])
	}
	if entry["internal_error"] != root.Error() {
		t.Fatalf("expected internal error in log, got %#v", entry["internal_error"])
	}
	if entry["root_error"] != root.Error() {
		t.Fatalf("expected root cause in log, got %#v", entry["root_error"])
	}
	if entry["app_code"] != "STORAGE_UNAVAILABLE" {
		t.Fatalf("expected app code in log, got %#v", entry["app_code"])
	}
	if entry["app_operation"] != "images.upload" {
		t.Fatalf("expected app operation in log, got %#v", entry["app_operation"])
	}
}

func TestMapDBErrorReturnsAppErrorWithPublicErrAndCause(t *testing.T) {
	publicErr := errors.New("auth dependency unavailable")
	internalErr := errors.New("auth internal error")
	cause := errors.New("dial tcp 10.0.0.1:5432: connect: connection refused")

	got := MapDBError(t.Context(), "auth", "accounts.find", cause, publicErr, internalErr)

	if !errors.Is(got, publicErr) {
		t.Fatalf("expected errors.Is to match public error, got %v", got)
	}
	if errors.Is(got, cause) {
		t.Fatalf("expected internal cause to be opaque to errors.Is")
	}

	var appErr *AppError
	if !errors.As(got, &appErr) {
		t.Fatalf("expected AppError, got %T", got)
	}
	if appErr.Internal != cause {
		t.Fatalf("expected internal cause to be preserved, got %v", appErr.Internal)
	}
	if appErr.Code != "DB_DEPENDENCY_UNAVAILABLE" {
		t.Fatalf("expected dependency code, got %q", appErr.Code)
	}
}

func TestModuleFromFunction(t *testing.T) {
	tests := []struct {
		name     string
		function string
		file     string
		want     string
	}{
		{
			name:     "internal handler method",
			function: "falzo-be/internal/post.(*Handler).Create",
			file:     "/Users/home/src/falzo/be/internal/post/handler.go",
			want:     "post",
		},
		{
			name:     "internal package function",
			function: "falzo-be/internal/auth.RequireAuth.func1",
			file:     "/Users/home/src/falzo/be/internal/auth/middleware.go",
			want:     "auth",
		},
		{
			name:     "pkg middleware",
			function: "falzo-be/pkg/http/middleware.Recover.func1",
			file:     "/Users/home/src/falzo/be/pkg/http/middleware/recover.go",
			want:     "http/middleware",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := moduleFromFunction(tt.function, tt.file); got != tt.want {
				t.Fatalf("expected module %q, got %q", tt.want, got)
			}
		})
	}
}
