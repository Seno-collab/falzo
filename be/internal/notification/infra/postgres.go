package infra

import (
	"context"
	"database/sql"
	"strings"

	"falzo-be/internal/notification"
	"falzo-be/internal/share"
	"falzo-be/pkg/database"
	"falzo-be/pkg/database/orm"
)

const notificationRepoService = "notification"

type PostgresRepository struct {
	db            database.Client
	notifications *orm.Table[notification.Notification]
}

func NewPostgresRepository(db database.Client) *PostgresRepository {
	repository := &PostgresRepository{db: db}
	if db != nil && db.Pool() != nil {
		repository.notifications = newNotificationTable(db.Pool())
	}
	return repository
}

func (r *PostgresRepository) Save(ctx context.Context, item notification.Notification) error {
	table, err := r.table()
	if err != nil {
		return err
	}

	_, err = table.InsertWithOptions(ctx, orm.Values{
		"actor_name":    item.ActorName,
		"actor_user_id": nullableUint64(item.ActorUserID),
		"body":          item.Body,
		"created_at":    item.CreatedAt,
		"id":            item.ID,
		"image_id":      nullableInt64(item.ImageID),
		"post_id":       nullableUint64(item.PostID),
		"resource":      item.Resource,
		"resource_id":   item.ResourceID,
		"title":         item.Title,
		"type":          item.Type,
		"user_id":       item.UserID,
	}, orm.InsertOptions{OnConflictDoNothing: true})
	if err != nil {
		return share.MapDBError(ctx, notificationRepoService, "notifications.insert", err, notification.ErrDependencyUnavailable, notification.ErrInternal)
	}

	return nil
}

func (r *PostgresRepository) ListByUser(ctx context.Context, userID uint64, limit int) ([]notification.Notification, error) {
	if userID == 0 {
		return nil, notification.ErrUserIDRequired
	}
	if limit <= 0 {
		return nil, notification.ErrInvalidLimit
	}
	table, err := r.table()
	if err != nil {
		return nil, err
	}
	items, err := table.List(ctx, orm.QueryOptions{
		Where:   "user_id = $1",
		Args:    []any{userID},
		OrderBy: `"created_at" DESC, "id" DESC`,
		Limit:   limit,
	})
	if err != nil {
		return nil, share.MapDBError(ctx, notificationRepoService, "notifications.rows", err, notification.ErrDependencyUnavailable, notification.ErrInternal)
	}

	return items, nil
}

func (r *PostgresRepository) MarkRead(ctx context.Context, userID uint64, ids []string) error {
	if r.db == nil || r.db.Pool() == nil {
		return notification.ErrDependencyUnavailable
	}
	if userID == 0 {
		return notification.ErrUserIDRequired
	}

	cleanIDs := make([]string, 0, len(ids))
	for _, id := range ids {
		value := strings.TrimSpace(id)
		if value != "" {
			cleanIDs = append(cleanIDs, value)
		}
	}
	if len(cleanIDs) == 0 {
		return nil
	}

	if _, err := r.db.Pool().Exec(ctx, `
		UPDATE notifications
		SET read_at = COALESCE(read_at, CURRENT_TIMESTAMP)
		WHERE user_id = $1 AND id = ANY($2)
	`, userID, cleanIDs); err != nil {
		return share.MapDBError(ctx, notificationRepoService, "notifications.mark_read", err, notification.ErrDependencyUnavailable, notification.ErrInternal)
	}

	return nil
}

func (r *PostgresRepository) table() (*orm.Table[notification.Notification], error) {
	if r == nil || r.db == nil || r.db.Pool() == nil {
		return nil, notification.ErrDependencyUnavailable
	}
	if r.notifications != nil {
		return r.notifications, nil
	}
	return newNotificationTable(r.db.Pool()), nil
}

func newNotificationTable(db orm.Queryer) *orm.Table[notification.Notification] {
	return orm.NewTable(
		db,
		"notifications",
		[]string{"id", "user_id", "actor_user_id", "actor_name", "type", "title", "body", "resource", "resource_id", "post_id", "image_id", "created_at", "read_at"},
		scanNotification,
	)
}

func scanNotification(scanner orm.Scanner) (notification.Notification, error) {
	var item notification.Notification
	var actorUserID sql.NullInt64
	var postID sql.NullInt64
	var imageID sql.NullInt64
	var readAt sql.NullTime
	err := scanner.Scan(
		&item.ID,
		&item.UserID,
		&actorUserID,
		&item.ActorName,
		&item.Type,
		&item.Title,
		&item.Body,
		&item.Resource,
		&item.ResourceID,
		&postID,
		&imageID,
		&item.CreatedAt,
		&readAt,
	)
	if err != nil {
		return notification.Notification{}, err
	}
	if actorUserID.Valid && actorUserID.Int64 > 0 {
		item.ActorUserID = uint64(actorUserID.Int64)
	}
	if postID.Valid && postID.Int64 > 0 {
		item.PostID = uint64(postID.Int64)
	}
	if imageID.Valid && imageID.Int64 > 0 {
		item.ImageID = imageID.Int64
	}
	if readAt.Valid {
		item.ReadAt = &readAt.Time
	}
	return item, nil
}

func nullableUint64(value uint64) any {
	if value == 0 {
		return nil
	}
	return value
}

func nullableInt64(value int64) any {
	if value == 0 {
		return nil
	}
	return value
}

var _ notification.Repository = (*PostgresRepository)(nil)
