package share

import (
	"context"

	"falzo-be/pkg/dberr"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func MapDBError(
	ctx context.Context,
	service string,
	operation string,
	err error,
	dependencyErr error,
	internalErr error,
) error {
	return dberr.Handle(err, service, operation, chimiddleware.GetReqID(ctx), func(kind dberr.Kind, cause error) error {
		publicErr := internalErr
		code := "DB_INTERNAL_ERROR"
		if kind == dberr.KindDependency {
			publicErr = dependencyErr
			code = "DB_DEPENDENCY_UNAVAILABLE"
		}
		if publicErr == nil {
			return cause
		}

		return NewAppError(code, publicErr.Error(), publicErr, cause, operation, map[string]string{
			"service":       service,
			"db_error_kind": string(kind),
		})
	})
}
