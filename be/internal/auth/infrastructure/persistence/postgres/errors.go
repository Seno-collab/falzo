package postgres

import (
	"context"

	"falzo-be/internal/auth/domain"
	"falzo-be/pkg/dberr"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func mapDBError(ctx context.Context, operation string, err error) error {
	return dberr.Handle(err, "auth", operation, chimiddleware.GetReqID(ctx), func(kind dberr.Kind, _ error) error {
		switch kind {
		case dberr.KindDependency:
			return domain.ErrAuthDependencyUnavailable
		default:
			return domain.ErrAuthInternal
		}
	})
}
