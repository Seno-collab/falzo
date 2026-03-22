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

	_, err := r.db.DB().ExecContext(ctx, `
		INSERT INTO refresh_tokens (session_id, user_id, username, subject, token_hash, expires_at)
		VALUES ($1, $2, $3, $4, $5, to_timestamp($6))
	`, session.SessionID, session.UserID, session.Username, session.Subject, session.RefreshTokenHash, session.ExpiresAtUnix)
	return mapDBError(ctx, "sessions.create", err)
}

func (r *SessionRepository) IsSessionActive(ctx context.Context, sessionID string) (bool, error) {
	if r.db == nil || r.db.DB() == nil {
		return false, domain.ErrAuthDependencyUnavailable
	}

	var active bool
	err := r.db.DB().QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM refresh_tokens
			WHERE session_id = $1
			  AND is_revoked = FALSE
			  AND expires_at > NOW()
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
		SELECT session_id, user_id, username, subject, EXTRACT(EPOCH FROM expires_at)::BIGINT
		FROM refresh_tokens
		WHERE token_hash = $1
		  AND is_revoked = FALSE
		  AND expires_at > NOW()
		LIMIT 1
	`, refreshTokenHash).Scan(
		&session.SessionID,
		&session.UserID,
		&session.Username,
		&session.Subject,
		&session.ExpiresAtUnix,
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
		    expires_at = to_timestamp($3)
		WHERE session_id = $1
		  AND is_revoked = FALSE
	`, sessionID, refreshTokenHash, expiresAtUnix)
	if err != nil {
		return err
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
		UPDATE refresh_tokens
		SET is_revoked = TRUE
		WHERE session_id = $1
	`, sessionID)
	if err != nil {
		return mapDBError(ctx, "sessions.revoke_by_session_id", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return mapDBError(ctx, "sessions.revoke_by_session_id.rows_affected", err)
	}
	if affected == 0 {
		return domain.ErrInvalidToken
	}

	return nil
}
