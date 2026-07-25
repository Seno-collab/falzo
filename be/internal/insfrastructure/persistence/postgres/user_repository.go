package postgres

import (
	domainuser "be/internal/domain/user"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

const userColumns = `id, username, password_hash, status, failed_attempts, locked_until, created_at, updated_at`
const userColumnsWithUserAlias = `u.id, u.username, u.password_hash, u.status, u.failed_attempts, u.locked_until, u.created_at, u.updated_at`

func (r *UserRepository) FindByUserName(ctx context.Context, username string) (*domainuser.User, error) {
	return scanUser(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE username = $1`, username))
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*domainuser.User, error) {
	return scanUser(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, id))
}

func (r *UserRepository) FindByIdentity(ctx context.Context, provider, subject string) (*domainuser.User, error) {
	return scanUser(r.db.QueryRow(ctx, `
		SELECT `+userColumnsWithUserAlias+`
		FROM users u
		JOIN auth_identities i ON i.user_id = u.id
		WHERE i.provider = $1 AND i.provider_subject = $2`, provider, subject))
}

func (r *UserRepository) CreateExternalUser(ctx context.Context, username, provider, subject, email string, now time.Time) (*domainuser.User, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	user := &domainuser.User{UserName: username, Status: domainuser.UserStatusActive, CreatedAt: now, UpdatedAt: now}
	if err := tx.QueryRow(ctx, `
		INSERT INTO users (username, password_hash, status, created_at, updated_at)
		VALUES ($1, NULL, $2, $3, $4) RETURNING id`,
		user.UserName, user.Status, user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID); err != nil {
		return r.resolveExternalIdentityConflict(ctx, tx, provider, subject, err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO auth_identities (user_id, provider, provider_subject, provider_email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		user.ID, provider, subject, email, now, now,
	); err != nil {
		return r.resolveExternalIdentityConflict(ctx, tx, provider, subject, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) resolveExternalIdentityConflict(
	ctx context.Context,
	tx pgx.Tx,
	provider string,
	subject string,
	cause error,
) (*domainuser.User, error) {
	if !IsUniqueViolation(cause) {
		return nil, cause
	}

	// Concurrent first logins can race between the initial lookup and insert.
	// End the failed transaction before reading the identity created by the winner.
	_ = tx.Rollback(ctx)
	user, err := r.FindByIdentity(ctx, provider, subject)
	if err != nil {
		return nil, cause
	}
	return user, nil
}

func (r *UserRepository) UpdateLoginState(ctx context.Context, user *domainuser.User) error {
	command, err := r.db.Exec(ctx, `UPDATE users SET status=$2, failed_attempts=$3, locked_until=$4, updated_at=$5 WHERE id=$1`, user.ID, user.Status, user.FailedAttempts, user.LockUntil, user.UpdatedAt)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return domainuser.ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, userID int64, passwordHash string) error {
	command, err := r.db.Exec(ctx, `UPDATE users SET password_hash=$2, status=$3, failed_attempts=0, locked_until=NULL, updated_at=now() WHERE id=$1`, userID, passwordHash, domainuser.UserStatusActive)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return domainuser.ErrUserNotFound
	}
	return nil
}

type rowScanner interface{ Scan(dest ...any) error }

func scanUser(row rowScanner) (*domainuser.User, error) {
	u := &domainuser.User{}
	var passwordHash *string
	err := row.Scan(&u.ID, &u.UserName, &passwordHash, &u.Status, &u.FailedAttempts, &u.LockUntil, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domainuser.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	if passwordHash != nil {
		u.PasswordHash = *passwordHash
	}
	return u, nil
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: db,
	}
}
