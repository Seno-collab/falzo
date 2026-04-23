package infrastructure

import (
	"context"

	"falzo-be/internal/location/domain"
	"falzo-be/pkg/dberr"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func mapDBError(ctx context.Context, service, operation string, err error) error {
	return dberr.MapDependencyOrInternal(
		err,
		service,
		operation,
		chimiddleware.GetReqID(ctx),
		domain.ErrLocationDependencyUnavailable,
		domain.ErrLocationInternal,
	)
}
