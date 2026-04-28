package aggregate

import (
	"time"

	"falzo-be/internal/auth/domain/entity"
	"falzo-be/internal/auth/domain/event"
	"falzo-be/internal/auth/domain/value_object"
)

type Account struct {
	User         entity.User
	Roles        []string
	domainEvents []any
}

func NewAccount(
	id uint64,
	username value_object.Username,
	email value_object.Email,
	passwordHash value_object.PasswordHash,
	roles []string,
) *Account {
	now := time.Now().UTC()

	account := &Account{
		User: entity.User{
			ID:           id,
			Username:     username,
			Email:        email,
			PasswordHash: passwordHash,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		Roles: append([]string(nil), roles...),
	}

	account.record(event.UserRegistered{
		UserID:     id,
		Username:   username.String(),
		Email:      email.String(),
		OccurredAt: now,
	})

	return account
}

func RehydrateAccount(user entity.User, roles []string) *Account {
	return &Account{
		User:  user,
		Roles: append([]string(nil), roles...),
	}
}

func (a *Account) PullEvents() []any {
	events := append([]any(nil), a.domainEvents...)
	a.domainEvents = nil
	return events
}

func (a *Account) record(evt any) {
	a.domainEvents = append(a.domainEvents, evt)
}
