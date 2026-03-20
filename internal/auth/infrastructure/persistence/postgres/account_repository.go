package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"falzo/internal/auth/domain"
	"falzo/internal/auth/domain/aggregate"
	"falzo/internal/auth/domain/entity"
	"falzo/internal/auth/domain/repository"
	"falzo/internal/auth/domain/valueobject"
	"falzo/pkg/database"
)

type AccountRepository struct {
	db database.Client
}

func NewAccountRepository(db database.Client) repository.AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Save(ctx context.Context, account *aggregate.Account) error {
	if r.db == nil || r.db.DB() == nil {
		return domain.ErrAuthUnavailable
	}

	tx, err := r.db.DB().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id
	`, account.User.Username.String(), account.User.Email.String(), account.User.PasswordHash.String()).
		Scan(&account.User.ID)
	if err != nil {
		if isDuplicateError(err) {
			return domain.ErrUserExists
		}
		return err
	}

	for _, role := range account.Roles {
		var roleID uint64
		err := tx.QueryRowContext(ctx, `SELECT id FROM roles WHERE name = $1 LIMIT 1`, role).Scan(&roleID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return err
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO user_roles (user_id, role_id)
			VALUES ($1, $2)
			ON CONFLICT (user_id, role_id) DO NOTHING
		`, account.User.ID, roleID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *AccountRepository) FindActiveByUsername(ctx context.Context, username valueobject.Username) (*aggregate.Account, error) {
	if r.db == nil || r.db.DB() == nil {
		return nil, domain.ErrAuthUnavailable
	}

	var (
		user            entity.User
		rawUsername     string
		rawEmail        string
		rawPasswordHash string
	)

	err := r.db.DB().QueryRowContext(ctx, `
		SELECT id, username, email, password_hash, is_active, created_at, updated_at
		FROM users
		WHERE username = $1 AND is_active = TRUE
		LIMIT 1
	`, username.String()).Scan(
		&user.ID,
		&rawUsername,
		&rawEmail,
		&rawPasswordHash,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	user.Username, err = valueobject.NewUsername(rawUsername)
	if err != nil {
		return nil, err
	}
	user.Email, err = valueobject.NewEmail(rawEmail)
	if err != nil {
		return nil, err
	}
	user.PasswordHash, err = valueobject.NewPasswordHash(rawPasswordHash)
	if err != nil {
		return nil, err
	}

	roles, err := r.loadRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return aggregate.RehydrateAccount(user, roles), nil
}

func (r *AccountRepository) loadRoles(ctx context.Context, userID uint64) ([]string, error) {
	rows, err := r.db.DB().QueryContext(ctx, `
		SELECT roles.name
		FROM user_roles
		JOIN roles ON roles.id = user_roles.role_id
		WHERE user_roles.user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, rows.Err()
}

func isDuplicateError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate key value violates unique constraint")
}
