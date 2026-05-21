package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

type Service struct {
	accounts     AccountRepository
	sessions     SessionRepository
	passwords    PasswordHasher
	tokenIssuer  TokenIssuer
	tokenAuth    TokenAuthenticator
	avatarEvents AvatarEventPublisher
	refreshTTL   time.Duration
}

func NewService(
	accounts AccountRepository,
	sessions SessionRepository,
	passwords PasswordHasher,
	tokenIssuer TokenIssuer,
	tokenAuth TokenAuthenticator,
	refreshTTL time.Duration,
) *Service {
	return &Service{
		accounts:    accounts,
		sessions:    sessions,
		passwords:   passwords,
		tokenIssuer: tokenIssuer,
		tokenAuth:   tokenAuth,
		refreshTTL:  refreshTTL,
	}
}

func (s *Service) SetAvatarEventPublisher(publisher AvatarEventPublisher) {
	s.avatarEvents = publisher
}

type RegisterInput struct {
	Username string
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type RefreshInput struct {
	RefreshToken string
}

type LogoutInput struct {
	Token string
}

type ChangePasswordInput struct {
	UserID          uint64
	CurrentPassword string
	NewPassword     string
}

type UpdateAvatarInput struct {
	UserID    uint64
	AvatarURL string
}

func (s *Service) Register(ctx context.Context, input RegisterInput) error {
	if s.accounts == nil || s.passwords == nil {
		return ErrDependencyUnavailable
	}

	username, err := NewUsername(input.Username)
	if err != nil {
		return err
	}
	email, err := NewEmail(input.Email)
	if err != nil {
		return err
	}
	password, err := NewRawPassword(input.Password)
	if err != nil {
		return err
	}

	hash, err := s.passwords.Hash(password)
	if err != nil {
		return err
	}

	account := NewAccount(username, email, hash, []string{"user"})
	return s.accounts.Save(ctx, account)
}

func (s *Service) Login(ctx context.Context, input LoginInput) (TokenPair, error) {
	if s.accounts == nil || s.sessions == nil || s.passwords == nil || s.tokenIssuer == nil || s.tokenAuth == nil {
		return TokenPair{}, ErrDependencyUnavailable
	}

	email, err := NewEmail(input.Email)
	if err != nil {
		return TokenPair{}, ErrInvalidCredentials
	}
	password, err := NewRawPassword(input.Password)
	if err != nil {
		return TokenPair{}, ErrInvalidCredentials
	}

	account, err := s.accounts.FindActiveByEmail(ctx, email)
	if err != nil {
		return TokenPair{}, err
	}
	if err := s.passwords.Compare(account.User.PasswordHash, password); err != nil {
		return TokenPair{}, ErrInvalidCredentials
	}

	principal, err := principalFromAccount(account)
	if err != nil {
		return TokenPair{}, err
	}

	accessToken, err := s.tokenIssuer.Issue(principal)
	if err != nil {
		return TokenPair{}, err
	}

	authenticatedPrincipal, err := s.tokenAuth.Authenticate(accessToken)
	if err != nil {
		return TokenPair{}, err
	}
	if authenticatedPrincipal.ExpiresAt == nil {
		return TokenPair{}, ErrInvalidToken
	}

	refreshToken, err := newOpaqueToken()
	if err != nil {
		return TokenPair{}, err
	}

	if err := s.sessions.Create(ctx, Session{
		SessionID:            principal.SessionID,
		UserID:               principal.UserID,
		Username:             principal.Username,
		Subject:              authenticatedPrincipal.Subject,
		RefreshTokenHash:     tokenHash(refreshToken),
		RefreshExpiresAtUnix: time.Now().Add(s.refreshTTL).Unix(),
	}); err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	}, nil
}

func (s *Service) Refresh(ctx context.Context, input RefreshInput) (TokenPair, error) {
	if s.sessions == nil || s.tokenIssuer == nil {
		return TokenPair{}, ErrDependencyUnavailable
	}
	if input.RefreshToken == "" {
		return TokenPair{}, ErrInvalidToken
	}

	session, err := s.sessions.FindActiveByRefreshTokenHash(ctx, tokenHash(input.RefreshToken))
	if err != nil {
		return TokenPair{}, err
	}

	principal := AuthenticatedUser{
		UserID:    session.UserID,
		Username:  session.Username,
		Subject:   session.Subject,
		SessionID: session.SessionID,
	}

	accessToken, err := s.tokenIssuer.Issue(principal)
	if err != nil {
		return TokenPair{}, err
	}

	refreshToken, err := newOpaqueToken()
	if err != nil {
		return TokenPair{}, err
	}

	if err := s.sessions.RotateRefreshToken(ctx, *session, tokenHash(refreshToken), time.Now().Add(s.refreshTTL).Unix()); err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	}, nil
}

