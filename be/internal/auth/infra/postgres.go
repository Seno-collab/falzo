package infra

import (
	"context"
	"errors"
	"time"

	"falzo-be/internal/auth"
	"falzo-be/internal/share"
	"falzo-be/pkg/database"
	"falzo-be/pkg/database/orm"
	"falzo-be/pkg/dberr"

	"github.com/jackc/pgx/v5"
)

const authRepoService = "auth"
const sessionCleanupJobName = "auth_session_cleanup"
const jobConfigsChangedChannel = "job_configs_changed"

type AccountRepository struct {
	db        database.Client
	users     *orm.Table[auth.User]
	roles     *orm.Table[roleRecord]
	userRoles *orm.Table[userRoleRecord]
}

func NewAccountRepository(db database.Client) *AccountRepository {
	repository := &AccountRepository{db: db}
	if db != nil && db.Pool() != nil {
		repository.users = newUserTable(db.Pool())
		repository.roles = newRoleTable(db.Pool())
		repository.userRoles = newUserRoleTable(db.Pool())
	}
	return repository
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

	users := newUserTable(tx)
	roles := newRoleTable(tx)
	userRoles := newUserRoleTable(tx)

	err = users.InsertReturning(ctx, orm.Values{
		"email":         account.User.Email.String(),
		"password_hash": account.User.PasswordHash.String(),
		"user_name":     account.User.Username.String(),
	}, []string{"id"}, &account.User.ID)
	if err != nil {
		if dberr.IsUniqueViolation(err) {
			return auth.ErrUserExists
		}
		return share.MapDBError(ctx, authRepoService, "accounts.insert_user", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}

	for _, role := range account.Roles {
		roleRecord, err := roles.FindOne(ctx, "name = $1", role)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			return share.MapDBError(ctx, authRepoService, "accounts.select_role", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
		}

		if _, err := userRoles.InsertWithOptions(ctx, orm.Values{
			"role_id": roleRecord.ID,
			"user_id": account.User.ID,
		}, orm.InsertOptions{OnConflictDoNothing: true}); err != nil {
			return share.MapDBError(ctx, authRepoService, "accounts.insert_user_role", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return share.MapDBError(ctx, authRepoService, "accounts.commit_tx", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}

	return nil
}

func (r *AccountRepository) FindActiveByEmail(ctx context.Context, email auth.Email) (*auth.Account, error) {
	users, err := r.userTable()
	if err != nil {
		return nil, err
	}

	user, err := users.FindOne(ctx, "email = $1 AND is_active = TRUE", email.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, auth.ErrInvalidCredentials
		}
		return nil, share.MapDBError(ctx, authRepoService, "accounts.find_active_by_email", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}

	roles, err := r.loadRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return auth.RehydrateAccount(user, roles), nil
}

func (r *AccountRepository) FindActiveByID(ctx context.Context, userID uint64) (*auth.Account, error) {
	users, err := r.userTable()
	if err != nil {
		return nil, err
	}

	user, err := users.FindOne(ctx, "id = $1 AND is_active = TRUE", userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, auth.ErrInvalidCredentials
		}
		return nil, share.MapDBError(ctx, authRepoService, "accounts.find_active_by_id", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}

	roles, err := r.loadRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return auth.RehydrateAccount(user, roles), nil
}

func (r *AccountRepository) UpdatePasswordHash(ctx context.Context, userID uint64, passwordHash auth.PasswordHash) error {
	users, err := r.userTable()
	if err != nil {
		return err
	}

	result, err := users.UpdateWhere(ctx, "id = $3 AND is_active = TRUE", orm.Values{
		"password_hash": passwordHash.String(),
		"updated_at":    time.Now().UTC(),
	}, userID)
	if err != nil {
		return share.MapDBError(ctx, authRepoService, "accounts.update_password_hash", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}
	if result.RowsAffected() == 0 {
		return auth.ErrInvalidCredentials
	}

	return nil
}

func (r *AccountRepository) UpdateAvatarURL(ctx context.Context, userID uint64, avatarURL auth.AvatarURL) error {
	users, err := r.userTable()
	if err != nil {
		return err
	}

	result, err := users.UpdateWhere(ctx, "id = $3 AND is_active = TRUE", orm.Values{
		"avatar_url": avatarURL.String(),
		"updated_at": time.Now().UTC(),
	}, userID)
	if err != nil {
		return share.MapDBError(ctx, authRepoService, "accounts.update_avatar_url", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}
	if result.RowsAffected() == 0 {
		return auth.ErrInvalidCredentials
	}

	return nil
}

func (r *AccountRepository) loadRoles(ctx context.Context, userID uint64) ([]string, error) {
	userRoles := orm.NewTable(r.db.Pool(), "user_roles JOIN roles ON roles.id = user_roles.role_id", []string{"roles.name"}, scanRoleName)
	records, err := userRoles.List(ctx, orm.QueryOptions{Where: "user_roles.user_id = $1", Args: []any{userID}})
	if err != nil {
		return nil, share.MapDBError(ctx, authRepoService, "accounts.load_roles", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}

	roles := make([]string, 0, len(records))
	for _, record := range records {
		roles = append(roles, record.Name)
	}

	return roles, nil
}

type SessionRepository struct {
	db                    database.Client
	sessions              *orm.Table[authSessionRecord]
	refreshTokens         *orm.Table[refreshTokenRecord]
	activeRefreshSessions *orm.Table[auth.Session]
	jobConfigs            *orm.Table[jobConfigRecord]
}

func NewSessionRepository(db database.Client) *SessionRepository {
	repository := &SessionRepository{db: db}
	if db != nil && db.Pool() != nil {
		repository.sessions = newAuthSessionTable(db.Pool())
		repository.refreshTokens = newRefreshTokenTable(db.Pool())
		repository.activeRefreshSessions = newActiveRefreshSessionTable(db.Pool())
		repository.jobConfigs = newJobConfigTable(db.Pool())
	}
	return repository
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

	sessions := newAuthSessionTable(tx)
	refreshTokens := newRefreshTokenTable(tx)

	if _, err := sessions.Insert(ctx, orm.Values{
		"session_id": session.SessionID,
		"subject":    session.Subject,
		"user_id":    session.UserID,
		"user_name":  session.Username,
	}); err != nil {
		return share.MapDBError(ctx, authRepoService, "sessions.insert_auth_session", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}

	if _, err := refreshTokens.Insert(ctx, orm.Values{
		"expires_at": time.Unix(session.RefreshExpiresAtUnix, 0).UTC(),
		"session_id": session.SessionID,
		"token_hash": session.RefreshTokenHash,
	}); err != nil {
		return share.MapDBError(ctx, authRepoService, "sessions.insert_refresh_token", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}

	if err := tx.Commit(ctx); err != nil {
		return share.MapDBError(ctx, authRepoService, "sessions.commit_tx", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}

	return nil
}

func (r *SessionRepository) IsSessionActive(ctx context.Context, sessionID string) (bool, error) {
	sessions, err := r.authSessionTable()
	if err != nil {
		return false, err
	}

	items, err := sessions.List(ctx, orm.QueryOptions{
		Where: "session_id = $1 AND is_revoked = FALSE",
		Args:  []any{sessionID},
		Limit: 1,
	})
	if err != nil {
		return false, share.MapDBError(ctx, authRepoService, "sessions.is_active", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}

	return len(items) > 0, nil
}

func (r *SessionRepository) FindActiveByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*auth.Session, error) {
	refreshTokens, err := r.activeRefreshSessionTable()
	if err != nil {
		return nil, err
	}

	session, err := refreshTokens.FindOne(ctx, "rt.token_hash = $1 AND rt.is_revoked = FALSE AND rt.expires_at > NOW() AND s.is_revoked = FALSE", refreshTokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, auth.ErrInvalidToken
		}
		return nil, share.MapDBError(ctx, authRepoService, "sessions.find_by_refresh_token_hash", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}

	return &session, nil
}

func (r *SessionRepository) RotateRefreshToken(ctx context.Context, session auth.Session, newRefreshTokenHash string, expiresAtUnix int64) error {
	refreshTokens, err := r.refreshTokenTable()
	if err != nil {
		return err
	}

	result, err := refreshTokens.UpdateWhere(ctx, "session_id = $4 AND token_hash = $5 AND is_revoked = FALSE", orm.Values{
		"expires_at": time.Unix(expiresAtUnix, 0).UTC(),
		"token_hash": newRefreshTokenHash,
		"updated_at": time.Now().UTC(),
	}, session.SessionID, session.RefreshTokenHash)
	if err != nil {
		return share.MapDBError(ctx, authRepoService, "sessions.rotate_refresh_token", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}
	if result.RowsAffected() == 0 {
		return auth.ErrInvalidToken
	}

	return nil
}

func (r *SessionRepository) RevokeBySessionID(ctx context.Context, sessionID string) error {
	sessions, err := r.authSessionTable()
	if err != nil {
		return err
	}

	result, err := sessions.UpdateWhere(ctx, "session_id = $3 AND is_revoked = FALSE", orm.Values{
		"is_revoked": true,
		"updated_at": time.Now().UTC(),
	}, sessionID)
	if err != nil {
		return share.MapDBError(ctx, authRepoService, "sessions.revoke_auth_session", err, auth.ErrDependencyUnavailable, auth.ErrInternal)
	}
	if result.RowsAffected() == 0 {
		return auth.ErrInvalidToken
	}

	refreshTokens, err := r.refreshTokenTable()
	if err != nil {
		return err
	}
	if _, err := refreshTokens.UpdateWhere(ctx, "session_id = $3 AND is_revoked = FALSE", orm.Values{
		"is_revoked": true,
		"updated_at": time.Now().UTC(),
	}, sessionID); err != nil {
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
	jobConfigs, err := r.jobConfigTable()
	if err != nil {
		return auth.SessionCleanupConfig{}, err
	}

	record, err := jobConfigs.FindOne(ctx, "job_name = $1", sessionCleanupJobName)
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

	return auth.SessionCleanupConfig{
		Enabled:   record.Enabled,
		Interval:  time.Duration(record.IntervalSeconds) * time.Second,
		Retention: time.Duration(record.RetentionSeconds) * time.Second,
	}, nil
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

type roleRecord struct {
	ID   uint64
	Name string
}

type userRoleRecord struct {
	UserID uint64
	RoleID uint64
}

type authSessionRecord struct {
	SessionID string
	UserID    uint64
	Username  string
	Subject   string
	IsRevoked bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type refreshTokenRecord struct {
	SessionID string
	TokenHash string
	IsRevoked bool
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type jobConfigRecord struct {
	Enabled          bool
	IntervalSeconds  int64
	RetentionSeconds int64
}

func (r *AccountRepository) userTable() (*orm.Table[auth.User], error) {
	if r == nil || r.db == nil || r.db.Pool() == nil {
		return nil, auth.ErrDependencyUnavailable
	}
	if r.users != nil {
		return r.users, nil
	}
	return newUserTable(r.db.Pool()), nil
}

func (r *SessionRepository) authSessionTable() (*orm.Table[authSessionRecord], error) {
	if r == nil || r.db == nil || r.db.Pool() == nil {
		return nil, auth.ErrDependencyUnavailable
	}
	if r.sessions != nil {
		return r.sessions, nil
	}
	return newAuthSessionTable(r.db.Pool()), nil
}

func (r *SessionRepository) refreshTokenTable() (*orm.Table[refreshTokenRecord], error) {
	if r == nil || r.db == nil || r.db.Pool() == nil {
		return nil, auth.ErrDependencyUnavailable
	}
	if r.refreshTokens != nil {
		return r.refreshTokens, nil
	}
	return newRefreshTokenTable(r.db.Pool()), nil
}

func (r *SessionRepository) activeRefreshSessionTable() (*orm.Table[auth.Session], error) {
	if r == nil || r.db == nil || r.db.Pool() == nil {
		return nil, auth.ErrDependencyUnavailable
	}
	if r.activeRefreshSessions != nil {
		return r.activeRefreshSessions, nil
	}
	return newActiveRefreshSessionTable(r.db.Pool()), nil
}

func (r *SessionRepository) jobConfigTable() (*orm.Table[jobConfigRecord], error) {
	if r == nil || r.db == nil || r.db.Pool() == nil {
		return nil, auth.ErrDependencyUnavailable
	}
	if r.jobConfigs != nil {
		return r.jobConfigs, nil
	}
	return newJobConfigTable(r.db.Pool()), nil
}

func newUserTable(db orm.Queryer) *orm.Table[auth.User] {
	return orm.NewTable(db, "users", []string{"id", "user_name", "email", "password_hash", "COALESCE(avatar_url, '')", "is_active", "created_at", "updated_at"}, scanUser)
}

func newRoleTable(db orm.Queryer) *orm.Table[roleRecord] {
	return orm.NewTable(db, "roles", []string{"id", "name"}, scanRole)
}

func newUserRoleTable(db orm.Queryer) *orm.Table[userRoleRecord] {
	return orm.NewTable(db, "user_roles", []string{"user_id", "role_id"}, scanUserRole)
}

func newAuthSessionTable(db orm.Queryer) *orm.Table[authSessionRecord] {
	return orm.NewTable(db, "auth_sessions", []string{"session_id", "user_id", "user_name", "subject", "is_revoked", "created_at", "updated_at"}, scanAuthSession)
}

func newRefreshTokenTable(db orm.Queryer) *orm.Table[refreshTokenRecord] {
	return orm.NewTable(db, "refresh_tokens", []string{"session_id", "token_hash", "is_revoked", "expires_at", "created_at", "updated_at"}, scanRefreshToken)
}

func newActiveRefreshSessionTable(db orm.Queryer) *orm.Table[auth.Session] {
	return orm.NewTable(
		db,
		"refresh_tokens rt JOIN auth_sessions s ON s.session_id = rt.session_id",
		[]string{"s.session_id", "s.user_id", "s.user_name", "s.subject", "rt.token_hash", "EXTRACT(EPOCH FROM rt.expires_at)::BIGINT"},
		scanRefreshSession,
	)
}

func newJobConfigTable(db orm.Queryer) *orm.Table[jobConfigRecord] {
	return orm.NewTable(db, "job_configs", []string{"enabled", "interval_seconds", "retention_seconds"}, scanJobConfig)
}

func scanUser(scanner orm.Scanner) (auth.User, error) {
	var (
		user            auth.User
		rawUsername     string
		rawEmail        string
		rawPasswordHash string
		rawAvatarURL    string
	)
	err := scanner.Scan(
		&user.ID,
		&rawUsername,
		&rawEmail,
		&rawPasswordHash,
		&rawAvatarURL,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return auth.User{}, err
	}
	user.Username, err = auth.NewUsername(rawUsername)
	if err != nil {
		return auth.User{}, err
	}
	user.Email, err = auth.NewEmail(rawEmail)
	if err != nil {
		return auth.User{}, err
	}
	user.PasswordHash, err = auth.NewPasswordHash(rawPasswordHash)
	if err != nil {
		return auth.User{}, err
	}
	if rawAvatarURL != "" {
		user.AvatarURL, err = auth.NewAvatarURL(rawAvatarURL)
		if err != nil {
			return auth.User{}, err
		}
	}
	return user, nil
}

func scanRole(scanner orm.Scanner) (roleRecord, error) {
	var record roleRecord
	err := scanner.Scan(&record.ID, &record.Name)
	return record, err
}

func scanUserRole(scanner orm.Scanner) (userRoleRecord, error) {
	var record userRoleRecord
	err := scanner.Scan(&record.UserID, &record.RoleID)
	return record, err
}

func scanRoleName(scanner orm.Scanner) (roleRecord, error) {
	var record roleRecord
	err := scanner.Scan(&record.Name)
	return record, err
}

func scanAuthSession(scanner orm.Scanner) (authSessionRecord, error) {
	var record authSessionRecord
	err := scanner.Scan(
		&record.SessionID,
		&record.UserID,
		&record.Username,
		&record.Subject,
		&record.IsRevoked,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	return record, err
}

func scanRefreshToken(scanner orm.Scanner) (refreshTokenRecord, error) {
	var record refreshTokenRecord
	err := scanner.Scan(
		&record.SessionID,
		&record.TokenHash,
		&record.IsRevoked,
		&record.ExpiresAt,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	return record, err
}

func scanRefreshSession(scanner orm.Scanner) (auth.Session, error) {
	var session auth.Session
	err := scanner.Scan(
		&session.SessionID,
		&session.UserID,
		&session.Username,
		&session.Subject,
		&session.RefreshTokenHash,
		&session.RefreshExpiresAtUnix,
	)
	return session, err
}

func scanJobConfig(scanner orm.Scanner) (jobConfigRecord, error) {
	var record jobConfigRecord
	err := scanner.Scan(&record.Enabled, &record.IntervalSeconds, &record.RetentionSeconds)
	return record, err
}

var _ auth.AccountRepository = (*AccountRepository)(nil)
var _ auth.SessionRepository = (*SessionRepository)(nil)
