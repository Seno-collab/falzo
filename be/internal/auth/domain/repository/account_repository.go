package repository

import (
	"context"

	"falzo-be/internal/auth/domain/aggregate"
	"falzo-be/internal/auth/domain/value_object"
)

type AccountRepository interface {
	Save(ctx context.Context, account *aggregate.Account) error
	FindActiveByEmail(ctx context.Context, email value_object.Email) (*aggregate.Account, error)
}
