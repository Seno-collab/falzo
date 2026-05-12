package infra

import (
	"context"

	"falzo-be/internal/notification"
	"falzo-be/internal/share"
	"falzo-be/pkg/database"
)

const notificationRepoService = "notification"

type PostgresRepository struct {
	db database.Client
}

func NewPostgresRepository(db database.Client) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Save(ctx context.Context, item notification.Notification) error {
	if r.db == nil || r.db.Pool() == nil {
		return notification.ErrDependencyUnavailable
	}

	_, err := r.db.Pool().Exec(ctx, `
		INSERT INTO notifications (
			id, user_id, actor_user_id, actor_name, type, title, body,
			resource, resource_id, post_id, image_id, created_at
		)
		VALUES ($1, $2, NULLIF($3::BIGINT, 0), $4, $5, $6, $7, $8, $9, NULLIF($10::BIGINT, 0), NULLIF($11::BIGINT, 0), $12)
		ON CONFLICT (id) DO NOTHING
	`,
		item.ID,
		item.UserID,
		item.ActorUserID,
		item.ActorName,
		item.Type,
		item.Title,
		item.Body,
		item.Resource,
		item.ResourceID,
		item.PostID,
		item.ImageID,
		item.CreatedAt,
	)
	if err != nil {
		return share.MapDBError(ctx, notificationRepoService, "notifications.insert", err, notification.ErrDependencyUnavailable, notification.ErrInternal)
	}

	return nil
}

func (r *PostgresRepository) ListByUser(ctx context.Context, userID uint64, limit int) ([]notification.Notification, error) {
	if r.db == nil || r.db.Pool() == nil {
		return nil, notification.ErrDependencyUnavailable
	}
	if userID == 0 {
		return nil, notification.ErrUserIDRequired
	}
	if limit <= 0 {
		return nil, notification.ErrInvalidLimit
	}

	rows, err := r.db.Pool().Query(ctx, `
		SELECT id, user_id, COALESCE(actor_user_id, 0), actor_name, type, title, body,
			resource, resource_id, COALESCE(post_id, 0), COALESCE(image_id, 0), created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, share.MapDBError(ctx, notificationRepoService, "notifications.list_by_user", err, notification.ErrDependencyUnavailable, notification.ErrInternal)
	}
	defer rows.Close()

	items := make([]notification.Notification, 0, limit)
	for rows.Next() {
		var item notification.Notification
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.ActorUserID,
			&item.ActorName,
			&item.Type,
			&item.Title,
			&item.Body,
			&item.Resource,
			&item.ResourceID,
			&item.PostID,
			&item.ImageID,
			&item.CreatedAt,
		); err != nil {
			return nil, share.MapDBError(ctx, notificationRepoService, "notifications.scan", err, notification.ErrDependencyUnavailable, notification.ErrInternal)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, share.MapDBError(ctx, notificationRepoService, "notifications.rows", err, notification.ErrDependencyUnavailable, notification.ErrInternal)
	}

	return items, nil
}

var _ notification.Repository = (*PostgresRepository)(nil)
