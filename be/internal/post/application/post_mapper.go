package application

import (
	"falzo-be/internal/post/application/query"
	"falzo-be/internal/post/domain/entity"
)

func mapPostEntity(post entity.Post) query.Post {
	return query.Post{
		ID:           post.ID,
		UserID:       post.UserID,
		ImageURL:     post.ImageURL.String(),
		Caption:      post.Caption.String(),
		LocationName: post.LocationName.String(),
		Latitude:     post.Latitude,
		Longitude:    post.Longitude,
		CreatedAt:    post.CreatedAt,
	}
}
