package category

import (
	"context"
	"testing"
)

type fakeCategoryRepository struct {
	updateCalled bool
}

func (f *fakeCategoryRepository) Create(context.Context, CategoryCreateInput) error {
	return nil
}

func (f *fakeCategoryRepository) GetByID(context.Context, uint64) (Category, error) {
	return Category{}, nil
}

func (f *fakeCategoryRepository) GetBySlug(context.Context, string) (Category, error) {
	return Category{}, nil
}

func (f *fakeCategoryRepository) List(context.Context) ([]Category, error) {
	return nil, nil
}

func (f *fakeCategoryRepository) Update(context.Context, uint64, string, string) (Category, error) {
	f.updateCalled = true
	return Category{}, nil
}

func (f *fakeCategoryRepository) Delete(context.Context, uint64) error {
	return nil
}

func TestUpdateCategoryValidatesNameBeforeRepository(t *testing.T) {
	repo := &fakeCategoryRepository{}
	service := NewService(repo)

	_, err := service.Update(t.Context(), 1, "", "travel")
	if err != ErrNameRequired {
		t.Fatalf("expected ErrNameRequired, got %v", err)
	}
	if repo.updateCalled {
		t.Fatal("repository update should not be called for invalid category name")
	}
}
