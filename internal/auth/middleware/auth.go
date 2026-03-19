package middleware

import (
	"net/http"
	"strings"

	"falzo/internal/auth"
	httpresponse "falzo/pkg/http/response"
)

func RequireAuth(service auth.Service) func(next http.Handler) http.Handler {
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

			next.ServeHTTP(w, r.WithContext(auth.WithClaims(r.Context(), claims)))
		})
	}
}
