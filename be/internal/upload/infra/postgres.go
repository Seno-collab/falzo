package infra

import (
	"context"
	"errors"

	"falzo-be/internal/share"
	"falzo-be/internal/upload"
	"falzo-be/pkg/database"

	"github.com/jackc/pgx/v5"
)

const uploadRepoService = "upload"

type PostgresRepository struct {
	db database.Client
}

func NewPostgresRepository(db database.Client) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Save(ctx context.Context, image *upload.Image) error {
	if r.db == nil || r.db.Pool() == nil {
		return upload.ErrDependencyUnavailable
	}
	if image == nil {
		return upload.ErrInternal
	}

	err := r.db.Pool().QueryRow(ctx, `
		INSERT INTO images (owner_id, object_key, url, mime_type, size, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`,
		image.OwnerID,
		image.ObjectKey,
		image.URL,
		image.MimeType,
		image.Size,
		image.Status,
	).Scan(&image.ID, &image.CreatedAt, &image.UpdatedAt)
	if err != nil {
		return share.MapDBError(ctx, uploadRepoService, "images.insert", err, upload.ErrDependencyUnavailable, upload.ErrInternal)
	}

	return nil
}

func (r *PostgresRepository) FindByID(ctx context.Context, id int64) (*upload.Image, error) {
	if r.db == nil || r.db.Pool() == nil {
		return nil, upload.ErrDependencyUnavailable
	}

	image := upload.Image{}
	err := r.db.Pool().QueryRow(ctx, `
		SELECT id, owner_id::text, object_key, url, mime_type, size, status, created_at, updated_at
		FROM images
		WHERE id = $1
		LIMIT 1
	`, id).Scan(
		&image.ID,
		&image.OwnerID,
		&image.ObjectKey,
		&image.URL,
		&image.MimeType,
		&image.Size,
		&image.Status,
		&image.CreatedAt,
		&image.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, upload.ErrImageNotFound
		}
		return nil, share.MapDBError(ctx, uploadRepoService, "images.find_by_id", err, upload.ErrDependencyUnavailable, upload.ErrInternal)
	}

	return &image, nil
}

func (r *PostgresRepository) Update(ctx context.Context, image *upload.Image) error {
	if r.db == nil || r.db.Pool() == nil {
		return upload.ErrDependencyUnavailable
	}
	if image == nil {
		return upload.ErrInternal
	}

	result, err := r.db.Pool().Exec(ctx, `
		UPDATE images
		SET object_key = $2,
		    url = $3,
		    mime_type = $4,
		    size = $5,
		    status = $6,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`,
		image.ID,
		image.ObjectKey,
		image.URL,
		image.MimeType,
		image.Size,
		image.Status,
	)
	if err != nil {
		return share.MapDBError(ctx, uploadRepoService, "images.update", err, upload.ErrDependencyUnavailable, upload.ErrInternal)
	}
	if result.RowsAffected() == 0 {
		return upload.ErrImageNotFound
	}

	return nil
}

var _ upload.ImageRepository = (*PostgresRepository)(nil)
