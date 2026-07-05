package apperror

import (
	"errors"
	"net/http"
)

type AppError struct {
	Code       Code
	Message    string
	HTTPStatus int
	Err        error
	Details    any
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return string(e.Code) + ": " + e.Err.Error()
	}
	return string(e.Code) + ": " + e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}
func New(code Code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

func Wrap(code Code, message string, httpStatus int, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Err:        err,
	}
}

func WithDetails(code Code, message string, httpStatus int, details any) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Details:    details,
	}
}

func FromError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	return New(
		CodeInternalServerError,
		"Internal server error",
		http.StatusInternalServerError,
	)
}

func IsCode(err error, code Code) bool {
	if appErr, ok := errors.AsType[*AppError](err); ok {
		return appErr.Code == code
	}
	return false
}

func InvalidRequest(message string) *AppError {
	return New(CodeInvalidRequest, message, http.StatusBadRequest)
}

func ValidationError(details any) *AppError {
	return WithDetails(CodeValidationError, "Validation failed", http.StatusBadRequest, details)
}

func Unauthorized(message string) *AppError {
	return New(CodeUnauthorized, message, http.StatusUnauthorized)
}

func Forbidden(message string) *AppError {
	return New(CodeForbidden, message, http.StatusForbidden)
}

func NotFound(message string) *AppError {
	return New(CodeNotFound, message, http.StatusNotFound)
}

func Conflict(message string) *AppError {
	return New(CodeConflict, message, http.StatusConflict)
}

func Internal(err error) *AppError {
	return Wrap(CodeInternalServerError, "Internal server error", http.StatusInternalServerError, err)
}

func InvalidCredentials() *AppError {
	return New(CodeInvalidCredentials, "Invalid username or password", http.StatusUnauthorized)
}

func UserLocked() *AppError {
	return New(CodeUserLocked, "User is locked", http.StatusLocked)
}

func UserDisabled() *AppError {
	return New(CodeUserDisabled, "User is disabled", http.StatusForbidden)
}

func UserNameExists() *AppError {
	return New(CodeUserNameExists, "Username already exists", http.StatusConflict)
}

func InvalidToken() *AppError {
	return New(CodeInvalidToken, "Invalid or expired token", http.StatusUnauthorized)
}

func WalletNotEnoughBalance() *AppError {
	return New(CodeWalletNotEnoughBalance, "Wallet balance is not enough", http.StatusBadRequest)
}
