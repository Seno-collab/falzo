package category

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"falzo-be/internal/auth"
)

type fakeCategoryService struct {
	createErr    error
	getBySlugErr error
}

func (f fakeCategoryService) Create(context.Context, string, string) error {
	return f.createErr
}

func (f fakeCategoryService) GetByID(context.Context, uint64) (Category, error) {
	return Category{}, nil
}

func (f fakeCategoryService) GetBySlug(context.Context, string) (Category, error) {
	return Category{}, f.getBySlugErr
}

func (f fakeCategoryService) List(context.Context) ([]Category, error) {
	return nil, nil
}

func (f fakeCategoryService) Update(context.Context, uint64, string, string) (Category, error) {
	return Category{}, nil
}

func (f fakeCategoryService) Delete(context.Context, uint64) error {
	return nil
}

type fakeCategoryAuthService struct{}

func (fakeCategoryAuthService) Authenticate(context.Context, string) (*auth.AuthenticatedUser, error) {
	return &auth.AuthenticatedUser{UserID: 7, Username: "admin"}, nil
}

func TestCreateCategoryWritesValidationError(t *testing.T) {
	handler := NewHandler(fakeCategoryService{createErr: ErrNameRequired}, fakeCategoryAuthService{})

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"","slug":"travel"}`))
	req.Header.Set("Authorization", "Bearer signed-token")
	rec := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}

func TestGetCategoryBySlugWritesNotFound(t *testing.T) {
	handler := NewHandler(fakeCategoryService{getBySlugErr: ErrNotFound}, fakeCategoryAuthService{})

	req := httptest.NewRequest(http.MethodGet, "/slug/missing", nil)
	rec := httptest.NewRecorder()

	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusNotFound, rec.Code, rec.Body.String())
	}
}

func TestParseIDRejectsTrailingGarbage(t *testing.T) {
	if _, err := parseID("12abc"); err == nil {
		t.Fatal("expected parseID to reject trailing garbage")
	}
}
