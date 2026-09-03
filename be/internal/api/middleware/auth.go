package middleware

import (
	"be/internal/api/http/response"
	authports "be/internal/application/ports/auth"
	"be/internal/shared/apperror"
	"context"
	"net/http"
	"strings"
	"time"
)

type Principal struct {
	UserID   int64
	UserName string
}

const webSocketBearerPrefix = "bearer."

type principalContextKey struct{}
type Authenticator struct{ tokens authports.TokenManager }

func NewAuthenticator(tokens authports.TokenManager) *Authenticator {
	return &Authenticator{tokens: tokens}
}

func (a *Authenticator) RequireAccessToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			response.Error(w, apperror.Unauthorized("Bearer token is required"))
			return
		}
		a.authenticate(next, w, r, token)
	})
}

func (a *Authenticator) RequireWebSocketAccessToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			token, ok = webSocketProtocolToken(r.Header.Get("Sec-WebSocket-Protocol"))
		}
		if !ok {
			response.Error(w, apperror.Unauthorized("WebSocket access token is required"))
			return
		}
		a.authenticate(next, w, r, token)
	})
}

func (a *Authenticator) authenticate(next http.Handler, w http.ResponseWriter, r *http.Request, token string) {
	claims, err := a.tokens.ParseAccess(token, time.Now().UTC())
	if err != nil {
		response.Error(w, apperror.InvalidToken())
		return
	}
	ctx := context.WithValue(r.Context(), principalContextKey{}, Principal{UserID: claims.UserID, UserName: claims.UserName})
	next.ServeHTTP(w, r.WithContext(ctx))
}

func bearerToken(authorization string) (string, bool) {
	parts := strings.Fields(authorization)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

func webSocketProtocolToken(header string) (string, bool) {
	for protocol := range strings.SplitSeq(header, ",") {
		protocol = strings.TrimSpace(protocol)
		if len(protocol) > len(webSocketBearerPrefix) && strings.EqualFold(protocol[:len(webSocketBearerPrefix)], webSocketBearerPrefix) {
			return protocol[len(webSocketBearerPrefix):], true
		}
	}
	return "", false
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalContextKey{}).(Principal)
	return p, ok
}
