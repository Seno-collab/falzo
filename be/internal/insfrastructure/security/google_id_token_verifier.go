package security

import (
	"be/internal/application/ports"
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

func (v *GoogleIDTokenVerifier) Verify(ctx context.Context, credential string) (ports.GoogleIdentity, error) {
	if v.clientID == "" {
		return ports.GoogleIdentity{}, ports.ErrIdentityProviderNotConfigured
	}
	payload, err := idtoken.Validate(ctx, credential, v.clientID)
	if err != nil {
		return ports.GoogleIdentity{}, fmt.Errorf("%w: %v", ports.ErrInvalidExternalIdentity, err)
	}
	if payload.Issuer != "https://accounts.google.com" && payload.Issuer != "accounts.google.com" {
		return ports.GoogleIdentity{}, fmt.Errorf("%w: invalid issuer", ports.ErrInvalidExternalIdentity)
	}

	email, _ := payload.Claims["email"].(string)
	emailVerified, _ := payload.Claims["email_verified"].(bool)
	return ports.GoogleIdentity{
		Subject:       payload.Subject,
		Email:         email,
		EmailVerified: emailVerified,
	}, nil
}
