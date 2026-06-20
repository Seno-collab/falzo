package location

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type fakeLocationService struct {
	getPlaceBySlugErr error
	nearbyErr         error
}

func (f fakeLocationService) Search(context.Context, SearchInput) ([]Location, error) {
	return nil, nil
}

func (f fakeLocationService) Nearby(context.Context, NearbyInput) ([]NearbyLocation, error) {
	return nil, f.nearbyErr
}

func (f fakeLocationService) GetPostsByLocation(context.Context, GetPostsByLocationInput) ([]LocationPost, error) {
	return nil, nil
}

func (f fakeLocationService) GetPlaceBySlug(context.Context, GetPlaceBySlugInput) (PlaceDetail, error) {
	return PlaceDetail{}, f.getPlaceBySlugErr
}

func TestGetPlaceBySlugWritesNotFound(t *testing.T) {
	handler := NewHandler(fakeLocationService{getPlaceBySlugErr: ErrPlaceNotFound})

	req := httptest.NewRequest(http.MethodGet, "/missing-place", nil)
	rec := httptest.NewRecorder()

	handler.PlaceRoutes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusNotFound, rec.Code, rec.Body.String())
	}
}

func TestNearbyRadiusErrorLogsContext(t *testing.T) {
	var output bytes.Buffer
	previous := log.Logger
	t.Cleanup(func() {
		log.Logger = previous
	})
	log.Logger = zerolog.New(&output)

	handler := NewHandler(fakeLocationService{nearbyErr: ErrRadiusMustBePositive})

	req := httptest.NewRequest(http.MethodGet, "/nearby?lat=10.75&lng=106.67&radius=0", nil)
	rec := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}

	var entry map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(output.Bytes()), &entry); err != nil {
		t.Fatalf("unmarshal log entry: %v", err)
	}

	if entry["operation"] != "nearby_location" {
		t.Fatalf("expected operation nearby_location, got %#v", entry["operation"])
	}
	if entry["app_operation"] != "locations.nearby.validate_radius" {
		t.Fatalf("expected app operation in log, got %#v", entry["app_operation"])
	}
	if _, ok := entry["source_line"].(float64); !ok {
		t.Fatalf("expected source line in log, got %#v", entry["source_line"])
	}

	metadata, ok := entry["app_metadata"].(map[string]any)
	if !ok {
		t.Fatalf("expected app metadata in log, got %#v", entry["app_metadata"])
	}
	if metadata["radius_meters"] != "0" {
		t.Fatalf("expected radius metadata, got %#v", metadata["radius_meters"])
	}
	if metadata["lat"] != "10.75" {
		t.Fatalf("expected lat metadata, got %#v", metadata["lat"])
	}
	if metadata["lng"] != "106.67" {
		t.Fatalf("expected lng metadata, got %#v", metadata["lng"])
	}
}

func TestNearbyRadiusErrorCanReturnDebugContext(t *testing.T) {
	t.Setenv("ERROR_RESPONSE_DEBUG_ENABLED", "true")

	handler := NewHandler(fakeLocationService{nearbyErr: ErrRadiusMustBePositive})

	req := httptest.NewRequest(http.MethodGet, "/nearby?lat=10.75&lng=106.67&radius=0", nil)
	rec := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}

	var payload struct {
		Errors []struct {
			Code  string `json:"code"`
			Field string `json:"field"`
			Debug *struct {
				Operation    string            `json:"operation"`
				SourceFile   string            `json:"source_file"`
				SourceLine   int               `json:"source_line"`
				AppCode      string            `json:"app_code"`
				AppOperation string            `json:"app_operation"`
				AppMetadata  map[string]string `json:"app_metadata"`
				RootError    string            `json:"root_error"`
			} `json:"debug"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(payload.Errors) != 1 {
		t.Fatalf("expected one error, got %d", len(payload.Errors))
	}

	item := payload.Errors[0]
	if item.Code != "INVALID_FIELD" {
		t.Fatalf("expected invalid field code, got %q", item.Code)
	}
	if item.Field != "radius" {
		t.Fatalf("expected radius field, got %q", item.Field)
	}
	if item.Debug == nil {
		t.Fatal("expected debug context")
	}
	if item.Debug.Operation != "nearby_location" {
		t.Fatalf("expected debug operation, got %q", item.Debug.Operation)
	}
	if item.Debug.SourceFile == "" || item.Debug.SourceLine == 0 {
		t.Fatalf("expected source location, got file=%q line=%d", item.Debug.SourceFile, item.Debug.SourceLine)
	}
	if item.Debug.AppCode != "LOCATION_RADIUS_INVALID" {
		t.Fatalf("expected app code, got %q", item.Debug.AppCode)
	}
	if item.Debug.AppOperation != "locations.nearby.validate_radius" {
		t.Fatalf("expected app operation, got %q", item.Debug.AppOperation)
	}
	if item.Debug.RootError != ErrRadiusMustBePositive.Error() {
		t.Fatalf("expected root error, got %q", item.Debug.RootError)
	}
	if item.Debug.AppMetadata["radius_meters"] != "0" {
		t.Fatalf("expected radius metadata, got %#v", item.Debug.AppMetadata)
	}
}
