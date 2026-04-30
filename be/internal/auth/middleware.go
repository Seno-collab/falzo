package auth

import (
	"context"
	"net/http"
	"strings"

	"falzo-be/internal/share"
)

type authenticationService interface {
	Authenticate(ctx context.Context, rawToken string) (*AuthenticatedUser, error)
}

func RequireAuth(service authenticationService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				share.WriteError(w, r, errMissingBearerToken, "authenticate", mapAuthError)
				return
			}

			token, ok := strings.CutPrefix(authHeader, "Bearer ")
			if !ok || token == "" {
				share.WriteError(w, r, errInvalidAuthorizationHeader, "authenticate", mapAuthError)
				return
			}

			principal, err := service.Authenticate(r.Context(), token)
			if err != nil {
				share.WriteError(w, r, err, "authenticate", mapAuthError)
				return
			}

			next.ServeHTTP(w, r.WithContext(withAuthenticatedUser(r.Context(), principal)))
		})
	}
}
