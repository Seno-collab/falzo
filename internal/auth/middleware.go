package auth

import (
	"context"
	"net/http"
	"strings"

	httpresponse "falzo/internal/http/response"
)

type contextKey string

const claimsContextKey contextKey = "auth_claims"

func RequireAuth(service Chi) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				httpresponse.Error(w, http.StatusUnauthorized, "unauthorized", "missing bearer token")
				return
			}

			token, ok := strings.CutPrefix(authHeader, "Bearer ")
			if !ok || token == "" {
				httpresponse.Error(w, http.StatusUnauthorized, "unauthorized", "invalid authorization header")
				return
			}

			claims, err := service.ParseToken(token)
			if err != nil {
				httpresponse.Error(w, http.StatusUnauthorized, "unauthorized", err.Error())
				return
			}

			ctx := context.WithValue(r.Context(), claimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(*Claims)
	return claims, ok
}
