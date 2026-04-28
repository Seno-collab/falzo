package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"falzo-be/internal/auth/application/command"
	"falzo-be/internal/auth/application/query"
	"falzo-be/internal/auth/domain"
	"falzo-be/internal/auth/domain/aggregate"
	"falzo-be/internal/auth/domain/value_object"
	"time"
)

func (s *service) Login(ctx context.Context, cmd command.Login) (query.TokenPair, error) {
	if s.accounts == nil || s.sessions == nil || s.passwords == nil || s.tokenIssuer == nil || s.tokenAuth == nil {
		return query.TokenPair{}, domain.ErrAuthDependencyUnavailable
	}

	email, err := value_object.NewEmail(cmd.Email)
	if err != nil {
		return query.TokenPair{}, domain.ErrInvalidCredentials
	}

	password, err := value_object.NewRawPassword(cmd.Password)
	if err != nil {
		return query.TokenPair{}, domain.ErrInvalidCredentials
	}

	account, err := s.accounts.FindActiveByEmail(ctx, email)
	if err != nil {
		return query.TokenPair{}, err
	}

	if err := s.passwords.Compare(account.User.PasswordHash, password); err != nil {
		return query.TokenPair{}, domain.ErrInvalidCredentials
	}

	principal, err := principalFromAccount(account)
	if err != nil {
		return query.TokenPair{}, err
	}

	accessToken, err := s.tokenIssuer.Issue(principal)
	if err != nil {
		return query.TokenPair{}, err
	}

	authenticatedPrincipal, err := s.tokenAuth.Authenticate(accessToken)
	if err != nil {
		return query.TokenPair{}, err
	}

	if authenticatedPrincipal.ExpiresAt == nil {
		return query.TokenPair{}, domain.ErrInvalidToken
	}

	refreshToken, err := newOpaqueToken()
	if err != nil {
		return query.TokenPair{}, err
	}

	if err := s.sessions.Create(ctx, query.Session{
		SessionID:            principal.SessionID,
		UserID:               principal.UserID,
		Username:             principal.Username,
		Subject:              authenticatedPrincipal.Subject,
		RefreshTokenHash:     tokenHash(refreshToken),
		RefreshExpiresAtUnix: time.Now().Add(s.refreshTTL).Unix(),
	}); err != nil {
		return query.TokenPair{}, err
	}

	return query.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	}, nil
}

func principalFromAccount(account *aggregate.Account) (query.AuthenticatedUser, error) {
	sessionID, err := newSessionID()
	if err != nil {
		return query.AuthenticatedUser{}, err
	}

	return query.AuthenticatedUser{
		UserID:    account.User.ID,
		Username:  account.User.Username.String(),
		SessionID: sessionID,
	}, nil
}

func newSessionID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}
