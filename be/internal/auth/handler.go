package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"falzo-be/internal/share"
	"falzo-be/pkg/config"
	httpResponse "falzo-be/pkg/response"

	"github.com/go-chi/chi/v5"
)

type handlerService interface {
	Register(ctx context.Context, input RegisterInput) error
	Login(ctx context.Context, input LoginInput) (TokenPair, error)
	Refresh(ctx context.Context, input RefreshInput) (TokenPair, error)
	Logout(ctx context.Context, input LogoutInput) error
	ChangePassword(ctx context.Context, input ChangePasswordInput) error
	Authenticate(ctx context.Context, rawToken string) (*AuthenticatedUser, error)
}

type Handler struct {
	service           handlerService
	publicMiddlewares []func(http.Handler) http.Handler
	protector         *authProtector
}

type Option func(*Handler)

func NewHandler(service handlerService, opts ...Option) *Handler {
	h := &Handler{service: service}
	for _, opt := range opts {
		opt(h)
	}

	if h.protector == nil {
		cfg := config.Load()
		h.protector = newAuthProtector(
			cfg.Auth.RateLimitPerMin,
			cfg.Auth.DependencyFailureThreshold,
			cfg.Auth.DependencyCoolDown,
		)
	}

	return h
}

func WithPublicMiddlewares(middlewares ...func(http.Handler) http.Handler) Option {
	return func(h *Handler) {
		h.publicMiddlewares = append(h.publicMiddlewares, middlewares...)
	}
}

func WithProtector(protector *authProtector) Option {
	return func(h *Handler) {
		if protector != nil {
			h.protector = protector
		}
	}
}

func WithProtectorConfig(limitPerMinute int, failureThreshold int, cooldown time.Duration) Option {
	return WithProtector(newAuthProtector(limitPerMinute, failureThreshold, cooldown))
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.With(h.publicMiddlewares...).Post("/register", h.Register)
	r.With(h.publicMiddlewares...).Post("/login", h.Login)
	r.With(h.publicMiddlewares...).Post("/refresh", h.Refresh)
	r.Group(func(protected chi.Router) {
		protected.Use(RequireAuth(h.service))
		protected.Get("/me", h.Me)
		protected.Post("/logout", h.Logout)
		protected.Post("/change-password", h.ChangePassword)
	})
	return r
}

type RegisterRequest struct {
	Username      string `json:"user_name"`
	UsernameAlias string `json:"userName"`
	Email         string `json:"email"`
	Password      string `json:"password"`
}

