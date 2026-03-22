package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"falzo-be/internal/auth/application/command"
	"falzo-be/internal/auth/domain"
	httpresponse "falzo-be/pkg/response"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "Validation failed", r, httpresponse.ErrorDetail{
			Code:    "INVALID_FORMAT",
			Message: "Invalid JSON payload",
		})
		return
	}

	if req.Username == "" || req.Password == "" {
		httpresponse.Error(w, http.StatusBadRequest, "Validation failed", r, httpresponse.ErrorDetail{
			Code:    "REQUIRED_FIELD",
			Message: "Username and password are required",
		})
		return
	}

	token, err := h.service.Login(r.Context(), command.Login{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		status := http.StatusInternalServerError
		message := "Login failed"
		errorCode := "INTERNAL_ERROR"
		if errors.Is(err, domain.ErrInvalidCredentials) {
			status = http.StatusUnauthorized
			message = "Invalid credentials"
			errorCode = "UNAUTHORIZED"
		}
		if errors.Is(err, domain.ErrAuthUnavailable) {
			status = http.StatusServiceUnavailable
			message = "Authentication service unavailable"
			errorCode = "SERVICE_UNAVAILABLE"
		}
		httpresponse.Error(w, status, message, r, httpresponse.ErrorDetail{
			Code:    errorCode,
			Message: err.Error(),
		})
		return
	}

	httpresponse.Success(w, http.StatusOK, "Login successful", LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
	}, r)
}
