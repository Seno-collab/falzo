package auth

import (
	"errors"
	"net/http"

	"falzo-be/internal/share"
)

var (
	errMissingBearerToken           = errors.New("missing bearer token")
	errInvalidAuthorizationHeader   = errors.New("invalid authorization header")
	errInvalidJSONPayload           = errors.New("invalid JSON payload")
	errRegisterFieldsRequired       = errors.New("register fields required")
	errLoginFieldsRequired          = errors.New("login fields required")
	errInvalidEmailField            = errors.New("invalid email field")
	errRegisterRateLimited          = errors.New("register rate limited")
	errLoginRateLimited             = errors.New("login rate limited")
	errRefreshRateLimited           = errors.New("refresh rate limited")
	errRefreshTokenRequired         = errors.New("refresh token required")
	errMissingAuthContext           = errors.New("missing auth context")
	errChangePasswordFieldsRequired = errors.New("change password fields required")
)

func mapAuthError(err error) share.ApiError {
	switch {
	case errors.Is(err, errMissingBearerToken):
		return share.UnauthorizedCredentials(share.Unauthorized, "Missing bearer token")
	case errors.Is(err, errInvalidAuthorizationHeader):
		return share.UnauthorizedCredentials(share.Unauthorized, "Invalid authorization header")
	case errors.Is(err, errInvalidJSONPayload):
		return share.ApiError{
			Status:  http.StatusBadRequest,
			Message: share.ValidationField,
			Code:    share.INVALID_FORMAT,
			Detail:  "Invalid JSON payload",
		}
	case errors.Is(err, errRegisterFieldsRequired):
		return share.Required("", "user_name, email and password are required")
	case errors.Is(err, errLoginFieldsRequired):
		return share.Required("", "email and password are required")
	case errors.Is(err, errInvalidEmailField):
		return share.BadRequest("email", "email must be a valid email")
	case errors.Is(err, errRegisterRateLimited):
		return share.TooManyRequests("Too many registration attempts, please try again later")
	case errors.Is(err, errLoginRateLimited):
		return share.TooManyRequests("Too many login attempts, please try again later")
	case errors.Is(err, errRefreshRateLimited):
		return share.TooManyRequests("Too many refresh attempts, please try again later")
	case errors.Is(err, errRefreshTokenRequired):
		return share.Required("refresh_token", "Refresh token is required")
	case errors.Is(err, errMissingAuthContext):
		return share.UnauthorizedCredentials(share.Unauthorized, "Missing auth context")
	case errors.Is(err, errChangePasswordFieldsRequired):
		return share.Required("", "current_password and new_password are required")
	case errors.Is(err, ErrInvalidCredentials):
		return share.UnauthorizedCredentials("Invalid credentials", "Email or password is incorrect")
	case errors.Is(err, ErrInvalidToken):
		return share.UnauthorizedCredentials(share.Unauthorized, "Token is invalid")
	case errors.Is(err, ErrSessionRevoked):
		return share.UnauthorizedCredentials(share.Unauthorized, "Session has been revoked or expired")
	case errors.Is(err, ErrUserExists):
		return share.ApiError{
			Status:  http.StatusConflict,
			Message: "Account already exists",
			Code:    "ALREADY_EXISTS",
			Detail:  "Username or email is already in use",
		}
	case errors.Is(err, ErrInvalidPassword):
		return share.BadRequest("new_password", "new_password must be at least 8 characters and contain letters and digits")
	case errors.Is(err, ErrDependencyUnavailable) || errors.Is(err, ErrTemporarilyUnavailable):
		return share.ServiceUnavailable("Authentication service unavailable", "Authentication service is temporarily unavailable")
	default:
		return share.Internal()
	}
}
