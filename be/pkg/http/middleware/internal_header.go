package middleware

import (
	"crypto/subtle"
	"errors"
	"falzo-be/internal/share"
	"net/http"
	"strings"
)

const defaultInternalHeaderName = "X-Falzo-Internal-Key"

type InternalHeaderConfig struct {
	Required bool
	Name     string
	Value    string
}

func InternalHeader(cfg InternalHeaderConfig) func(http.Handler) http.Handler {
	name := strings.TrimSpace(cfg.Name)
	if name == "" {
		name = defaultInternalHeaderName
	}

	expected := strings.TrimSpace(cfg.Value)

	return func(next http.Handler) http.Handler {
		if !cfg.Required {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actual := r.Header.Get(name)
			if subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) != 1 {
				share.WriteError(w, r, errInvalidInternalHeader, "internal_header", mapInternalHeaderError)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

var errInvalidInternalHeader = errors.New("invalid internal header")

func mapInternalHeaderError(error) share.ApiError {
	return share.Forbidden("Forbidden", "Request is not allowed")
}
