package application

import (
	"context"

	"falzo-be/internal/post/application/command"
	"falzo-be/internal/post/domain"
)

func (s *service) SavePost(ctx context.Context, cmd command.SavePost) error {
	if s.posts == nil {
		return domain.ErrPostDependencyUnavailable
	}
	if cmd.PostID == 0 {
		return ErrPostIDRequired
	}
	if cmd.UserID == 0 {
		return ErrUserIDRequired
	}

	return s.posts.Save(ctx, cmd.PostID, cmd.UserID)
}
