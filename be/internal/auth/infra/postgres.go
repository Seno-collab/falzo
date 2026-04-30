package infra

import (
	"context"
	"errors"
	"time"

	"falzo-be/internal/auth"
	"falzo-be/internal/share"
	"falzo-be/pkg/database"
	"falzo-be/pkg/dberr"

	"github.com/jackc/pgx/v5"
)

const authRepoService = "auth"
const sessionCleanupJobName = "auth_session_cleanup"
const jobConfigsChangedChannel = "job_configs_changed"

type AccountRepository struct {
	db database.Client
}

func NewAccountRepository(db database.Client) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Save(ctx context.Context, account *auth.Account) error {
	if r.db == nil || r.db.Pool() == nil {
		return auth.ErrDependencyUnavailable
	}

	tx, err := r.db.Pool().Begin(ctx)
	if err != nil {
		return share.MapDBError(ctx, authRepoService, "accounts.begin_tx", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO users (user_name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id
	`, account.User.Username.String(), account.User.Email.String(), account.User.PasswordHash.String()).
		Scan(&account.User.ID)
	if err != nil {
		if dberr.IsUniqueViolation(err) {
			return auth.ErrUserExists
		}
		return share.MapDBError(ctx, authRepoService, "accounts.insert_user", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}

	for _, role := range account.Roles {
		var roleID uint64
		err := tx.QueryRow(ctx, `SELECT id FROM roles WHERE name = $1 LIMIT 1`, role).Scan(&roleID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			return share.MapDBError(ctx, authRepoService, "accounts.select_role", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO user_roles (user_id, role_id)
			VALUES ($1, $2)
			ON CONFLICT (user_id, role_id) DO NOTHING
		`, account.User.ID, roleID); err != nil {
			return share.MapDBError(ctx, authRepoService, "accounts.insert_user_role", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return share.MapDBError(ctx, authRepoService, "accounts.commit_tx", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}

	return nil
}

func (r *AccountRepository) FindActiveByEmail(ctx context.Context, email auth.Email) (*auth.Account, error) {
	if r.db == nil || r.db.Pool() == nil {
		return nil, auth.ErrDependencyUnavailable
	}

	var (
		user            auth.User
		rawUsername     string
		rawEmail        string
		rawPasswordHash string
	)

	err := r.db.Pool().QueryRow(ctx, `
		SELECT id, user_name, email, password_hash, is_active, created_at, updated_at
		FROM users
		WHERE email = $1 AND is_active = TRUE
		LIMIT 1
	`, email.String()).Scan(
		&user.ID,
		&rawUsername,
		&rawEmail,
		&rawPasswordHash,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, auth.ErrInvalidCredentials
		}
		return nil, share.MapDBError(ctx, authRepoService, "accounts.find_active_by_email", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}

	user.Username, err = auth.NewUsername(rawUsername)
	if err != nil {
		return nil, err
	}
	user.Email, err = auth.NewEmail(rawEmail)
	if err != nil {
		return nil, err
	}
	user.PasswordHash, err = auth.NewPasswordHash(rawPasswordHash)
	if err != nil {
		return nil, err
	}

	roles, err := r.loadRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return auth.RehydrateAccount(user, roles), nil
}

func (r *AccountRepository) loadRoles(ctx context.Context, userID uint64) ([]string, error) {
	rows, err := r.db.Pool().Query(ctx, `
		SELECT roles.name
		FROM user_roles
		JOIN roles ON roles.id = user_roles.role_id
		WHERE user_roles.user_id = $1
	`, userID)
	if err != nil {
		return nil, share.MapDBError(ctx, authRepoService, "accounts.load_roles", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, share.MapDBError(ctx, authRepoService, "accounts.scan_role", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, share.MapDBError(ctx, authRepoService, "accounts.iterate_roles", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}

	return roles, nil
}

type SessionRepository struct {
	db database.Client
}

func NewSessionRepository(db database.Client) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, session auth.Session) error {
	if r.db == nil || r.db.Pool() == nil {
		return auth.ErrDependencyUnavailable
	}

	tx, err := r.db.Pool().Begin(ctx)
	if err != nil {
		return share.MapDBError(ctx, authRepoService, "sessions.begin_tx", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO auth_sessions (session_id, user_id, user_name, subject)
		VALUES ($1, $2, $3, $4)
	`, session.SessionID, session.UserID, session.Username, session.Subject); err != nil {
		return share.MapDBError(ctx, authRepoService, "sessions.insert_auth_session", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO refresh_tokens (session_id, token_hash, expires_at)
		VALUES ($1, $2, to_timestamp($3))
	`, session.SessionID, session.RefreshTokenHash, session.RefreshExpiresAtUnix); err != nil {
		return share.MapDBError(ctx, authRepoService, "sessions.insert_refresh_token", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}

	if err := tx.Commit(ctx); err != nil {
		return share.MapDBError(ctx, authRepoService, "sessions.commit_tx", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}

	return nil
}

func (r *SessionRepository) IsSessionActive(ctx context.Context, sessionID string) (bool, error) {
	if r.db == nil || r.db.Pool() == nil {
		return false, auth.ErrDependencyUnavailable
	}

	var active bool
	err := r.db.Pool().QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM auth_sessions
			WHERE session_id = $1
				AND is_revoked = FALSE
		)
	`, sessionID).Scan(&active)
	if err != nil {
		return false, share.MapDBError(ctx, authRepoService, "sessions.is_active", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}

	return active, nil
}

func (r *SessionRepository) FindActiveByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*auth.Session, error) {
	if r.db == nil || r.db.Pool() == nil {
		return nil, auth.ErrDependencyUnavailable
	}

	var session auth.Session
	err := r.db.Pool().QueryRow(ctx, `
		SELECT s.session_id, s.user_id, s.user_name, s.subject, rt.token_hash, EXTRACT(EPOCH FROM rt.expires_at)::BIGINT
		FROM refresh_tokens rt
		JOIN auth_sessions s ON s.session_id = rt.session_id
		WHERE rt.token_hash = $1
			AND rt.is_revoked = FALSE
			AND rt.expires_at > NOW()
			AND s.is_revoked = FALSE
		LIMIT 1
	`, refreshTokenHash).Scan(
		&session.SessionID,
		&session.UserID,
		&session.Username,
		&session.Subject,
		&session.RefreshTokenHash,
		&session.RefreshExpiresAtUnix,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, auth.ErrInvalidToken
		}
		return nil, share.MapDBError(ctx, authRepoService, "sessions.find_by_refresh_token_hash", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}

	return &session, nil
}

func (r *SessionRepository) RotateRefreshToken(ctx context.Context, session auth.Session, newRefreshTokenHash string, expiresAtUnix int64) error {
	if r.db == nil || r.db.Pool() == nil {
		return auth.ErrDependencyUnavailable
	}

	result, err := r.db.Pool().Exec(ctx, `
		UPDATE refresh_tokens
		SET token_hash = $2,
		    expires_at = to_timestamp($3),
		    updated_at = CURRENT_TIMESTAMP
		WHERE session_id = $1
			AND token_hash = $4
			AND is_revoked = FALSE
	`, session.SessionID, newRefreshTokenHash, expiresAtUnix, session.RefreshTokenHash)
	if err != nil {
		return share.MapDBError(ctx, authRepoService, "sessions.rotate_refresh_token", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}
	if result.RowsAffected() == 0 {
		return auth.ErrInvalidToken
	}

	return nil
}

func (r *SessionRepository) RevokeBySessionID(ctx context.Context, sessionID string) error {
	if r.db == nil || r.db.Pool() == nil {
		return auth.ErrDependencyUnavailable
	}

	result, err := r.db.Pool().Exec(ctx, `
		UPDATE auth_sessions
		SET is_revoked = TRUE,
			updated_at = CURRENT_TIMESTAMP
		WHERE session_id = $1
			AND is_revoked = FALSE
	`, sessionID)
	if err != nil {
		return share.MapDBError(ctx, authRepoService, "sessions.revoke_auth_session", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}
	if result.RowsAffected() == 0 {
		return auth.ErrInvalidToken
	}

	if _, err := r.db.Pool().Exec(ctx, `
		UPDATE refresh_tokens
		SET is_revoked = TRUE,
		    updated_at = CURRENT_TIMESTAMP
		WHERE session_id = $1
			AND is_revoked = FALSE
	`, sessionID); err != nil {
		return share.MapDBError(ctx, authRepoService, "sessions.revoke_refresh_tokens", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}

	return nil
}

func (r *SessionRepository) CleanupExpired(ctx context.Context, retention time.Duration) (int64, error) {
	if r.db == nil || r.db.Pool() == nil {
		return 0, auth.ErrDependencyUnavailable
	}

	if retention < 0 {
		retention = 0
	}

	cutoff := time.Now().UTC().Add(-retention)
	result, err := r.db.Pool().Exec(ctx, `
		DELETE FROM auth_sessions s
		WHERE EXISTS (
			SELECT 1
			FROM refresh_tokens rt
			WHERE rt.session_id = s.session_id
				AND (
					rt.expires_at < $1
					OR (rt.is_revoked = TRUE AND rt.updated_at < $1)
					OR (s.is_revoked = TRUE AND s.updated_at < $1)
				)
		)
	`, cutoff)
	if err != nil {
		return 0, share.MapDBError(ctx, authRepoService, "sessions.cleanup_expired", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}

	return result.RowsAffected(), nil
}

func (r *SessionRepository) SessionCleanupConfig(ctx context.Context) (auth.SessionCleanupConfig, error) {
	if r.db == nil || r.db.Pool() == nil {
		return auth.SessionCleanupConfig{}, auth.ErrDependencyUnavailable
	}

	var (
		cfg              auth.SessionCleanupConfig
		intervalSeconds  int64
		retentionSeconds int64
	)
	err := r.db.Pool().QueryRow(ctx, `
		SELECT enabled, interval_seconds, retention_seconds
		FROM job_configs
		WHERE job_name = $1
		LIMIT 1
	`, sessionCleanupJobName).Scan(&cfg.Enabled, &intervalSeconds, &retentionSeconds)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return auth.SessionCleanupConfig{
				Enabled:   false,
				Interval:  time.Minute,
				Retention: 0,
			}, nil
		}
		return auth.SessionCleanupConfig{}, share.MapDBError(ctx, authRepoService, "sessions.cleanup_config", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}

	cfg.Interval = time.Duration(intervalSeconds) * time.Second
	cfg.Retention = time.Duration(retentionSeconds) * time.Second
	return cfg, nil
}

func (r *SessionRepository) WaitSessionCleanupConfigChange(ctx context.Context) error {
	if r.db == nil || r.db.Pool() == nil {
		return auth.ErrDependencyUnavailable
	}

	conn, err := r.db.Pool().Acquire(ctx)
	if err != nil {
		return share.MapDBError(ctx, authRepoService, "sessions.cleanup_config_listener.acquire", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "LISTEN "+jobConfigsChangedChannel); err != nil {
		return share.MapDBError(ctx, authRepoService, "sessions.cleanup_config_listener.listen", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}

	_, err = conn.Conn().WaitForNotification(ctx)
	return err
}

var _ auth.AccountRepository = (*AccountRepository)(nil)
var _ auth.SessionRepository = (*SessionRepository)(nil)
