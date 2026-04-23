package aggregate

import (
	"time"

	"falzo-be/internal/post/domain/entity"
	"falzo-be/internal/post/domain/event"
	"falzo-be/internal/post/domain/valueobject"
)

type Post struct {
	Post         entity.Post
	domainEvents []any
}

func NewPost(
	id uint64,
	userID uint64,
	imageURL valueobject.ImageURL,
	caption valueobject.Caption,
	locationName valueobject.LocationName,
	latitude float64,
	longitude float64,
) *Post {
	now := time.Now().UTC()

	post := &Post{
		Post: entity.Post{
			ID:           id,
			UserID:       userID,
			ImageURL:     imageURL,
			Caption:      caption,
			LocationName: locationName,
			Latitude:     latitude,
			Longitude:    longitude,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}

	post.record(event.PostCreated{
		PostID:        id,
		UserID:        userID,
		ImageURL:      imageURL.String(),
		LocationName:  locationName.String(),
		OccurredAtUTC: now,
	})

	return post
}

func RehydratePost(post entity.Post) *Post {
	return &Post{Post: post}
}

func (p *Post) PullEvents() []any {
	events := append([]any(nil), p.domainEvents...)
	p.domainEvents = nil
	return events
}

func (p *Post) record(evt any) {
	p.domainEvents = append(p.domainEvents, evt)
}
