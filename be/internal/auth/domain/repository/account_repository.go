package repository

import (
	"context"

	"falzo-be/internal/auth/domain/aggregate"
	"falzo-be/internal/auth/domain/valueobject"
)

type AccountRepository interface {
	Save(ctx context.Context, account *aggregate.Account) error
	FindActiveByUsername(ctx context.Context, username valueobject.Username) (*aggregate.Account, error)
}
