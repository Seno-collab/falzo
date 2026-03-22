package postgres

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"strings"

	"falzo-be/internal/auth/domain"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog/log"
)

func mapDBError(ctx context.Context, operation string, err error) error {
	if err == nil {
		return nil
	}

	if isDependencyError(err) {
		log.Error().
			Err(err).
			Str("db_operation", operation).
			Str("request_id", chimiddleware.GetReqID(ctx)).
			Msg("auth db dependency unavailable")
		return domain.ErrAuthDependencyUnavailable
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		log.Error().
			Err(err).
			Str("db_operation", operation).
			Str("request_id", chimiddleware.GetReqID(ctx)).
			Str("pg_code", pgErr.Code).
			Msg("auth db query failed")
		return domain.ErrAuthInternal
	}

	log.Error().
		Err(err).
		Str("db_operation", operation).
		Str("request_id", chimiddleware.GetReqID(ctx)).
		Msg("auth db operation failed")
	return domain.ErrAuthInternal
}

func isDependencyError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, sql.ErrConnDone) {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "i/o timeout") ||
		strings.Contains(msg, "dial tcp") ||
		strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "connection reset")
}
