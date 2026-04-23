package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"falzo-be/internal/post/domain"
	"falzo-be/internal/post/domain/aggregate"
	"falzo-be/internal/post/domain/entity"
	"falzo-be/internal/post/domain/repository"
	"falzo-be/internal/post/domain/valueobject"
	"falzo-be/pkg/database"
)

type PostRepository struct {
	db database.Client
}

const postRepoService = "post"

func NewPostRepository(db database.Client) repository.PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(ctx context.Context, post *aggregate.Post) error {
	if r.db == nil || r.db.DB() == nil {
		return domain.ErrPostDependencyUnavailable
	}
	if post == nil {
		return domain.ErrPostInternal
	}

	tx, err := r.db.DB().BeginTx(ctx, nil)
	if err != nil {
		return mapDBError(ctx, postRepoService, "posts.begin_tx", err)
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, `
		INSERT INTO posts (user_id, image_url, caption, location_name, latitude, longitude)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`,
		post.Post.UserID,
		post.Post.ImageURL.String(),
		post.Post.Caption.String(),
		post.Post.LocationName.String(),
		post.Post.Latitude,
		post.Post.Longitude,
	).Scan(&post.Post.ID, &post.Post.CreatedAt, &post.Post.UpdatedAt)
	if err != nil {
		return mapDBError(ctx, postRepoService, "posts.insert", err)
	}

	if err := tx.Commit(); err != nil {
		return mapDBError(ctx, postRepoService, "posts.commit_tx", err)
	}

	return nil
}

func (r *PostRepository) Like(ctx context.Context, postID uint64, userID uint64) error {
	if r.db == nil || r.db.DB() == nil {
		return domain.ErrPostDependencyUnavailable
	}

	if err := r.ensurePostExists(ctx, postID); err != nil {
		return err
	}

	if _, err := r.db.DB().ExecContext(ctx, `
		INSERT INTO post_likes (post_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (post_id, user_id) DO NOTHING
	`, postID, userID); err != nil {
		return mapDBError(ctx, postRepoService, "posts.like", err)
	}

	return nil
}

func (r *PostRepository) Save(ctx context.Context, postID uint64, userID uint64) error {
	if r.db == nil || r.db.DB() == nil {
		return domain.ErrPostDependencyUnavailable
	}

	if err := r.ensurePostExists(ctx, postID); err != nil {
		return err
	}

	if _, err := r.db.DB().ExecContext(ctx, `
		INSERT INTO post_saves (post_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (post_id, user_id) DO NOTHING
	`, postID, userID); err != nil {
		return mapDBError(ctx, postRepoService, "posts.save", err)
	}

	return nil
}

func (r *PostRepository) GetPosts(ctx context.Context, page int, limit int) ([]entity.Post, error) {
	if r.db == nil || r.db.DB() == nil {
		return nil, domain.ErrPostDependencyUnavailable
	}

	offset := (page - 1) * limit
	rows, err := r.db.DB().QueryContext(ctx, `
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

	posts := make([]entity.Post, 0)
	for rows.Next() {
		post, err := scanPost(rows)
		if err != nil {
			return nil, mapDBError(ctx, postRepoService, "posts.get_posts.scan", err)
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, mapDBError(ctx, postRepoService, "posts.get_posts.iterate", err)
	}

	return posts, nil
}

func (r *PostRepository) GetPostDetail(ctx context.Context, postID uint64) (*entity.Post, error) {
	if r.db == nil || r.db.DB() == nil {
		return nil, domain.ErrPostDependencyUnavailable
	}

	row := r.db.DB().QueryRowContext(ctx, `
		SELECT id, user_id, image_url, caption, location_name,
		       COALESCE(latitude, 0), COALESCE(longitude, 0),
		       created_at, updated_at
		FROM posts
		WHERE id = $1
		LIMIT 1
	`, postID)

	post, err := scanPost(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPostNotFound
		}
		return nil, mapDBError(ctx, postRepoService, "posts.get_post_detail", err)
	}

	return &post, nil
}

func (r *PostRepository) GetPostsByLocation(ctx context.Context, locationName valueobject.LocationName) ([]entity.Post, error) {
	if r.db == nil || r.db.DB() == nil {
		return nil, domain.ErrPostDependencyUnavailable
	}

	pattern := "%" + strings.TrimSpace(locationName.String()) + "%"
	rows, err := r.db.DB().QueryContext(ctx, `
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

	posts := make([]entity.Post, 0)
	for rows.Next() {
		post, err := scanPost(rows)
		if err != nil {
			return nil, mapDBError(ctx, postRepoService, "posts.get_posts_by_location.scan", err)
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, mapDBError(ctx, postRepoService, "posts.get_posts_by_location.iterate", err)
	}

	return posts, nil
}

func (r *PostRepository) ensurePostExists(ctx context.Context, postID uint64) error {
	var exists bool
	err := r.db.DB().QueryRowContext(ctx, `
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
		return domain.ErrPostNotFound
	}

	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPost(scanner rowScanner) (entity.Post, error) {
	var (
		post            entity.Post
		rawImageURL     string
		rawCaption      string
		rawLocationName string
	)

	err := scanner.Scan(
		&post.ID,
		&post.UserID,
		&rawImageURL,
		&rawCaption,
		&rawLocationName,
		&post.Latitude,
		&post.Longitude,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return entity.Post{}, err
	}

	imageURL, err := valueobject.NewImageURL(rawImageURL)
	if err != nil {
		return entity.Post{}, err
	}
	caption, err := valueobject.NewCaption(rawCaption)
	if err != nil {
		return entity.Post{}, err
	}
	locationName, err := valueobject.NewLocationName(rawLocationName)
	if err != nil {
		return entity.Post{}, err
	}

	post.ImageURL = imageURL
	post.Caption = caption
	post.LocationName = locationName
	if post.CreatedAt.IsZero() {
		post.CreatedAt = time.Now().UTC()
	}

	return post, nil
}
