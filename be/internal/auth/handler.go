package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"falzo-be/internal/share"
	"falzo-be/pkg/config"
	httpResponse "falzo-be/pkg/response"

	"github.com/go-chi/chi/v5"
)

const Unauthorized = "Unauthorized"

var (
	errMissingBearerToken         = errors.New("missing bearer token")
	errInvalidAuthorizationHeader = errors.New("invalid authorization header")
	errInvalidJSONPayload         = errors.New("invalid JSON payload")
	errRegisterFieldsRequired     = errors.New("register fields required")
	errLoginFieldsRequired        = errors.New("login fields required")
	errInvalidEmailField          = errors.New("invalid email field")
	errRegisterRateLimited        = errors.New("register rate limited")
	errLoginRateLimited           = errors.New("login rate limited")
	errRefreshRateLimited         = errors.New("refresh rate limited")
	errRefreshTokenRequired       = errors.New("refresh token required")
	errMissingAuthContext         = errors.New("missing auth context")
)

type handlerService interface {
	Register(ctx context.Context, input RegisterInput) error
	Login(ctx context.Context, input LoginInput) (TokenPair, error)
	Refresh(ctx context.Context, input RefreshInput) (TokenPair, error)
	Logout(ctx context.Context, input LogoutInput) error
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
	})
	return r
}

type contextKey string

const principalContextKey contextKey = "auth_principal"

func withAuthenticatedUser(ctx context.Context, principal *AuthenticatedUser) context.Context {
	return context.WithValue(ctx, principalContextKey, principal)
}

func authenticatedUserFromContext(ctx context.Context) (*AuthenticatedUser, bool) {
	principal, ok := ctx.Value(principalContextKey).(*AuthenticatedUser)
	return principal, ok
}

func RequireAuth(service interface {
	Authenticate(ctx context.Context, rawToken string) (*AuthenticatedUser, error)
}) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				share.WriteError(w, r, errMissingBearerToken, "authenticate", mapAuthError)
				return
			}

			token, ok := strings.CutPrefix(authHeader, "Bearer ")
			if !ok || token == "" {
				share.WriteError(w, r, errInvalidAuthorizationHeader, "authenticate", mapAuthError)
				return
			}

			principal, err := service.Authenticate(r.Context(), token)
			if err != nil {
				share.WriteError(w, r, err, "authenticate", mapAuthError)
				return
			}

			next.ServeHTTP(w, r.WithContext(withAuthenticatedUser(r.Context(), principal)))
		})
	}
}

type RegisterRequest struct {
	Username string `json:"user_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
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

	if req.Username == "" || req.Email == "" || req.Password == "" {
		share.WriteError(w, r, errRegisterFieldsRequired, "register", mapAuthError)
		return
	}

	if !h.protector.registerLimiter.allow(authClientIP(r), now) {
		share.WriteError(w, r, errRegisterRateLimited, "register", mapAuthError)
		return
	}

	err := h.service.Register(r.Context(), RegisterInput{
		Username: req.Username,
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

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := authenticatedUserFromContext(r.Context())
	if !ok {
		share.WriteError(w, r, errMissingAuthContext, "me", mapAuthError)
		return
	}

	httpResponse.Success(w, http.StatusOK, "Authenticated user fetched successfully", map[string]any{
		"user_name": claims.Username,
		"subject":   claims.Subject,
		"expires":   claims.ExpiresAt,
	}, r)
}

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
	case errors.Is(err, ErrDependencyUnavailable) || errors.Is(err, ErrTemporarilyUnavailable):
		return share.ServiceUnavailable("Authentication service unavailable", "Authentication service is temporarily unavailable")
	default:
		return share.Internal()
	}
}
