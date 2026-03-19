package auth

import "context"

type Service interface {
	Login(ctx context.Context, username string, password string) (string, error)
	Logout(ctx context.Context, token string) error
}

type NoopService struct{}

func (NoopService) Login(ctx context.Context, username string, password string) (string, error) {
	return "", nil
}

func (NoopService) Logout(ctx context.Context, token string) error {
	return nil
}
