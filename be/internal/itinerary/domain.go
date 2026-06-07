package itinerary

import (
	"context"
	"errors"
)

const (
	defaultListPage  = 1
	defaultListLimit = 12
	maxListLimit     = 50
)

var (
	ErrDependencyUnavailable = errors.New("itinerary dependency unavailable")
	ErrInternal              = errors.New("itinerary internal error")
	ErrNotFound              = errors.New("itinerary not found")
	ErrSlugRequired          = errors.New("itinerary slug is required")
	ErrPageMustBePositive    = errors.New("page must be greater than 0")
	ErrLimitMustBePositive   = errors.New("limit must be greater than 0")
	ErrLimitTooLarge         = errors.New("limit exceeds maximum")
	ErrInvalidDurationDays   = errors.New("duration days must be between 1 and 14")
	ErrInvalidBudgetMax      = errors.New("budget max must not be negative")
)

type Repository interface {
	ListPublic(ctx context.Context, filter ListFilter) (ListPage, error)
	GetPublicBySlug(ctx context.Context, slug string) (Detail, error)
}

type ListFilter struct {
	Province     string
	DurationDays int
	BudgetMax    int
	TravelStyle  string
	Page         int
	Limit        int
	Offset       int
}

type ListPage struct {
	Items []ListItem `json:"items"`
	Page  int        `json:"page"`
	Limit int        `json:"limit"`
	Total int        `json:"total"`
}

type ListItem struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Slug           string `json:"slug"`
	Province       string `json:"province"`
	DurationDays   int    `json:"durationDays"`
	BudgetMin      int    `json:"budgetMin"`
	BudgetMax      int    `json:"budgetMax"`
	TravelStyle    string `json:"travelStyle"`
	Transportation string `json:"transportation"`
	CoverImageURL  string `json:"coverImageUrl"`
	StopCount      int    `json:"stopCount"`
}

type Detail struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Slug           string `json:"slug"`
	Province       string `json:"province"`
	DurationDays   int    `json:"durationDays"`
	BudgetMin      int    `json:"budgetMin"`
	BudgetMax      int    `json:"budgetMax"`
	TravelStyle    string `json:"travelStyle"`
	Transportation string `json:"transportation"`
	StartLocation  string `json:"startLocation"`
	Description    string `json:"description"`
	CoverImageURL  string `json:"coverImageUrl"`
	Days           []Day  `json:"days"`
}

type Day struct {
	DayNumber int    `json:"dayNumber"`
	Stops     []Stop `json:"stops"`
}

type Stop struct {
	ID            string  `json:"id"`
	LocationID    string  `json:"locationId"`
	LocationName  string  `json:"locationName"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	StartTime     string  `json:"startTime"`
	EndTime       string  `json:"endTime"`
	Note          string  `json:"note"`
	EstimatedCost int     `json:"estimatedCost"`
	StopOrder     int     `json:"stopOrder"`
	DayNumber     int     `json:"-"`
}
