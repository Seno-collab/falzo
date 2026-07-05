package handler

import (
	"be/internal/api/http/request"
	"be/internal/api/http/response"
	authapp "be/internal/application/auth"
	domainuser "be/internal/domain/user"
	"be/internal/shared/apperror"
	"errors"
	"net/http"
)

type AuthHandler struct {
	register         *authapp.RegisterUseCase
	login            *authapp.LoginUseCase
	refresh          *authapp.RefreshTokenUseCase
	forgotPassword   *authapp.ForgotPasswordUseCase
	resetPassword    *authapp.ResetPasswordUseCase
	logout           *authapp.LogoutUseCase
	exposeResetToken bool
}

func NewAuthHandler(register *authapp.RegisterUseCase, login *authapp.LoginUseCase, refresh *authapp.RefreshTokenUseCase, forgot *authapp.ForgotPasswordUseCase, reset *authapp.ResetPasswordUseCase, logout *authapp.LogoutUseCase, exposeResetToken bool) *AuthHandler {
	return &AuthHandler{register: register, login: login, refresh: refresh, forgotPassword: forgot, resetPassword: reset, logout: logout, exposeResetToken: exposeResetToken}
}

type registerRequest struct {
	UserName string `json:"username" validate:"required,min=3,max=100"`
	Password string `json:"password" validate:"required,min=8,max=72"`
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

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	body, err := request.DecodeJSON[registerRequest](w, r)
	if err != nil {
		response.Error(w, err)
		return
	}
	result, err := h.register.Execute(r.Context(), authapp.RegisterInput{UserName: body.UserName, Password: body.Password})
	if err != nil {
		response.Error(w, mapAuthError(err))
		return
	}
	response.Created(w, result)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	body, err := request.DecodeJSON[loginRequest](w, r)
	if err != nil {
		response.Error(w, err)
		return
	}
	result, err := h.login.Execute(r.Context(), authapp.LoginInput{UserName: body.UserName, Password: body.Password})
	if err != nil {
		response.Error(w, mapAuthError(err))
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
		response.Error(w, mapAuthError(err))
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
		response.Error(w, mapAuthError(err))
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
		response.Error(w, mapAuthError(err))
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
		response.Error(w, mapAuthError(err))
		return
	}
	response.NoContent(w)
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
	default:
		return apperror.Internal(err)
	}
}
