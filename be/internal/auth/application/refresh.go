package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"falzo-be/internal/auth/application/command"
	"falzo-be/internal/auth/application/query"
	"falzo-be/internal/auth/domain"
)

func (s *service) Refresh(ctx context.Context, cmd command.Refresh) (query.TokenPair, error) {
	if s.sessions == nil || s.tokenIssuer == nil {
		return query.TokenPair{}, domain.ErrAuthDependencyUnavailable
	}
	if cmd.RefreshToken == "" {
		return query.TokenPair{}, domain.ErrInvalidToken
	}

	session, err := s.sessions.FindActiveByRefreshTokenHash(ctx, tokenHash(cmd.RefreshToken))
	if err != nil {
		return query.TokenPair{}, err
	}

	principal := query.AuthenticatedUser{
		UserID:    session.UserID,
		Username:  session.Username,
		Subject:   session.Subject,
		SessionID: session.SessionID,
	}

	accessToken, err := s.tokenIssuer.Issue(principal)
	if err != nil {
		return query.TokenPair{}, err
	}

	refreshToken, err := newOpaqueToken()
	if err != nil {
		return query.TokenPair{}, err
	}

	if err := s.sessions.RotateRefreshToken(ctx, session.SessionID, tokenHash(refreshToken), time.Now().Add(s.refreshTTL).Unix()); err != nil {
		return query.TokenPair{}, err
	}

	return query.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	}, nil
}

func newOpaqueToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}
