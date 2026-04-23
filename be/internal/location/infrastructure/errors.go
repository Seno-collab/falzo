package infrastructure

import (
	"context"

	"falzo-be/internal/location/domain"
	"falzo-be/pkg/dberr"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func mapDBError(ctx context.Context, service, operation string, err error) error {
	return dberr.Handle(err, service, operation, chimiddleware.GetReqID(ctx), func(kind dberr.Kind, _ error) error {
		switch kind {
		case dberr.KindDependency:
			return domain.ErrLocationDependencyUnavailable
		default:
			return domain.ErrLocationInternal
		}
	})
}
