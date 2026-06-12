package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"falzo-be/internal/share"
)

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				share.WriteError(w, r, panicAppError(recovered), "recover", mapRecoverError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func panicAppError(recovered any) error {
	internalErr, ok := recovered.(error)
	if !ok {
		internalErr = errors.New(fmt.Sprint(recovered))
	}

	return share.NewAppError(
		"REQUEST_PANIC",
		errRequestPanic.Error(),
		errRequestPanic,
		internalErr,
		"recover",
		map[string]string{
			"panic":      fmt.Sprint(recovered),
			"panic_type": fmt.Sprintf("%T", recovered),
			"stack":      string(debug.Stack()),
		},
	)
}
