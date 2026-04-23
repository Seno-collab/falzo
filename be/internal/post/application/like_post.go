package application

import (
	"context"
	"errors"

	"falzo-be/internal/post/application/command"
	"falzo-be/internal/post/domain"
)

var ErrPostIDRequired = errors.New("post id is required")

func (s *service) LikePost(ctx context.Context, cmd command.LikePost) error {
	if s.posts == nil {
		return domain.ErrPostDependencyUnavailable
	}
	if cmd.PostID == 0 {
		return ErrPostIDRequired
	}
	if cmd.UserID == 0 {
		return ErrUserIDRequired
	}

	return s.posts.Like(ctx, cmd.PostID, cmd.UserID)
}
