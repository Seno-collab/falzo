package httpapi

import (
	"net/http"
	"strings"

	"falzo-be/internal/auth/application"
	httpResponse "falzo-be/pkg/response"
)

func requireAuth(service application.Service) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				httpResponse.Error(w, http.StatusUnauthorized, "Unauthorized", r, httpResponse.ErrorDetail{
					Code:    "UNAUTHORIZED",
					Message: "Missing bearer token",
				})
				return
			}

			token, ok := strings.CutPrefix(authHeader, "Bearer ")
			if !ok || token == "" {
				httpResponse.Error(w, http.StatusUnauthorized, "Unauthorized", r, httpResponse.ErrorDetail{
					Code:    "UNAUTHORIZED",
					Message: "Invalid authorization header",
				})
				return
			}

			principal, err := service.Authenticate(r.Context(), token)
			if err != nil {
				writeAuthError(w, r, err, "authenticate")
				return
			}

			next.ServeHTTP(w, r.WithContext(withAuthenticatedUser(r.Context(), principal)))
		})
	}
}
