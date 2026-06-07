package location

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeLocationService struct {
	getPlaceBySlugErr error
}

func (f fakeLocationService) Search(context.Context, SearchInput) ([]Location, error) {
	return nil, nil
}

func (f fakeLocationService) Nearby(context.Context, NearbyInput) ([]NearbyLocation, error) {
	return nil, nil
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
