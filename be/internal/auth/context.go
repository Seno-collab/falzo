package auth

import "context"

type contextKey string

const principalContextKey contextKey = "auth_principal"

func withAuthenticatedUser(ctx context.Context, principal *AuthenticatedUser) context.Context {
	return context.WithValue(ctx, principalContextKey, principal)
}

func authenticatedUserFromContext(ctx context.Context) (*AuthenticatedUser, bool) {
	principal, ok := ctx.Value(principalContextKey).(*AuthenticatedUser)
	return principal, ok
}

func AuthenticatedUserFromContext(ctx context.Context) (*AuthenticatedUser, bool) {
	return authenticatedUserFromContext(ctx)
}
