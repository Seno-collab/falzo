package infra

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"falzo-be/internal/post"
	"falzo-be/pkg/database"
	"falzo-be/pkg/dberr"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type PostgresRepository struct {
	db database.Client
}

const postRepoService = "post"

func NewPostgresRepository(db database.Client) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, item *post.Post) error {
	if r.db == nil || r.db.Pool() == nil {
		return post.ErrDependencyUnavailable
	}
	if item == nil {
		return post.ErrInternal
	}

	err := r.db.Pool().QueryRow(ctx, `
		INSERT INTO posts (user_id, image_url, caption, location_name, latitude, longitude)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`,
		item.UserID,
		item.ImageURL.String(),
		item.Caption.String(),
		item.LocationName.String(),
		item.Latitude,
		item.Longitude,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return mapDBError(ctx, postRepoService, "posts.insert", err)
	}

	return nil
}

func (r *PostgresRepository) Like(ctx context.Context, postID uint64, userID uint64) error {
	if r.db == nil || r.db.Pool() == nil {
		return post.ErrDependencyUnavailable
	}
	if err := r.ensurePostExists(ctx, postID); err != nil {
		return err
	}

	if _, err := r.db.Pool().Exec(ctx, `
		INSERT INTO post_likes (post_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (post_id, user_id) DO NOTHING
	`, postID, userID); err != nil {
		return mapDBError(ctx, postRepoService, "posts.like", err)
	}

	return nil
}

func (r *PostgresRepository) Save(ctx context.Context, postID uint64, userID uint64) error {
	if r.db == nil || r.db.Pool() == nil {
		return post.ErrDependencyUnavailable
	}
	if err := r.ensurePostExists(ctx, postID); err != nil {
		return err
	}

	if _, err := r.db.Pool().Exec(ctx, `
		INSERT INTO post_saves (post_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (post_id, user_id) DO NOTHING
	`, postID, userID); err != nil {
		return mapDBError(ctx, postRepoService, "posts.save", err)
	}

	return nil
}

func (r *PostgresRepository) GetPosts(ctx context.Context, page int, limit int) ([]post.Post, error) {
	if r.db == nil || r.db.Pool() == nil {
		return nil, post.ErrDependencyUnavailable
	}

	offset := (page - 1) * limit
	rows, err := r.db.Pool().Query(ctx, `
		SELECT id, user_id, image_url, caption, location_name,
				COALESCE(latitude, 0), COALESCE(longitude, 0),
				created_at, updated_at
		FROM posts
		ORDER BY created_at DESC, id DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, mapDBError(ctx, postRepoService, "posts.get_posts", err)
	}
	defer rows.Close()

	posts := make([]post.Post, 0)
	for rows.Next() {
		item, err := scanPost(rows)
		if err != nil {
			return nil, mapDBError(ctx, postRepoService, "posts.get_posts.scan", err)
		}
		posts = append(posts, item)
	}
	if err := rows.Err(); err != nil {
		return nil, mapDBError(ctx, postRepoService, "posts.get_posts.iterate", err)
	}

	return posts, nil
}

func (r *PostgresRepository) GetPostDetail(ctx context.Context, postID uint64) (*post.Post, error) {
	if r.db == nil || r.db.Pool() == nil {
		return nil, post.ErrDependencyUnavailable
	}

	item, err := scanPost(r.db.Pool().QueryRow(ctx, `
		SELECT id, user_id, image_url, caption, location_name,
				COALESCE(latitude, 0), COALESCE(longitude, 0),
				created_at, updated_at
		FROM posts
		WHERE id = $1
		LIMIT 1
	`, postID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, post.ErrNotFound
		}
		return nil, mapDBError(ctx, postRepoService, "posts.get_post_detail", err)
	}

	return &item, nil
}

func (r *PostgresRepository) GetPostsByLocation(ctx context.Context, locationName post.LocationName) ([]post.Post, error) {
	if r.db == nil || r.db.Pool() == nil {
		return nil, post.ErrDependencyUnavailable
	}

	pattern := "%" + strings.TrimSpace(locationName.String()) + "%"
	rows, err := r.db.Pool().Query(ctx, `
		SELECT id, user_id, image_url, caption, location_name,
				COALESCE(latitude, 0), COALESCE(longitude, 0),
				created_at, updated_at
		FROM posts
		WHERE location_name ILIKE $1
		ORDER BY created_at DESC, id DESC
		LIMIT 100
	`, pattern)
	if err != nil {
		return nil, mapDBError(ctx, postRepoService, "posts.get_posts_by_location", err)
	}
	defer rows.Close()

	posts := make([]post.Post, 0)
	for rows.Next() {
		item, err := scanPost(rows)
		if err != nil {
			return nil, mapDBError(ctx, postRepoService, "posts.get_posts_by_location.scan", err)
		}
		posts = append(posts, item)
	}
	if err := rows.Err(); err != nil {
		return nil, mapDBError(ctx, postRepoService, "posts.get_posts_by_location.iterate", err)
	}

	return posts, nil
}

func (r *PostgresRepository) ensurePostExists(ctx context.Context, postID uint64) error {
	var exists bool
	err := r.db.Pool().QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM posts
			WHERE id = $1
		)
	`, postID).Scan(&exists)
	if err != nil {
		return mapDBError(ctx, postRepoService, "posts.exists", err)
	}
	if !exists {
		return post.ErrNotFound
	}

	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPost(scanner rowScanner) (post.Post, error) {
	var (
		item            post.Post
		rawImageURL     string
		rawCaption      string
		rawLocationName string
	)

	err := scanner.Scan(
		&item.ID,
		&item.UserID,
		&rawImageURL,
		&rawCaption,
		&rawLocationName,
		&item.Latitude,
		&item.Longitude,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return post.Post{}, err
	}

	item.ImageURL, err = post.NewImageURL(rawImageURL)
	if err != nil {
		return post.Post{}, err
	}
	item.Caption, err = post.NewCaption(rawCaption)
	if err != nil {
		return post.Post{}, err
	}
	item.LocationName, err = post.NewLocationName(rawLocationName)
	if err != nil {
		return post.Post{}, err
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now().UTC()
	}

	return item, nil
}

func mapDBError(ctx context.Context, service, operation string, err error) error {
	return dberr.MapDependencyOrInternal(
		err,
		service,
		operation,
		chimiddleware.GetReqID(ctx),
		post.ErrDependencyUnavailable,
		post.ErrInternal,
	)
}

var _ post.Repository = (*PostgresRepository)(nil)
