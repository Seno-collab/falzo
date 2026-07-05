package postgres

import (
	domainuser "be/internal/domain/user"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

const userColumns = `id, username, password_hash, status, failed_attempts, locked_until, created_at, updated_at`

func (r *UserRepository) FindByUserName(ctx context.Context, username string) (*domainuser.User, error) {
	return scanUser(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE username = $1`, username))
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*domainuser.User, error) {
	return scanUser(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, id))
}

func (r *UserRepository) Create(ctx context.Context, user *domainuser.User) error {
	err := r.db.QueryRow(ctx, `
		INSERT INTO users (username, password_hash, status, failed_attempts, locked_until, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		user.UserName, user.PasswordHash, user.Status, user.FailedAttempts, user.LockUntil, user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return domainuser.ErrUserNameAlreadyExists
	}
	return err
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
	err := row.Scan(&u.ID, &u.UserName, &u.PasswordHash, &u.Status, &u.FailedAttempts, &u.LockUntil, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domainuser.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: db,
	}
}
