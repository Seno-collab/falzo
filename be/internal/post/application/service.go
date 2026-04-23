package application

import (
	"context"

	"falzo-be/internal/post/application/command"
	"falzo-be/internal/post/application/query"
	"falzo-be/internal/post/domain/repository"
)

type Service interface {
	CreatePost(ctx context.Context, cmd command.CreatePost) (query.Post, error)
	LikePost(ctx context.Context, cmd command.LikePost) error
	SavePost(ctx context.Context, cmd command.SavePost) error
	GetPosts(ctx context.Context, input query.GetPosts) ([]query.Post, error)
	GetPostDetail(ctx context.Context, input query.GetPostDetail) (*query.Post, error)
	GetPostsByLocation(ctx context.Context, input query.GetPostsByLocation) ([]query.Post, error)
}

type service struct {
	posts repository.PostRepository
}

func New(posts repository.PostRepository) Service {
	return &service{
		posts: posts,
	}
}
