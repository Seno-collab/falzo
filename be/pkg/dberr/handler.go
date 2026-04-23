package dberr

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog/log"
)

type Kind string

const (
	KindDependency Kind = "dependency"
	KindUnique     Kind = "unique_violation"
	KindForeignKey Kind = "foreign_key_violation"
	KindInternal   Kind = "internal"
)

type Mapper func(kind Kind, err error) error

func Handle(err error, module string, operation string, requestID string, mapFn Mapper) error {
	if err == nil {
		return nil
	}

	kind, pgCode := Classify(err)

	entry := log.Error().
		Err(err).
		Str("module", module).
		Str("db_operation", operation).
		Str("db_error_kind", string(kind))
	if requestID != "" {
		entry = entry.Str("request_id", requestID)
	}
	if pgCode != "" {
		entry = entry.Str("pg_code", pgCode)
	}
	entry.Msg("database operation failed")

	if mapFn == nil {
		return err
	}

	return mapFn(kind, err)
}

func MapDependencyOrInternal(
	err error,
	module string,
	operation string,
	requestID string,
	dependencyErr error,
	internalErr error,
) error {
	return Handle(err, module, operation, requestID, func(kind Kind, _ error) error {
		if kind == KindDependency {
			if dependencyErr != nil {
				return dependencyErr
			}
			return err
		}

		if internalErr != nil {
			return internalErr
		}

		return err
	})
}

func Classify(err error) (Kind, string) {
	if err == nil {
		return KindInternal, ""
	}

	if isDependencyError(err) {
		return KindDependency, ""
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return KindUnique, pgErr.Code
		case "23503":
			return KindForeignKey, pgErr.Code
		default:
			return KindInternal, pgErr.Code
		}
	}

	return KindInternal, ""
}

func IsUniqueViolation(err error) bool {
	kind, _ := Classify(err)
	return kind == KindUnique
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