func (req RegisterRequest) normalizedUsername() string {
	if req.Username != "" {
		return req.Username
	}

	return req.UsernameAlias
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	if !h.protector.allowOperation(now) {
		share.WriteError(w, r, ErrTemporarilyUnavailable, "register", mapAuthError)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		share.WriteError(w, r, errInvalidJSONPayload, "register", mapAuthError)
		return
	}

	username := req.normalizedUsername()
	if username == "" || req.Email == "" || req.Password == "" {
		share.WriteError(w, r, errRegisterFieldsRequired, "register", mapAuthError)
		return
	}

	if !h.protector.registerLimiter.allow(authClientIP(r), now) {
		share.WriteError(w, r, errRegisterRateLimited, "register", mapAuthError)
		return
	}

	err := h.service.Register(r.Context(), RegisterInput{
		Username: username,
		Email:    req.Email,
		Password: req.Password,
	})
	h.protector.observe(err, time.Now())
	if err != nil {
		share.WriteError(w, r, err, "register", mapAuthError)
		return
	}

	httpResponse.Success(w, http.StatusCreated, "Account created successfully", map[string]string{
		"message": "account created",
	}, r)
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	if !h.protector.allowOperation(now) {
		share.WriteError(w, r, ErrTemporarilyUnavailable, "login", mapAuthError)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		share.WriteError(w, r, errInvalidJSONPayload, "login", mapAuthError)
		return
	}

	if req.Email == "" || req.Password == "" {
		share.WriteError(w, r, errLoginFieldsRequired, "login", mapAuthError)
		return
	}

	if _, err := NewEmail(req.Email); err != nil {
		share.WriteError(w, r, errInvalidEmailField, "login", mapAuthError)
		return
	}

	key := authClientIP(r) + ":" + strings.ToLower(strings.TrimSpace(req.Email))
	if !h.protector.loginLimiter.allow(key, now) {
		share.WriteError(w, r, errLoginRateLimited, "login", mapAuthError)
		return
	}

	tokens, err := h.service.Login(r.Context(), LoginInput{Email: req.Email, Password: req.Password})
	h.protector.observe(err, time.Now())
	if err != nil {
		share.WriteError(w, r, err, "login", mapAuthError)
		return
	}

	httpResponse.Success(w, http.StatusOK, "Login successful", tokens, r)
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	if !h.protector.allowOperation(now) {
		share.WriteError(w, r, ErrTemporarilyUnavailable, "refresh", mapAuthError)
		return
	}

	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		share.WriteError(w, r, errInvalidJSONPayload, "refresh", mapAuthError)
		return
	}

	if req.RefreshToken == "" {
		share.WriteError(w, r, errRefreshTokenRequired, "refresh", mapAuthError)
		return
	}

	if !h.protector.refreshLimiter.allow(authClientIP(r), now) {
		share.WriteError(w, r, errRefreshRateLimited, "refresh", mapAuthError)
		return
	}

	tokens, err := h.service.Refresh(r.Context(), RefreshInput{RefreshToken: req.RefreshToken})
	h.protector.observe(err, time.Now())
	if err != nil {
		share.WriteError(w, r, err, "refresh", mapAuthError)
		return
	}

	httpResponse.Success(w, http.StatusOK, "Token refreshed successfully", tokens, r)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	token, ok := strings.CutPrefix(authHeader, "Bearer ")
	if !ok || token == "" {
		share.WriteError(w, r, errInvalidAuthorizationHeader, "logout", mapAuthError)
		return
	}

	if err := h.service.Logout(r.Context(), LogoutInput{Token: token}); err != nil {
		share.WriteError(w, r, err, "logout", mapAuthError)
		return
	}

	httpResponse.Success(w, http.StatusOK, "Logout acknowledged", map[string]string{
		"message": "logout acknowledged",
	}, r)
}

type ChangePasswordRequest struct {
	CurrentPassword      string `json:"current_password"`
	CurrentPasswordAlias string `json:"currentPassword"`
	NewPassword          string `json:"new_password"`
	NewPasswordAlias     string `json:"newPassword"`
}

func (req ChangePasswordRequest) currentPassword() string {
	if req.CurrentPassword != "" {
		return req.CurrentPassword
	}

	return req.CurrentPasswordAlias
}

func (req ChangePasswordRequest) newPassword() string {
	if req.NewPassword != "" {
		return req.NewPassword
	}

	return req.NewPasswordAlias
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	principal, ok := authenticatedUserFromContext(r.Context())
	if !ok || principal == nil || principal.UserID == 0 {
		share.WriteError(w, r, errMissingAuthContext, "change_password", mapAuthError)
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		share.WriteError(w, r, errInvalidJSONPayload, "change_password", mapAuthError)
		return
	}

	currentPassword := req.currentPassword()
	newPassword := req.newPassword()
	if currentPassword == "" || newPassword == "" {
		share.WriteError(w, r, errChangePasswordFieldsRequired, "change_password", mapAuthError)
		return
	}

	if err := h.service.ChangePassword(r.Context(), ChangePasswordInput{
		UserID:          principal.UserID,
		CurrentPassword: currentPassword,
		NewPassword:     newPassword,
	}); err != nil {
		share.WriteError(w, r, err, "change_password", mapAuthError)
		return
	}

	httpResponse.Success(w, http.StatusOK, "Password changed successfully", map[string]string{
		"message": "password changed",
	}, r)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := authenticatedUserFromContext(r.Context())
	if !ok {
		share.WriteError(w, r, errMissingAuthContext, "me", mapAuthError)
		return
	}

	httpResponse.Success(w, http.StatusOK, "Authenticated user fetched successfully", map[string]any{
		"user_id":   claims.UserID,
		"userId":    claims.UserID,
		"user_name": claims.Username,
		"userName":  claims.Username,
		"subject":   claims.Subject,
		"expires":   claims.ExpiresAt,
	}, r)
}
