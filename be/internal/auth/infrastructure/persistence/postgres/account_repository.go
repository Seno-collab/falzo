package postgres

import (
	"context"
	"database/sql"
	"errors"
	"falzo-be/internal/auth/domain"
	"falzo-be/internal/auth/domain/aggregate"
	"falzo-be/internal/auth/domain/entity"
	"falzo-be/internal/auth/domain/repository"
	"falzo-be/internal/auth/domain/valueobject"
	"falzo-be/pkg/database"
	"falzo-be/pkg/dberr"
)

type AccountRepository struct {
	db database.Client
}

const accountRepoService = "auth"

func NewAccountRepository(db database.Client) repository.AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Save(ctx context.Context, account *aggregate.Account) error {
	if r.db == nil || r.db.DB() == nil {
		return domain.ErrAuthDependencyUnavailable
	}

	tx, err := r.db.DB().BeginTx(ctx, nil)
	if err != nil {
		return mapDBError(ctx, accountRepoService, "accounts.begin_tx", err)
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, `
		INSERT INTO users (user_name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id
	`, account.User.Username.String(), account.User.Email.String(), account.User.PasswordHash.String()).
		Scan(&account.User.ID)
	if err != nil {
		if isDuplicateError(err) {
			return domain.ErrUserExists
		}
		return mapDBError(ctx, accountRepoService, "accounts.insert_user", err)
	}

	for _, role := range account.Roles {
		var roleID uint64
		err := tx.QueryRowContext(ctx, `SELECT id FROM roles WHERE name = $1 LIMIT 1`, role).Scan(&roleID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return mapDBError(ctx, accountRepoService, "accounts.select_role", err)
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO user_roles (user_id, role_id)
			VALUES ($1, $2)
			ON CONFLICT (user_id, role_id) DO NOTHING
		`, account.User.ID, roleID); err != nil {
			return mapDBError(ctx, accountRepoService, "accounts.insert_user_role", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return mapDBError(ctx, accountRepoService, "accounts.commit_tx", err)
	}

	return nil
}

func (r *AccountRepository) FindActiveByEmail(ctx context.Context, email valueobject.Email) (*aggregate.Account, error) {
	if r.db == nil || r.db.DB() == nil {
		return nil, domain.ErrAuthDependencyUnavailable
	}

	var (
		user            entity.User
		rawUsername     string
		rawEmail        string
		rawPasswordHash string
	)

	err := r.db.DB().QueryRowContext(ctx, `
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, mapDBError(ctx, accountRepoService, "accounts.find_active_by_email", err)
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
		return nil, mapDBError(ctx, accountRepoService, "accounts.load_roles", err)
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, mapDBError(ctx, accountRepoService, "accounts.scan_role", err)
		}
		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, mapDBError(ctx, accountRepoService, "accounts.iterate_roles", err)
	}

	return roles, nil
}

func isDuplicateError(err error) bool {
	return dberr.IsUniqueViolation(err)
}
