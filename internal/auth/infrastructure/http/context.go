package httpauth

import (
	"context"

	"falzo/internal/auth/application/query"
)

type contextKey string

const principalContextKey contextKey = "auth_principal"

func WithAuthenticatedUser(ctx context.Context, principal *query.AuthenticatedUser) context.Context {
	return context.WithValue(ctx, principalContextKey, principal)
}

func AuthenticatedUserFromContext(ctx context.Context) (*query.AuthenticatedUser, bool) {
	principal, ok := ctx.Value(principalContextKey).(*query.AuthenticatedUser)
	return principal, ok
}
