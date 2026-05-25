package infra

import (
	"context"
	"errors"
	"strconv"
	"time"

	"falzo-be/internal/share"
	"falzo-be/internal/upload"
	"falzo-be/pkg/database"
	"falzo-be/pkg/database/orm"

	"github.com/jackc/pgx/v5"
)

const uploadRepoService = "upload"

type PostgresRepository struct {
	db     database.Client
	images *orm.Table[upload.Image]
}

func NewPostgresRepository(db database.Client) *PostgresRepository {
	repository := &PostgresRepository{db: db}
	if db != nil && db.Pool() != nil {
		repository.images = newImageTable(db.Pool())
	}
	return repository
}

func (r *PostgresRepository) Save(ctx context.Context, image *upload.Image) error {
	if image == nil {
		return upload.ErrInternal
	}
	table, err := r.table()
	if err != nil {
		return err
	}

	err = table.InsertReturning(ctx, orm.Values{
		"mime_type":  image.MimeType,
		"object_key": image.ObjectKey,
		"owner_id":   image.OwnerID,
		"size":       image.Size,
		"status":     image.Status,
		"url":        image.URL,
	}, []string{"id", "created_at", "updated_at"}, &image.ID, &image.CreatedAt, &image.UpdatedAt)
	if err != nil {
		return share.MapDBError(ctx, uploadRepoService, "images.insert", err, upload.ErrDependencyUnavailable, upload.ErrInternal)
	}

	return nil
}

func (r *PostgresRepository) FindByID(ctx context.Context, id int64) (*upload.Image, error) {
	table, err := r.table()
	if err != nil {
		return nil, err
	}

	image, err := table.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, upload.ErrImageNotFound
		}
		return nil, share.MapDBError(ctx, uploadRepoService, "images.find_by_id", err, upload.ErrDependencyUnavailable, upload.ErrInternal)
	}

	return &image, nil
}

func (r *PostgresRepository) Update(ctx context.Context, image *upload.Image) error {
	if image == nil {
		return upload.ErrInternal
	}
	table, err := r.table()
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	result, err := table.UpdateWhere(ctx, "id = $7 AND owner_id::text = $8", orm.Values{
		"mime_type":  image.MimeType,
		"object_key": image.ObjectKey,
		"size":       image.Size,
		"status":     image.Status,
		"updated_at": now,
		"url":        image.URL,
	}, image.ID, image.OwnerID)
	if err != nil {
		return share.MapDBError(ctx, uploadRepoService, "images.update", err, upload.ErrDependencyUnavailable, upload.ErrInternal)
	}
	if result.RowsAffected() == 0 {
		return upload.ErrImageNotFound
	}

	return nil
}

func (r *PostgresRepository) table() (*orm.Table[upload.Image], error) {
	if r == nil || r.db == nil || r.db.Pool() == nil {
		return nil, upload.ErrDependencyUnavailable
	}
	if r.images != nil {
		return r.images, nil
	}
	return newImageTable(r.db.Pool()), nil
}

func newImageTable(db orm.Queryer) *orm.Table[upload.Image] {
	return orm.NewTable(db, "images", []string{"id", "owner_id", "object_key", "url", "mime_type", "size", "status", "created_at", "updated_at"}, scanImage)
}

func scanImage(scanner orm.Scanner) (upload.Image, error) {
	var image upload.Image
	var ownerID int64
	err := scanner.Scan(
		&image.ID,
		&ownerID,
		&image.ObjectKey,
		&image.URL,
		&image.MimeType,
		&image.Size,
		&image.Status,
		&image.CreatedAt,
		&image.UpdatedAt,
	)
	if err != nil {
		return upload.Image{}, err
	}
	image.OwnerID = strconv.FormatInt(ownerID, 10)
	return image, nil
}

var _ upload.ImageRepository = (*PostgresRepository)(nil)
