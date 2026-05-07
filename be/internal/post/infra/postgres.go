package infra

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"falzo-be/internal/post"
	"falzo-be/internal/share"
	"falzo-be/pkg/database"
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
		RETURNING id, (SELECT user_name FROM users WHERE id = $1), created_at, updated_at
	`,
		item.UserID,
		item.ImageURL.String(),
		item.Caption.String(),
		item.LocationName.String(),
		item.Latitude,
		item.Longitude,
	).Scan(&item.ID, &item.UserName, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return share.MapDBError(ctx, postRepoService, "posts.insert", err, post.ErrDependencyUnavailable, post.ErrInternal)
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
		return share.MapDBError(ctx, postRepoService, "posts.like", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}

	return nil
}

func (r *PostgresRepository) Unlike(ctx context.Context, postID uint64, userID uint64) error {
	if r.db == nil || r.db.Pool() == nil {
		return post.ErrDependencyUnavailable
	}
	if err := r.ensurePostExists(ctx, postID); err != nil {
		return err
	}

	if _, err := r.db.Pool().Exec(ctx, `
		DELETE FROM post_likes
		WHERE post_id = $1 AND user_id = $2
	`, postID, userID); err != nil {
		return share.MapDBError(ctx, postRepoService, "posts.unlike", err, post.ErrDependencyUnavailable, post.ErrInternal)
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
		return share.MapDBError(ctx, postRepoService, "posts.save", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}

	return nil
}

func (r *PostgresRepository) Unsave(ctx context.Context, postID uint64, userID uint64) error {
	if r.db == nil || r.db.Pool() == nil {
		return post.ErrDependencyUnavailable
	}
	if err := r.ensurePostExists(ctx, postID); err != nil {
		return err
	}

	if _, err := r.db.Pool().Exec(ctx, `
		DELETE FROM post_saves
		WHERE post_id = $1 AND user_id = $2
	`, postID, userID); err != nil {
		return share.MapDBError(ctx, postRepoService, "posts.unsave", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}

	return nil
}

func (r *PostgresRepository) Comment(ctx context.Context, comment *post.Comment) error {
	if r.db == nil || r.db.Pool() == nil {
		return post.ErrDependencyUnavailable
	}
	if comment == nil {
		return post.ErrInternal
	}
	if err := r.ensurePostExists(ctx, comment.PostID); err != nil {
		return err
	}
	if comment.ReplyToCommentID != 0 {
		if err := r.ensureCommentBelongsToPost(ctx, comment.ReplyToCommentID, comment.PostID); err != nil {
			return err
		}
	}

	var (
		rawContent          string
		rawReplyToCommentID sql.NullInt64
		rawReplyToUserID    sql.NullInt64
		rawReplyToUserName  sql.NullString
		rawReplyToContent   sql.NullString
	)

	err := r.db.Pool().QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO post_comments (post_id, user_id, content, parent_comment_id)
			VALUES ($1, $2, $3, NULLIF($4, 0))
			RETURNING id, post_id, user_id, content, parent_comment_id, created_at, updated_at
		)
		SELECT inserted.id, inserted.post_id, inserted.user_id, users.user_name, inserted.content, inserted.created_at, inserted.updated_at,
			parent.id, parent.user_id, parent_user.user_name, parent.content
		FROM inserted
		INNER JOIN users ON users.id = inserted.user_id
		LEFT JOIN post_comments parent ON parent.id = inserted.parent_comment_id
		LEFT JOIN users parent_user ON parent_user.id = parent.user_id
	`, comment.PostID, comment.UserID, comment.Content.String(), comment.ReplyToCommentID).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.UserID,
		&comment.UserName,
		&rawContent,
		&comment.CreatedAt,
		&comment.UpdatedAt,
		&rawReplyToCommentID,
		&rawReplyToUserID,
		&rawReplyToUserName,
		&rawReplyToContent,
	)
	if err != nil {
		return share.MapDBError(ctx, postRepoService, "posts.comment", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}
	if err := applyScannedCommentContent(comment, rawContent); err != nil {
		return err
	}
	applyScannedReply(comment, rawReplyToCommentID, rawReplyToUserID, rawReplyToUserName, rawReplyToContent)

	return nil
}

