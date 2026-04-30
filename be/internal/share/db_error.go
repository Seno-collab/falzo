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
	return dberr.MapDependencyOrInternal(
		err,
		service,
		operation,
		chimiddleware.GetReqID(ctx),
		dependencyErr,
		internalErr,
	)
}
