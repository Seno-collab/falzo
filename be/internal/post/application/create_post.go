package application

import (
	"context"
	"errors"

	"falzo-be/internal/post/application/command"
	"falzo-be/internal/post/application/query"
	"falzo-be/internal/post/domain"
	"falzo-be/internal/post/domain/aggregate"
	"falzo-be/internal/post/domain/value_object"
)

var ErrUserIDRequired = errors.New("user id is required")
var ErrLatitudeOutOfRange = errors.New("latitude must be between -90 and 90")
var ErrLongitudeOutOfRange = errors.New("longitude must be between -180 and 180")

func (s *service) CreatePost(ctx context.Context, cmd command.CreatePost) (query.Post, error) {
	if s.posts == nil {
		return query.Post{}, domain.ErrPostDependencyUnavailable
	}

	if cmd.UserID == 0 {
		return query.Post{}, ErrUserIDRequired
	}

	if cmd.Latitude < -90 || cmd.Latitude > 90 {
		return query.Post{}, ErrLatitudeOutOfRange
	}
	if cmd.Longitude < -180 || cmd.Longitude > 180 {
		return query.Post{}, ErrLongitudeOutOfRange
	}

	imageURL, err := value_object.NewImageURL(cmd.ImageURL)
	if err != nil {
		return query.Post{}, err
	}
	caption, err := value_object.NewCaption(cmd.Caption)
	if err != nil {
		return query.Post{}, err
	}
	locationName, err := value_object.NewLocationName(cmd.LocationName)
	if err != nil {
		return query.Post{}, err
	}

	post := aggregate.NewPost(
		0,
		cmd.UserID,
		imageURL,
		caption,
		locationName,
		cmd.Latitude,
		cmd.Longitude,
	)

	if err := s.posts.Create(ctx, post); err != nil {
		return query.Post{}, err
	}

	return mapPostEntity(post.Post), nil
}
