package post

import (
	"context"
	"strings"
)

type Service struct {
	posts Repository
}

func NewService(posts Repository) *Service {
	return &Service{posts: posts}
}

type CreatePostInput = NewPostInput

type PostActionInput struct {
	PostID uint64
	UserID uint64
}

type ListPostsInput struct {
	Page  int
	Limit int
}

type GetPostDetailInput struct {
	PostID uint64
}

type GetPostsByLocationInput struct {
	LocationName string
}

func (s *Service) CreatePost(ctx context.Context, input CreatePostInput) (PostView, error) {
	if s.posts == nil {
		return PostView{}, ErrDependencyUnavailable
	}

	post, err := NewPost(NewPostInput(input))
	if err != nil {
		return PostView{}, err
	}

	if err := s.posts.Create(ctx, &post); err != nil {
		return PostView{}, err
	}

	return post.View(), nil
}

func (s *Service) LikePost(ctx context.Context, input PostActionInput) error {
	if s.posts == nil {
		return ErrDependencyUnavailable
	}
	if input.PostID == 0 {
		return ErrPostIDRequired
	}
	if input.UserID == 0 {
		return ErrUserIDRequired
	}

	return s.posts.Like(ctx, input.PostID, input.UserID)
}

func (s *Service) SavePost(ctx context.Context, input PostActionInput) error {
	if s.posts == nil {
		return ErrDependencyUnavailable
	}
	if input.PostID == 0 {
		return ErrPostIDRequired
	}
	if input.UserID == 0 {
		return ErrUserIDRequired
	}

	return s.posts.Save(ctx, input.PostID, input.UserID)
}

func (s *Service) GetPosts(ctx context.Context, input ListPostsInput) ([]PostView, error) {
	if s.posts == nil {
		return nil, ErrDependencyUnavailable
	}
	if input.Page <= 0 {
		return nil, ErrPageMustBePositive
	}
	if input.Limit <= 0 {
		return nil, ErrLimitMustBePositive
	}

	items, err := s.posts.GetPosts(ctx, input.Page, input.Limit)
	if err != nil {
		return nil, err
	}

	posts := make([]PostView, 0, len(items))
	for _, item := range items {
		posts = append(posts, item.View())
	}

	return posts, nil
}

func (s *Service) GetPostDetail(ctx context.Context, input GetPostDetailInput) (*PostView, error) {
	if s.posts == nil {
		return nil, ErrDependencyUnavailable
	}
	if input.PostID == 0 {
		return nil, ErrPostIDRequired
	}

	item, err := s.posts.GetPostDetail(ctx, input.PostID)
	if err != nil {
		return nil, err
	}

	mapped := item.View()
	return &mapped, nil
}

func (s *Service) GetPostsByLocation(ctx context.Context, input GetPostsByLocationInput) ([]PostView, error) {
	if s.posts == nil {
		return nil, ErrDependencyUnavailable
	}
	if strings.TrimSpace(input.LocationName) == "" {
		return nil, ErrLocationNameRequired
	}

	locationName, err := NewLocationName(input.LocationName)
	if err != nil {
		return nil, err
	}

	items, err := s.posts.GetPostsByLocation(ctx, locationName)
	if err != nil {
		return nil, err
	}

	posts := make([]PostView, 0, len(items))
	for _, item := range items {
		posts = append(posts, item.View())
	}

	return posts, nil
}
