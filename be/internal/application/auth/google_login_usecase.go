package authapp

import (
	authports "be/internal/application/ports/auth"
	domainuser "be/internal/domain/user"
	"be/internal/shared/clock"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const googleProvider = "google"

var (
	ErrInvalidGoogleCredential  = errors.New("invalid google credential")
	ErrGoogleLoginNotConfigured = errors.New("google login is not configured")
)

type GoogleLoginUseCase struct {
	identityVerifier authports.GoogleIdentityVerifier
	users            authports.UserRepository
	tokens           authports.TokenManager
	sessions         authports.TokenSessionStore
	clock            clock.Clock
}

type GoogleLoginInput struct {
	Credential string
}

func NewGoogleLoginUseCase(
	identityVerifier authports.GoogleIdentityVerifier,
	users authports.UserRepository,
	tokens authports.TokenManager,
	sessions authports.TokenSessionStore,
	c clock.Clock,
) *GoogleLoginUseCase {
	return &GoogleLoginUseCase{
		identityVerifier: identityVerifier,
		users:            users,
		tokens:           tokens,
		sessions:         sessions,
		clock:            c,
	}
}

func (uc *GoogleLoginUseCase) Execute(ctx context.Context, input GoogleLoginInput) (*LoginOutput, error) {
	identity, err := uc.identityVerifier.Verify(ctx, input.Credential)
	if err != nil {
		if errors.Is(err, authports.ErrIdentityProviderNotConfigured) {
			return nil, ErrGoogleLoginNotConfigured
		}
		return nil, ErrInvalidGoogleCredential
	}
	if identity.Subject == "" || identity.Email == "" || !identity.EmailVerified {
		return nil, ErrInvalidGoogleCredential
	}

	now := uc.clock.Now()
	user, err := uc.users.FindByIdentity(ctx, googleProvider, identity.Subject)
	if errors.Is(err, domainuser.ErrUserNotFound) {
		user, err = uc.users.CreateExternalUser(ctx, googleUsername(identity.Email, identity.Subject), googleProvider, identity.Subject, identity.Email, now)
	}
	if err != nil {
		return nil, err
	}
	if err := user.CanLogin(now); err != nil {
		return nil, err
	}
	user.RecordSuccessfulLogin(now)
	if err := uc.users.UpdateLoginState(ctx, user); err != nil {
		return nil, err
	}

	pair, err := uc.tokens.GeneratePair(user.ID, user.UserName, now)
	if err != nil {
		return nil, err
	}
	if err := uc.sessions.SaveRefresh(ctx, pair.RefreshTokenID, user.ID, pair.RefreshExpiresAt.Sub(now)); err != nil {
		return nil, err
	}
	return &LoginOutput{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(pair.AccessExpiresAt.Sub(now).Seconds()),
		UserName:     user.UserName,
	}, nil
}

func googleUsername(email, subject string) string {
	localPart := strings.SplitN(strings.ToLower(strings.TrimSpace(email)), "@", 2)[0]
	var builder strings.Builder
	for _, char := range localPart {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' || char == '_' {
			builder.WriteRune(char)
		}
	}
	base := strings.Trim(builder.String(), "-_")
	if base == "" {
		base = "google-user"
	}

	hash := sha256.Sum256([]byte(subject))
	suffix := hex.EncodeToString(hash[:])[:10]
	const maxUsernameLength = 100
	maxBaseLength := maxUsernameLength - len(suffix) - 1
	if len(base) > maxBaseLength {
		base = strings.TrimRight(base[:maxBaseLength], "-_")
	}
	return fmt.Sprintf("%s-%s", base, suffix)
}
