package itinerary

import (
	"context"
	"strings"
)

type Service struct {
	repository Repository
}

type ListInput struct {
	Province     string
	DurationDays int
	BudgetMax    int
	TravelStyle  string
	Page         int
	Limit        int
}

type GetBySlugInput struct {
	Slug string
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) ListPublic(ctx context.Context, input ListInput) (ListPage, error) {
	if s.repository == nil {
		return ListPage{}, ErrDependencyUnavailable
	}

	filter, err := normalizeListInput(input)
	if err != nil {
		return ListPage{}, err
	}

	return s.repository.ListPublic(ctx, filter)
}

func (s *Service) GetPublicBySlug(ctx context.Context, input GetBySlugInput) (Detail, error) {
	if s.repository == nil {
		return Detail{}, ErrDependencyUnavailable
	}

	slug := strings.TrimSpace(input.Slug)
	if slug == "" {
		return Detail{}, ErrSlugRequired
	}

	return s.repository.GetPublicBySlug(ctx, slug)
}

func normalizeListInput(input ListInput) (ListFilter, error) {
	page := input.Page
	if page == 0 {
		page = defaultListPage
	}
	if page < 0 {
		return ListFilter{}, ErrPageMustBePositive
	}

	limit := input.Limit
	if limit == 0 {
		limit = defaultListLimit
	}
	if limit < 0 {
		return ListFilter{}, ErrLimitMustBePositive
	}
	if limit > maxListLimit {
		return ListFilter{}, ErrLimitTooLarge
	}

	if input.DurationDays < 0 || input.DurationDays > 14 {
		return ListFilter{}, ErrInvalidDurationDays
	}
	if input.BudgetMax < 0 {
		return ListFilter{}, ErrInvalidBudgetMax
	}

	return ListFilter{
		Province:     strings.TrimSpace(input.Province),
		DurationDays: input.DurationDays,
		BudgetMax:    input.BudgetMax,
		TravelStyle:  strings.TrimSpace(input.TravelStyle),
		Page:         page,
		Limit:        limit,
		Offset:       (page - 1) * limit,
	}, nil
}
