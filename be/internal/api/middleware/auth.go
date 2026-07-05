package middleware

import (
	"be/internal/api/http/response"
	"be/internal/application/ports"
	"be/internal/shared/apperror"
	"context"
	"net/http"
	"strings"
	"time"
)

type Principal struct {
	UserID   int64
	UserName string
}
type principalContextKey struct{}
type Authenticator struct{ tokens ports.TokenManager }

func NewAuthenticator(tokens ports.TokenManager) *Authenticator {
	return &Authenticator{tokens: tokens}
}

func (a *Authenticator) RequireAccessToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Fields(r.Header.Get("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Error(w, apperror.Unauthorized("Bearer token is required"))
			return
		}
		claims, err := a.tokens.ParseAccess(parts[1], time.Now().UTC())
		if err != nil {
			response.Error(w, apperror.InvalidToken())
			return
		}
		ctx := context.WithValue(r.Context(), principalContextKey{}, Principal{UserID: claims.UserID, UserName: claims.UserName})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalContextKey{}).(Principal)
	return p, ok
}
