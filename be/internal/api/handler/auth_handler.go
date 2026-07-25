package handler

import (
	"be/internal/api/http/request"
	"be/internal/api/http/response"
	authapp "be/internal/application/auth"
	domainuser "be/internal/domain/user"
	"be/internal/shared/apperror"
	"errors"
	"log/slog"
	"net/http"
)

type AuthHandler struct {
	login            *authapp.LoginUseCase
	refresh          *authapp.RefreshTokenUseCase
	forgotPassword   *authapp.ForgotPasswordUseCase
	resetPassword    *authapp.ResetPasswordUseCase
	logout           *authapp.LogoutUseCase
	googleLogin      *authapp.GoogleLoginUseCase
	exposeResetToken bool
	logger           *slog.Logger
}

func NewAuthHandler(
	login *authapp.LoginUseCase,
	refresh *authapp.RefreshTokenUseCase,
	forgot *authapp.ForgotPasswordUseCase,
	reset *authapp.ResetPasswordUseCase,
	logout *authapp.LogoutUseCase,
	exposeResetToken bool,
	logger *slog.Logger,
	googleLogin *authapp.GoogleLoginUseCase,
) *AuthHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &AuthHandler{
		login:            login,
		refresh:          refresh,
		forgotPassword:   forgot,
		resetPassword:    reset,
		logout:           logout,
		googleLogin:      googleLogin,
		exposeResetToken: exposeResetToken,
		logger:           logger,
	}
}

type loginRequest struct {
	UserName string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}
type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
type forgotPasswordRequest struct {
	UserName string `json:"username" validate:"required"`
}
type resetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=72"`
}
type googleLoginRequest struct {
	Credential string `json:"credential" validate:"required,max=8192"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	body, err := request.DecodeJSON[loginRequest](w, r)
	if err != nil {
		response.Error(w, err)
		return
	}
	result, err := h.login.Execute(r.Context(), authapp.LoginInput{UserName: body.UserName, Password: body.Password})
	if err != nil {
		h.writeAuthError(w, r, "login", err)
		return
	}
	response.OK(w, result)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	body, err := request.DecodeJSON[refreshRequest](w, r)
	if err != nil {
		response.Error(w, err)
		return
	}
	result, err := h.refresh.Execute(r.Context(), authapp.RefreshTokenInput{RefreshToken: body.RefreshToken})
	if err != nil {
		h.writeAuthError(w, r, "refresh", err)
		return
	}
	response.OK(w, result)
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	body, err := request.DecodeJSON[forgotPasswordRequest](w, r)
	if err != nil {
		response.Error(w, err)
		return
	}
	result, err := h.forgotPassword.Execute(r.Context(), authapp.ForgotPasswordInput{UserName: body.UserName})
	if err != nil {
		h.writeAuthError(w, r, "forgot_password", err)
		return
	}
	data := map[string]any{"accepted": true}
	if h.exposeResetToken && result.ResetToken != "" {
		data["reset_token"] = result.ResetToken
	}
	response.OK(w, data)
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	body, err := request.DecodeJSON[resetPasswordRequest](w, r)
	if err != nil {
		response.Error(w, err)
		return
	}
	err = h.resetPassword.Execute(r.Context(), authapp.ResetPasswordInput{Token: body.Token, NewPassword: body.NewPassword})
	if err != nil {
		h.writeAuthError(w, r, "reset_password", err)
		return
	}
	response.NoContent(w)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	body, err := request.DecodeJSON[refreshRequest](w, r)
	if err != nil {
		response.Error(w, err)
		return
	}
	if err := h.logout.Execute(r.Context(), body.RefreshToken); err != nil {
		h.writeAuthError(w, r, "logout", err)
		return
	}
	response.NoContent(w)
}

func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	body, err := request.DecodeJSON[googleLoginRequest](w, r)
	if err != nil {
		response.Error(w, err)
		return
	}
	result, err := h.googleLogin.Execute(r.Context(), authapp.GoogleLoginInput{Credential: body.Credential})
	if err != nil {
		h.writeAuthError(w, r, "google_login", err)
		return
	}
	response.OK(w, result)
}

func (h *AuthHandler) writeAuthError(w http.ResponseWriter, r *http.Request, operation string, err error) {
	mappedErr := mapAuthError(err)
	appErr := apperror.FromError(mappedErr)
	level := slog.LevelWarn
	if appErr.Code == apperror.CodeInternalServerError {
		level = slog.LevelError
	}
	h.logger.LogAttrs(r.Context(), level, "auth operation failed",
		slog.String("operation", operation),
		slog.String("code", string(appErr.Code)),
		slog.Any("error", err),
	)
	response.Error(w, mappedErr)
}

func mapAuthError(err error) error {
	switch {
	case errors.Is(err, domainuser.ErrInvalidUsernameOrPassword):
		return apperror.InvalidCredentials()
	case errors.Is(err, domainuser.ErrUserLocked):
		return apperror.UserLocked()
	case errors.Is(err, domainuser.ErrUserDisabled):
		return apperror.UserDisabled()
	case errors.Is(err, domainuser.ErrUserNameAlreadyExists):
		return apperror.UserNameExists()
	case errors.Is(err, domainuser.ErrInvalidToken):
		return apperror.InvalidToken()
	case errors.Is(err, authapp.ErrInvalidGoogleCredential):
		return apperror.InvalidCredentials()
	default:
		return apperror.Internal(err)
	}
}
