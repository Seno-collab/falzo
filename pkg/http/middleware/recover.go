package middleware

import (
	"fmt"
	"net/http"

	httpresponse "falzo/pkg/http/response"

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
				httpresponse.Error(w, http.StatusInternalServerError, "internal server error", fmt.Sprintf("%v", recovered))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
