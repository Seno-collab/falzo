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
