package postgres

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

// SQLSTATE codes commonly returned by PostgreSQL. Keeping them here avoids
// scattering magic strings throughout repositories.
const (
	CodeStringDataRightTruncation = "22001"
	CodeNotNullViolation          = "23502"
	CodeForeignKeyViolation       = "23503"
	CodeUniqueViolation           = "23505"
	CodeCheckViolation            = "23514"
	CodeSerializationFailure      = "40001"
	CodeDeadlockDetected          = "40P01"
)

type ErrorKind string

const (
	ErrorKindInvalidData ErrorKind = "invalid_data"
	ErrorKindConflict    ErrorKind = "conflict"
	ErrorKindRetryable   ErrorKind = "retryable"
	ErrorKindUnavailable ErrorKind = "unavailable"
	ErrorKindUnknown     ErrorKind = "unknown"
)

// ErrorInfo is a safe, compact summary of a PostgreSQL error for application
// decisions and structured logs. It deliberately excludes Detail, Hint and
// the original message because those fields can contain database internals or
// user-provided values and must not be returned directly to API clients.
type ErrorInfo struct {
	Code       string
	Kind       ErrorKind
	Constraint string
	Retryable  bool
}

// InspectError extracts PostgreSQL metadata even when the PgError is wrapped.
func InspectError(err error) (ErrorInfo, bool) {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return ErrorInfo{}, false
	}

	info := ErrorInfo{
		Code:       pgErr.Code,
		Kind:       classifyErrorCode(pgErr.Code),
		Constraint: pgErr.ConstraintName,
	}
	info.Retryable = info.Kind == ErrorKindRetryable
	return info, true
}

func IsErrorCode(err error, code string) bool {
	info, ok := InspectError(err)
	return ok && info.Code == code
}

func IsUniqueViolation(err error) bool {
	return IsErrorCode(err, CodeUniqueViolation)
}

func classifyErrorCode(code string) ErrorKind {
	switch code {
	case CodeStringDataRightTruncation, CodeNotNullViolation, CodeCheckViolation:
		return ErrorKindInvalidData
	case CodeForeignKeyViolation, CodeUniqueViolation:
		return ErrorKindConflict
	case CodeSerializationFailure, CodeDeadlockDetected:
		return ErrorKindRetryable
	}

	// SQLSTATE class 08 is connection errors; class 53 is insufficient
	// resources; class 57 contains operator intervention/shutdown errors.
	if strings.HasPrefix(code, "08") || strings.HasPrefix(code, "53") || strings.HasPrefix(code, "57") {
		return ErrorKindUnavailable
	}
	return ErrorKindUnknown
}