func (r *PostgresRepository) UpdateComment(ctx context.Context, postID uint64, commentID uint64, userID uint64, content post.Content) (post.Comment, error) {
	if r.db == nil || r.db.Pool() == nil {
		return post.Comment{}, post.ErrDependencyUnavailable
	}
	if err := r.ensurePostExists(ctx, postID); err != nil {
		return post.Comment{}, err
	}

	var exists bool
	var ownerUserID uint64
	err := r.db.Pool().QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM post_comments
			WHERE id = $1 AND post_id = $2
		),
		COALESCE((
			SELECT user_id
			FROM post_comments
			WHERE id = $1 AND post_id = $2
			LIMIT 1
		), 0)
	`, commentID, postID).Scan(&exists, &ownerUserID)
	if err != nil {
		return post.Comment{}, share.MapDBError(ctx, postRepoService, "posts.comment_owner", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}
	if !exists {
		return post.Comment{}, post.ErrCommentNotFound
	}
	if ownerUserID != userID {
		return post.Comment{}, post.ErrCommentUpdateForbidden
	}

	var (
		comment             post.Comment
		rawContent          string
		rawReplyToCommentID sql.NullInt64
		rawReplyToUserID    sql.NullInt64
		rawReplyToUserName  sql.NullString
		rawReplyToContent   sql.NullString
	)
	err = r.db.Pool().QueryRow(ctx, `
		WITH updated AS (
			UPDATE post_comments
			SET content = $4, updated_at = CURRENT_TIMESTAMP
			WHERE id = $1 AND post_id = $2 AND user_id = $3
			RETURNING id, post_id, user_id, content, parent_comment_id, created_at, updated_at
		)
		SELECT updated.id, updated.post_id, updated.user_id, users.user_name, updated.content, updated.created_at, updated.updated_at,
			parent.id, parent.user_id, parent_user.user_name, parent.content
		FROM updated
		INNER JOIN users ON users.id = updated.user_id
		LEFT JOIN post_comments parent ON parent.id = updated.parent_comment_id
		LEFT JOIN users parent_user ON parent_user.id = parent.user_id
	`, commentID, postID, userID, content.String()).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.UserID,
		&comment.UserName,
		&rawContent,
		&comment.CreatedAt,
		&comment.UpdatedAt,
		&rawReplyToCommentID,
		&rawReplyToUserID,
		&rawReplyToUserName,
		&rawReplyToContent,
	)
	if err != nil {
		return post.Comment{}, share.MapDBError(ctx, postRepoService, "posts.update_comment", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}
	if err := applyScannedCommentContent(&comment, rawContent); err != nil {
		return post.Comment{}, err
	}
	applyScannedReply(&comment, rawReplyToCommentID, rawReplyToUserID, rawReplyToUserName, rawReplyToContent)

	return comment, nil
}

func (r *PostgresRepository) GetPosts(ctx context.Context, page int, limit int, viewerUserID uint64) ([]post.Post, error) {
	if r.db == nil || r.db.Pool() == nil {
		return nil, post.ErrDependencyUnavailable
	}

	offset := (page - 1) * limit
	rows, err := r.db.Pool().Query(ctx, `
		SELECT posts.id, posts.user_id, users.user_name, posts.image_url, posts.caption, posts.location_name,
				COALESCE(posts.latitude, 0), COALESCE(posts.longitude, 0),
				EXISTS (
					SELECT 1
					FROM post_likes
					WHERE post_likes.post_id = posts.id AND post_likes.user_id = $3
				),
				EXISTS (
					SELECT 1
					FROM post_saves
					WHERE post_saves.post_id = posts.id AND post_saves.user_id = $3
				),
				posts.created_at, posts.updated_at
		FROM posts
		INNER JOIN users ON users.id = posts.user_id
		ORDER BY posts.created_at DESC, posts.id DESC
		LIMIT $1 OFFSET $2
	`, limit, offset, viewerUserID)
	if err != nil {
		return nil, share.MapDBError(ctx, postRepoService, "posts.get_posts", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}
	defer rows.Close()

	posts := make([]post.Post, 0)
	for rows.Next() {
		item, err := scanPost(rows)
		if err != nil {
			return nil, share.MapDBError(ctx, postRepoService, "posts.get_posts.scan", err, post.ErrDependencyUnavailable, post.ErrInternal)
		}
		posts = append(posts, item)
	}
	if err := rows.Err(); err != nil {
		return nil, share.MapDBError(ctx, postRepoService, "posts.get_posts.iterate", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}

	return posts, nil
}

func (r *PostgresRepository) GetComments(ctx context.Context, postID uint64, page int, limit int) ([]post.Comment, error) {
	if r.db == nil || r.db.Pool() == nil {
		return nil, post.ErrDependencyUnavailable
	}
	if err := r.ensurePostExists(ctx, postID); err != nil {
		return nil, err
	}

	offset := (page - 1) * limit
	rows, err := r.db.Pool().Query(ctx, `
		SELECT post_comments.id, post_comments.post_id, post_comments.user_id, users.user_name, post_comments.content, post_comments.created_at,
			post_comments.updated_at,
			parent.id, parent.user_id, parent_user.user_name, parent.content
		FROM post_comments
		INNER JOIN users ON users.id = post_comments.user_id
		LEFT JOIN post_comments parent ON parent.id = post_comments.parent_comment_id
		LEFT JOIN users parent_user ON parent_user.id = parent.user_id
		WHERE post_comments.post_id = $1
		ORDER BY post_comments.created_at ASC, post_comments.id ASC
		LIMIT $2 OFFSET $3
	`, postID, limit, offset)
	if err != nil {
		return nil, share.MapDBError(ctx, postRepoService, "posts.get_comments", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}
	defer rows.Close()

	comments := make([]post.Comment, 0)
	for rows.Next() {
		var (
			item                post.Comment
			rawContent          string
			rawReplyToCommentID sql.NullInt64
			rawReplyToUserID    sql.NullInt64
			rawReplyToUserName  sql.NullString
			rawReplyToContent   sql.NullString
		)
		if err := rows.Scan(
			&item.ID,
			&item.PostID,
			&item.UserID,
			&item.UserName,
			&rawContent,
			&item.CreatedAt,
			&item.UpdatedAt,
			&rawReplyToCommentID,
			&rawReplyToUserID,
			&rawReplyToUserName,
			&rawReplyToContent,
		); err != nil {
			return nil, share.MapDBError(ctx, postRepoService, "posts.get_comments.scan", err, post.ErrDependencyUnavailable, post.ErrInternal)
		}
		if err := applyScannedCommentContent(&item, rawContent); err != nil {
			return nil, err
		}
		applyScannedReply(&item, rawReplyToCommentID, rawReplyToUserID, rawReplyToUserName, rawReplyToContent)
		comments = append(comments, item)
	}
	if err := rows.Err(); err != nil {
		return nil, share.MapDBError(ctx, postRepoService, "posts.get_comments.iterate", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}

	return comments, nil
}

func (r *PostgresRepository) GetPostDetail(ctx context.Context, postID uint64, viewerUserID uint64) (*post.Post, error) {
	if r.db == nil || r.db.Pool() == nil {
		return nil, post.ErrDependencyUnavailable
	}

	item, err := scanPost(r.db.Pool().QueryRow(ctx, `
		SELECT posts.id, posts.user_id, users.user_name, posts.image_url, posts.caption, posts.location_name,
				COALESCE(posts.latitude, 0), COALESCE(posts.longitude, 0),
				EXISTS (
					SELECT 1
					FROM post_likes
					WHERE post_likes.post_id = posts.id AND post_likes.user_id = $2
				),
				EXISTS (
					SELECT 1
					FROM post_saves
					WHERE post_saves.post_id = posts.id AND post_saves.user_id = $2
				),
				posts.created_at, posts.updated_at
		FROM posts
		INNER JOIN users ON users.id = posts.user_id
		WHERE posts.id = $1
		LIMIT 1
	`, postID, viewerUserID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, post.ErrNotFound
		}
		return nil, share.MapDBError(ctx, postRepoService, "posts.get_post_detail", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}

	return &item, nil
}

func (r *PostgresRepository) GetPostsByLocation(ctx context.Context, locationName post.LocationName) ([]post.Post, error) {
	if r.db == nil || r.db.Pool() == nil {
		return nil, post.ErrDependencyUnavailable
	}

	pattern := "%" + strings.TrimSpace(locationName.String()) + "%"
	rows, err := r.db.Pool().Query(ctx, `
		SELECT posts.id, posts.user_id, users.user_name, posts.image_url, posts.caption, posts.location_name,
				COALESCE(posts.latitude, 0), COALESCE(posts.longitude, 0),
				false,
				false,
				posts.created_at, posts.updated_at
		FROM posts
		INNER JOIN users ON users.id = posts.user_id
		WHERE posts.location_name ILIKE $1
		ORDER BY posts.created_at DESC, posts.id DESC
		LIMIT 100
	`, pattern)
	if err != nil {
		return nil, share.MapDBError(ctx, postRepoService, "posts.get_posts_by_location", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}
	defer rows.Close()

	posts := make([]post.Post, 0)
	for rows.Next() {
		item, err := scanPost(rows)
		if err != nil {
			return nil, share.MapDBError(ctx, postRepoService, "posts.get_posts_by_location.scan", err, post.ErrDependencyUnavailable, post.ErrInternal)
		}
		posts = append(posts, item)
	}
	if err := rows.Err(); err != nil {
		return nil, share.MapDBError(ctx, postRepoService, "posts.get_posts_by_location.iterate", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}

	return posts, nil
}

func (r *PostgresRepository) ensureCommentBelongsToPost(ctx context.Context, commentID uint64, postID uint64) error {
	var exists bool
	err := r.db.Pool().QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM post_comments
			WHERE id = $1 AND post_id = $2
		)
	`, commentID, postID).Scan(&exists)
	if err != nil {
		return share.MapDBError(ctx, postRepoService, "posts.comment_reply_exists", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}
	if !exists {
		return post.ErrReplyCommentNotFound
	}

	return nil
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
		return share.MapDBError(ctx, postRepoService, "posts.exists", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}
	if !exists {
		return post.ErrNotFound
	}

	return nil
}

