package postgres

import (
	"context"
	"database/sql"

	"falzo-be/internal/auth/application/query"
	"falzo-be/internal/auth/domain"
	"falzo-be/internal/auth/domain/repository"
	"falzo-be/pkg/database"
)

type SessionRepository struct {
	db database.Client
}

func NewSessionRepository(db database.Client) repository.SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, session query.Session) error {
	if r.db == nil || r.db.DB() == nil {
		return domain.ErrAuthDependencyUnavailable
	}

	tx, err := r.db.DB().BeginTx(ctx, nil)
	if err != nil {
		return mapDBError(ctx, "sessions.begin_tx", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO auth_sessions (session_id, user_id, username, subject)
		VALUES ($1, $2, $3, $4)
	`, session.SessionID, session.UserID, session.Username, session.Subject); err != nil {
		return mapDBError(ctx, "sessions.insert_auth_session", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO refresh_tokens (session_id, token_hash, expires_at)
		VALUES ($1, $2, to_timestamp($3))
	`, session.SessionID, session.RefreshTokenHash, session.RefreshExpiresAtUnix); err != nil {
		return mapDBError(ctx, "sessions.insert_refresh_token", err)
	}

	if err := tx.Commit(); err != nil {
		return mapDBError(ctx, "sessions.commit_tx", err)
	}

	return nil
}

func (r *SessionRepository) IsSessionActive(ctx context.Context, sessionID string) (bool, error) {
	if r.db == nil || r.db.DB() == nil {
		return false, domain.ErrAuthDependencyUnavailable
	}

	var active bool
	err := r.db.DB().QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM auth_sessions
			WHERE session_id = $1
			  AND is_revoked = FALSE
		)
	`, sessionID).Scan(&active)
	if err != nil {
		return false, mapDBError(ctx, "sessions.is_active", err)
	}

	return active, nil
}

func (r *SessionRepository) FindActiveByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*query.Session, error) {
	if r.db == nil || r.db.DB() == nil {
		return nil, domain.ErrAuthDependencyUnavailable
	}

	var session query.Session
	err := r.db.DB().QueryRowContext(ctx, `
		SELECT s.session_id, s.user_id, s.username, s.subject, EXTRACT(EPOCH FROM rt.expires_at)::BIGINT
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
		&session.RefreshExpiresAtUnix,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrInvalidToken
		}
		return nil, mapDBError(ctx, "sessions.find_by_refresh_token_hash", err)
	}

	return &session, nil
}

func (r *SessionRepository) RotateRefreshToken(ctx context.Context, sessionID string, refreshTokenHash string, expiresAtUnix int64) error {
	if r.db == nil || r.db.DB() == nil {
		return domain.ErrAuthDependencyUnavailable
	}

	result, err := r.db.DB().ExecContext(ctx, `
		UPDATE refresh_tokens
		SET token_hash = $2,
		    expires_at = to_timestamp($3),
		    updated_at = CURRENT_TIMESTAMP
		WHERE session_id = $1
		  AND is_revoked = FALSE
	`, sessionID, refreshTokenHash, expiresAtUnix)
	if err != nil {
		return mapDBError(ctx, "sessions.rotate_refresh_token", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return mapDBError(ctx, "sessions.rotate_refresh_token.rows_affected", err)
	}
	if affected == 0 {
		return domain.ErrInvalidToken
	}

	return nil
}

func (r *SessionRepository) RevokeBySessionID(ctx context.Context, sessionID string) error {
	if r.db == nil || r.db.DB() == nil {
		return domain.ErrAuthDependencyUnavailable
	}

	result, err := r.db.DB().ExecContext(ctx, `
		UPDATE auth_sessions
		SET is_revoked = TRUE,
		    updated_at = CURRENT_TIMESTAMP
		WHERE session_id = $1
		  AND is_revoked = FALSE
	`, sessionID)
	if err != nil {
		return mapDBError(ctx, "sessions.revoke_auth_session", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return mapDBError(ctx, "sessions.revoke_auth_session.rows_affected", err)
	}
	if affected == 0 {
		return domain.ErrInvalidToken
	}

	if _, err := r.db.DB().ExecContext(ctx, `
		UPDATE refresh_tokens
		SET is_revoked = TRUE,
		    updated_at = CURRENT_TIMESTAMP
		WHERE session_id = $1
		  AND is_revoked = FALSE
	`, sessionID); err != nil {
		return mapDBError(ctx, "sessions.revoke_refresh_tokens", err)
	}

	return nil
}
