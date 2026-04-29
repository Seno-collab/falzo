package postgres

import (
	"context"
	"errors"

	"falzo-be/internal/auth/application/query"
	"falzo-be/internal/auth/domain"
	"falzo-be/internal/auth/domain/repository"
	"falzo-be/pkg/database"

	"github.com/jackc/pgx/v5"
)

type SessionRepository struct {
	db database.Client
}

const sessionRepoService = "auth"

func NewSessionRepository(db database.Client) repository.SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, session query.Session) error {
	if r.db == nil || r.db.Pool() == nil {
		return domain.ErrAuthDependencyUnavailable
	}

	tx, err := r.db.Pool().Begin(ctx)
	if err != nil {
		return mapDBError(ctx, sessionRepoService, "sessions.begin_tx", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO auth_sessions (session_id, user_id, user_name, subject)
		VALUES ($1, $2, $3, $4)
	`, session.SessionID, session.UserID, session.Username, session.Subject); err != nil {
		return mapDBError(ctx, sessionRepoService, "sessions.insert_auth_session", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO refresh_tokens (session_id, token_hash, expires_at)
		VALUES ($1, $2, to_timestamp($3))
	`, session.SessionID, session.RefreshTokenHash, session.RefreshExpiresAtUnix); err != nil {
		return mapDBError(ctx, sessionRepoService, "sessions.insert_refresh_token", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return mapDBError(ctx, sessionRepoService, "sessions.commit_tx", err)
	}

	return nil
}

func (r *SessionRepository) IsSessionActive(ctx context.Context, sessionID string) (bool, error) {
	if r.db == nil || r.db.Pool() == nil {
		return false, domain.ErrAuthDependencyUnavailable
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
		return false, mapDBError(ctx, sessionRepoService, "sessions.is_active", err)
	}

	return active, nil
}

func (r *SessionRepository) FindActiveByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*query.Session, error) {
	if r.db == nil || r.db.Pool() == nil {
		return nil, domain.ErrAuthDependencyUnavailable
	}

	var session query.Session
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
			return nil, domain.ErrInvalidToken
		}
		return nil, mapDBError(ctx, sessionRepoService, "sessions.find_by_refresh_token_hash", err)
	}

	return &session, nil
}

func (r *SessionRepository) RotateRefreshToken(ctx context.Context, session query.Session, newRefreshTokenHash string, expiresAtUnix int64) error {
	if r.db == nil || r.db.Pool() == nil {
		return domain.ErrAuthDependencyUnavailable
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
		return mapDBError(ctx, sessionRepoService, "sessions.rotate_refresh_token", err)
	}

	affected := result.RowsAffected()
	if err != nil {
		return mapDBError(ctx, sessionRepoService, "sessions.rotate_refresh_token.rows_affected", err)
	}
	if affected == 0 {
		return domain.ErrInvalidToken
	}

	return nil
}

func (r *SessionRepository) RevokeBySessionID(ctx context.Context, sessionID string) error {
	if r.db == nil || r.db.Pool() == nil {
		return domain.ErrAuthDependencyUnavailable
	}

	result, err := r.db.Pool().Query(ctx, `
		UPDATE auth_sessions
		SET is_revoked = TRUE,
			updated_at = CURRENT_TIMESTAMP
		WHERE session_id = $1
			AND is_revoked = FALSE
	`, sessionID)
	if err != nil {
		return mapDBError(ctx, sessionRepoService, "sessions.revoke_auth_session", err)
	}

	affected := result.CommandTag().RowsAffected()
	// If no session was updated, it means the session ID was invalid or already revoked.
	if affected == 0 {
		return domain.ErrInvalidToken
	}

	if _, err := r.db.Pool().Query(ctx, `
		UPDATE refresh_tokens
		SET is_revoked = TRUE,
		    updated_at = CURRENT_TIMESTAMP
		WHERE session_id = $1
			AND is_revoked = FALSE
	`, sessionID); err != nil {
		return mapDBError(ctx, sessionRepoService, "sessions.revoke_refresh_tokens", err)
	}

	return nil
}
