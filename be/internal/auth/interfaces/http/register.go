package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"falzo-be/internal/auth/application/command"
	"falzo-be/internal/auth/domain"
	httpresponse "falzo-be/pkg/response"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresponse.Error(w, http.StatusBadRequest, "Validation failed", r, httpresponse.ErrorDetail{
			Code:    "INVALID_FORMAT",
			Message: "Invalid JSON payload",
		})
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		httpresponse.Error(w, http.StatusBadRequest, "Validation failed", r, httpresponse.ErrorDetail{
			Code:    "REQUIRED_FIELD",
			Message: "Username, email and password are required",
		})
		return
	}

	err := h.service.Register(r.Context(), command.Register{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		status := http.StatusInternalServerError
		message := "Register failed"
		errorCode := "INTERNAL_ERROR"
		if errors.Is(err, domain.ErrUserExists) {
			status = http.StatusConflict
			message = "Account already exists"
			errorCode = "ALREADY_EXISTS"
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

	httpresponse.Success(w, http.StatusCreated, "Account created successfully", map[string]string{
		"message": "account created",
	}, r)
}
