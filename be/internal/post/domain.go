package post

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"
)

var (
	ErrNotFound              = errors.New("post not found")
	ErrDependencyUnavailable = errors.New("post dependency unavailable")
	ErrInternal              = errors.New("post internal error")
	ErrUserIDRequired        = errors.New("user id is required")
	ErrPostIDRequired        = errors.New("post id is required")
	ErrPageMustBePositive    = errors.New("page must be greater than 0")
	ErrLimitMustBePositive   = errors.New("limit must be greater than 0")
	ErrLocationNameRequired  = errors.New("location name is required")
	ErrLatitudeOutOfRange    = errors.New("latitude must be between -90 and 90")
	ErrLongitudeOutOfRange   = errors.New("longitude must be between -180 and 180")
	ErrImageURLRequired      = errors.New("image url is required")
	ErrInvalidImageURL       = errors.New("invalid image url")
	ErrCaptionTooLong        = errors.New("caption exceeds max length")
	ErrLocationNameTooLong   = errors.New("location name exceeds max length")
)

type Repository interface {
	Create(ctx context.Context, post *Post) error
	Like(ctx context.Context, postID uint64, userID uint64) error
	Save(ctx context.Context, postID uint64, userID uint64) error
	GetPosts(ctx context.Context, page int, limit int) ([]Post, error)
	GetPostDetail(ctx context.Context, postID uint64) (*Post, error)
	GetPostsByLocation(ctx context.Context, locationName LocationName) ([]Post, error)
}

type Post struct {
	ID           uint64       `json:"id"`
	UserID       uint64       `json:"user_id"`
	ImageURL     ImageURL     `json:"-"`
	Caption      Caption      `json:"-"`
	LocationName LocationName `json:"-"`
	Latitude     float64      `json:"latitude"`
	Longitude    float64      `json:"longitude"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"-"`
}

type PostView struct {
	ID           uint64    `json:"id"`
	UserID       uint64    `json:"user_id"`
	ImageURL     string    `json:"image_url"`
	Caption      string    `json:"caption"`
	LocationName string    `json:"location_name"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	CreatedAt    time.Time `json:"created_at"`
}

func (p Post) View() PostView {
	return PostView{
		ID:           p.ID,
		UserID:       p.UserID,
		ImageURL:     p.ImageURL.String(),
		Caption:      p.Caption.String(),
		LocationName: p.LocationName.String(),
		Latitude:     p.Latitude,
		Longitude:    p.Longitude,
		CreatedAt:    p.CreatedAt,
	}
}

type NewPostInput struct {
	UserID       uint64
	ImageURL     string
	Caption      string
	LocationName string
	Latitude     float64
	Longitude    float64
}

func NewPost(input NewPostInput) (Post, error) {
	if input.UserID == 0 {
		return Post{}, ErrUserIDRequired
	}
	if input.Latitude < -90 || input.Latitude > 90 {
		return Post{}, ErrLatitudeOutOfRange
	}
	if input.Longitude < -180 || input.Longitude > 180 {
		return Post{}, ErrLongitudeOutOfRange
	}

	imageURL, err := NewImageURL(input.ImageURL)
	if err != nil {
		return Post{}, err
	}
	caption, err := NewCaption(input.Caption)
	if err != nil {
		return Post{}, err
	}
	locationName, err := NewLocationName(input.LocationName)
	if err != nil {
		return Post{}, err
	}

	now := time.Now().UTC()
	return Post{
		UserID:       input.UserID,
		ImageURL:     imageURL,
		Caption:      caption,
		LocationName: locationName,
		Latitude:     input.Latitude,
		Longitude:    input.Longitude,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

type ImageURL string

func NewImageURL(raw string) (ImageURL, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", ErrImageURLRequired
	}

	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed == nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", ErrInvalidImageURL
	}

	return ImageURL(value), nil
}

func (i ImageURL) String() string {
	return string(i)
}

type Caption string

func NewCaption(raw string) (Caption, error) {
	value := strings.TrimSpace(raw)
	if len(value) > 2000 {
		return "", ErrCaptionTooLong
	}

	return Caption(value), nil
}

func (c Caption) String() string {
	return string(c)
}

type LocationName string

func NewLocationName(raw string) (LocationName, error) {
	value := strings.TrimSpace(raw)
	if len(value) > 255 {
		return "", ErrLocationNameTooLong
	}

	return LocationName(value), nil
}

func (l LocationName) String() string {
	return string(l)
}
