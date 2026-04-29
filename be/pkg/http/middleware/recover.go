package middleware

import (
	"errors"
	"net/http"

	"falzo-be/internal/share"

	"github.com/rs/zerolog/log"
)

var errRequestPanic = errors.New("request panicked")

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Error().
					Interface("panic", recovered).
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Msg("request panicked")
				share.WriteError(w, r, errRequestPanic, "recover", mapRecoverError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func mapRecoverError(err error) share.ApiError {
	return share.Internal()
}
