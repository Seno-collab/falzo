package httpapi

import (
	"context"

	"falzo-be/internal/auth/application/query"
)

type contextKey string

const principalContextKey contextKey = "auth_principal"

func withAuthenticatedUser(ctx context.Context, principal *query.AuthenticatedUser) context.Context {
	return context.WithValue(ctx, principalContextKey, principal)
}

func authenticatedUserFromContext(ctx context.Context) (*query.AuthenticatedUser, bool) {
	principal, ok := ctx.Value(principalContextKey).(*query.AuthenticatedUser)
	return principal, ok
}
