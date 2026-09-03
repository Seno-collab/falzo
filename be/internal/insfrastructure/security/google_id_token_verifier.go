package security

import (
	authports "be/internal/application/ports/auth"
	"context"
	"fmt"

	"google.golang.org/api/idtoken"
)

type GoogleIDTokenVerifier struct {
	clientID string
}

func NewGoogleIDTokenVerifier(clientID string) *GoogleIDTokenVerifier {
	return &GoogleIDTokenVerifier{clientID: clientID}
}

func (v *GoogleIDTokenVerifier) Verify(ctx context.Context, credential string) (authports.GoogleIdentity, error) {
	if v.clientID == "" {
		return authports.GoogleIdentity{}, authports.ErrIdentityProviderNotConfigured
	}
	payload, err := idtoken.Validate(ctx, credential, v.clientID)
	if err != nil {
		return authports.GoogleIdentity{}, fmt.Errorf("%w: %v", authports.ErrInvalidExternalIdentity, err)
	}
	if payload.Issuer != "https://accounts.google.com" && payload.Issuer != "accounts.google.com" {
		return authports.GoogleIdentity{}, fmt.Errorf("%w: invalid issuer", authports.ErrInvalidExternalIdentity)
	}

	email, _ := payload.Claims["email"].(string)
	emailVerified, _ := payload.Claims["email_verified"].(bool)
	return authports.GoogleIdentity{
		Subject:       payload.Subject,
		Email:         email,
		EmailVerified: emailVerified,
	}, nil
}
