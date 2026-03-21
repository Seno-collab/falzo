package application

import (
	"context"
	"falzo-be/internal/auth/application/command"
)

func (s *service) Logout(ctx context.Context, cmd command.Logout) error {
	return nil
}
