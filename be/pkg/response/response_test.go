package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSON(t *testing.T) {
	rec := httptest.NewRecorder()

	JSON(rec, http.StatusCreated, Envelope{
		Success: true,
		Message: "created",
		Data:    map[string]string{"message": "created"},
		Meta: Meta{
			Timestamp: "2026-03-22T11:40:00Z",
		},
	})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected content type application/json, got %q", got)
	}
}

func TestError(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)

	Error(rec, http.StatusBadRequest, "Validation field", req, ErrorDetail{
		Code:    "INVALID_FORMAT",
		Message: "Bad payload",
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var payload Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unexpected json error: %v", err)
	}

	if payload.Message != "Validation field" {
		t.Fatalf("expected message to be set, got %q", payload.Message)
	}

	if len(payload.Errors) != 1 {
		t.Fatalf("expected one error detail, got %d", len(payload.Errors))
	}

	if payload.Errors[0].Code != "INVALID_FORMAT" {
		t.Fatalf("expected error code to be set, got %q", payload.Errors[0].Code)
	}
}
