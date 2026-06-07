package itinerary

import (
	"context"
	"errors"
	"testing"
)

type fakeRepository struct {
	filter ListFilter
	detail Detail
	err    error
}

func (f *fakeRepository) ListPublic(_ context.Context, filter ListFilter) (ListPage, error) {
	f.filter = filter
	return ListPage{
		Items: []ListItem{},
		Page:  filter.Page,
		Limit: filter.Limit,
		Total: 0,
	}, f.err
}

func (f *fakeRepository) GetPublicBySlug(context.Context, string) (Detail, error) {
	return f.detail, f.err
}

func TestListPublicNormalizesFilterAndPagination(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)

	page, err := service.ListPublic(t.Context(), ListInput{
		Province:     " Phú Yên ",
		DurationDays: 2,
		BudgetMax:    1500000,
		TravelStyle:  " bien ",
		Page:         2,
		Limit:        12,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if page.Page != 2 || page.Limit != 12 {
		t.Fatalf("expected response page=2 limit=12, got %+v", page)
	}
	if repo.filter.Province != "Phú Yên" || repo.filter.TravelStyle != "bien" {
		t.Fatalf("expected trimmed filters, got %+v", repo.filter)
	}
	if repo.filter.Offset != 12 {
		t.Fatalf("expected offset 12, got %d", repo.filter.Offset)
	}
}

func TestListPublicRejectsLimitTooLarge(t *testing.T) {
	service := NewService(&fakeRepository{})

	_, err := service.ListPublic(t.Context(), ListInput{Limit: maxListLimit + 1})
	if !errors.Is(err, ErrLimitTooLarge) {
		t.Fatalf("expected ErrLimitTooLarge, got %v", err)
	}
}

func TestGetPublicBySlugRequiresSlug(t *testing.T) {
	service := NewService(&fakeRepository{})

	_, err := service.GetPublicBySlug(t.Context(), GetBySlugInput{Slug: " "})
	if !errors.Is(err, ErrSlugRequired) {
		t.Fatalf("expected ErrSlugRequired, got %v", err)
	}
}
