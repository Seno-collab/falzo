package middleware

import (
	"net/http"

	httpResponse "falzo-be/pkg/response"

	"github.com/rs/zerolog/log"
)

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Error().
					Interface("panic", recovered).
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Msg("request panicked")
				httpResponse.Error(w, http.StatusInternalServerError, "Internal server error", r, httpResponse.ErrorDetail{
					Code:    "INTERNAL_ERROR",
					Message: "An unexpected error occurred",
				})
			}
		}()

		next.ServeHTTP(w, r)
	})
}
