package httpauth

import (
	"net/http"
	"strings"

	"falzo-be/internal/auth/application"
	httpresponse "falzo-be/pkg/response"
)

func RequireAuth(service application.Service) func(next http.Handler) http.Handler {
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

			principal, err := service.Authenticate(r.Context(), token)
			if err != nil {
				httpresponse.Error(w, http.StatusUnauthorized, "unauthorized", err.Error())
				return
			}

			next.ServeHTTP(w, r.WithContext(WithAuthenticatedUser(r.Context(), principal)))
		})
	}
}
