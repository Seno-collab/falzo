package itinerary

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeHandlerService struct {
	listInput ListInput
	listErr   error
	detailErr error
}

func (f *fakeHandlerService) ListPublic(_ context.Context, input ListInput) (ListPage, error) {
	f.listInput = input
	return ListPage{
		Items: []ListItem{},
		Page:  input.Page,
		Limit: input.Limit,
		Total: 0,
	}, f.listErr
}

func (f *fakeHandlerService) GetPublicBySlug(context.Context, GetBySlugInput) (Detail, error) {
	return Detail{}, f.detailErr
}

func TestListParsesQueryFilters(t *testing.T) {
	service := &fakeHandlerService{}
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/?province=Ph%C3%BA+Y%C3%AAn&durationDays=2&budgetMax=1500000&travelStyle=bien&page=2&limit=12", nil)
	rec := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}
	if service.listInput.Province != "Phú Yên" ||
		service.listInput.DurationDays != 2 ||
		service.listInput.BudgetMax != 1500000 ||
		service.listInput.TravelStyle != "bien" ||
		service.listInput.Page != 2 ||
		service.listInput.Limit != 12 {
		t.Fatalf("unexpected parsed input: %+v", service.listInput)
	}
}

func TestListRejectsInvalidPageParam(t *testing.T) {
	handler := NewHandler(&fakeHandlerService{})

	req := httptest.NewRequest(http.MethodGet, "/?page=abc", nil)
	rec := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}

func TestGetBySlugWritesNotFound(t *testing.T) {
	handler := NewHandler(&fakeHandlerService{detailErr: ErrNotFound})

	req := httptest.NewRequest(http.MethodGet, "/missing-itinerary", nil)
	rec := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusNotFound, rec.Code, rec.Body.String())
	}
}