func (s *Service) Logout(ctx context.Context, input LogoutInput) error {
	if s.sessions == nil || s.tokenAuth == nil {
		return ErrDependencyUnavailable
	}
	if input.Token == "" {
		return ErrInvalidToken
	}

	principal, err := s.tokenAuth.Authenticate(input.Token)
	if err != nil {
		return err
	}

	return s.sessions.RevokeBySessionID(ctx, principal.SessionID)
}

func (s *Service) ChangePassword(ctx context.Context, input ChangePasswordInput) error {
	if s.accounts == nil || s.passwords == nil {
		return ErrDependencyUnavailable
	}
	if input.UserID == 0 {
		return ErrInvalidCredentials
	}

	currentPassword, err := NewRawPassword(input.CurrentPassword)
	if err != nil {
		return ErrInvalidCredentials
	}
	newPassword, err := NewRawPassword(input.NewPassword)
	if err != nil {
		return err
	}

	account, err := s.accounts.FindActiveByID(ctx, input.UserID)
	if err != nil {
		return err
	}
	if err := s.passwords.Compare(account.User.PasswordHash, currentPassword); err != nil {
		return ErrInvalidCredentials
	}

	hash, err := s.passwords.Hash(newPassword)
	if err != nil {
		return err
	}

	return s.accounts.UpdatePasswordHash(ctx, input.UserID, hash)
}

func (s *Service) Profile(ctx context.Context, userID uint64) (UserProfile, error) {
	if s.accounts == nil {
		return UserProfile{}, ErrDependencyUnavailable
	}
	if userID == 0 {
		return UserProfile{}, ErrInvalidCredentials
	}

	account, err := s.accounts.FindActiveByID(ctx, userID)
	if err != nil {
		return UserProfile{}, err
	}

	return profileFromAccount(account), nil
}

func (s *Service) UpdateAvatar(ctx context.Context, input UpdateAvatarInput) (UserProfile, error) {
	if s.accounts == nil {
		return UserProfile{}, ErrDependencyUnavailable
	}
	if input.UserID == 0 {
		return UserProfile{}, ErrInvalidCredentials
	}

	avatarURL, err := NewAvatarURL(input.AvatarURL)
	if err != nil {
		return UserProfile{}, err
	}

	if err := s.accounts.UpdateAvatarURL(ctx, input.UserID, avatarURL); err != nil {
		return UserProfile{}, err
	}

	profile, err := s.Profile(ctx, input.UserID)
	if err != nil {
		return UserProfile{}, err
	}
	if s.avatarEvents != nil {
		_ = s.avatarEvents.PublishAvatarUpdated(ctx, AvatarUpdatedEvent{
			UserID:         profile.UserID,
			AvatarURL:      profile.AvatarURL,
			AvatarURLAlias: profile.AvatarURLAlias,
		})
	}

	return profile, nil
}

func (s *Service) Authenticate(ctx context.Context, rawToken string) (*AuthenticatedUser, error) {
	if s.tokenAuth == nil || s.sessions == nil {
		return nil, ErrInvalidToken
	}

	principal, err := s.tokenAuth.Authenticate(rawToken)
	if err != nil {
		return nil, err
	}

	active, err := s.sessions.IsSessionActive(ctx, principal.SessionID)
	if err != nil {
		return nil, err
	}
	if !active {
		return nil, ErrSessionRevoked
	}

	return principal, nil
}

func principalFromAccount(account *Account) (AuthenticatedUser, error) {
	sessionID, err := newSessionID()
	if err != nil {
		return AuthenticatedUser{}, err
	}

	return AuthenticatedUser{
		UserID:    account.User.ID,
		Username:  account.User.Username.String(),
		Roles:     append([]string(nil), account.Roles...),
		SessionID: sessionID,
	}, nil
}

func profileFromAccount(account *Account) UserProfile {
	if account == nil {
		return UserProfile{}
	}

	avatarURL := account.User.AvatarURL.String()
	return UserProfile{
		UserID:         account.User.ID,
		UserIDAlias:    account.User.ID,
		Username:       account.User.Username.String(),
		UsernameAlias:  account.User.Username.String(),
		Email:          account.User.Email.String(),
		AvatarURL:      avatarURL,
		AvatarURLAlias: avatarURL,
	}
}

func newSessionID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}

func newOpaqueToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}

func tokenHash(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}
