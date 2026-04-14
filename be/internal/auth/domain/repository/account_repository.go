package repository

import (
	"context"

	"falzo-be/internal/auth/domain/aggregate"
	"falzo-be/internal/auth/domain/valueobject"
)

type AccountRepository interface {
	Save(ctx context.Context, account *aggregate.Account) error
	FindActiveByEmail(ctx context.Context, email valueobject.Email) (*aggregate.Account, error)
}