func applyScannedCommentContent(comment *post.Comment, rawContent string) error {
	content, err := post.NewContent(rawContent)
	if err != nil {
		return err
	}

	comment.Content = content
	return nil
}

func applyScannedReply(
	comment *post.Comment,
	rawReplyToCommentID sql.NullInt64,
	rawReplyToUserID sql.NullInt64,
	rawReplyToUserName sql.NullString,
	rawReplyToContent sql.NullString,
) {
	if rawReplyToCommentID.Valid && rawReplyToCommentID.Int64 > 0 {
		comment.ReplyToCommentID = uint64(rawReplyToCommentID.Int64)
	}
	if rawReplyToUserID.Valid && rawReplyToUserID.Int64 > 0 {
		comment.ReplyToUserID = uint64(rawReplyToUserID.Int64)
	}
	if rawReplyToUserName.Valid {
		comment.ReplyToUserName = rawReplyToUserName.String
	}
	if rawReplyToContent.Valid {
		comment.ReplyToContent = rawReplyToContent.String
	}
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
		&item.UserName,
		&rawImageURL,
		&rawCaption,
		&rawLocationName,
		&item.Latitude,
		&item.Longitude,
		&item.IsLiked,
		&item.IsSaved,
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

var _ post.Repository = (*PostgresRepository)(nil)
