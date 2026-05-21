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
	"falzo-be/pkg/dberr"

	"github.com/jackc/pgx/v5/pgconn"
)

type PostgresRepository struct {
	db database.Client
}

const postRepoService = "post"

func NewPostgresRepository(db database.Client) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func postSelectSQL(viewerParam string) string {
	return `
		SELECT posts.id, posts.user_id, users.user_name, COALESCE(users.avatar_url, ''), posts.image_url, posts.caption, posts.location_name,
				COALESCE(categories.id, 0), COALESCE(categories.name, ''), COALESCE(categories.slug, ''),
				COALESCE(posts.latitude, 0), COALESCE(posts.longitude, 0),
				EXISTS (
					SELECT 1
					FROM post_likes
					WHERE post_likes.post_id = posts.id AND post_likes.user_id = ` + viewerParam + `
				),
				EXISTS (
					SELECT 1
					FROM post_saves
					WHERE post_saves.post_id = posts.id AND post_saves.user_id = ` + viewerParam + `
				),
				posts.status,
				(SELECT COUNT(*) FROM post_likes WHERE post_likes.post_id = posts.id),
				(SELECT COUNT(*) FROM post_comments WHERE post_comments.post_id = posts.id AND post_comments.deleted_at IS NULL AND post_comments.status = 'visible'),
				(SELECT COUNT(*) FROM post_saves WHERE post_saves.post_id = posts.id),
				posts.created_at, posts.updated_at
		FROM posts
		INNER JOIN users ON users.id = posts.user_id
		LEFT JOIN categories ON categories.id = posts.category_id
	`
}

func (r *PostgresRepository) loadPost(ctx context.Context, postID uint64, viewerUserID uint64) (post.Post, error) {
	item, err := scanPost(r.db.Pool().QueryRow(ctx, postSelectSQL("$2")+`
		WHERE posts.id = $1
		LIMIT 1
	`, postID, viewerUserID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return post.Post{}, post.ErrNotFound
		}
		return post.Post{}, share.MapDBError(ctx, postRepoService, "posts.load", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}

	return item, nil
}

func (r *PostgresRepository) Create(ctx context.Context, item *post.Post) error {
	if r.db == nil || r.db.Pool() == nil {
		return post.ErrDependencyUnavailable
	}
	if item == nil {
		return post.ErrInternal
	}
	if item.CategoryID != 0 {
		if err := r.ensureCategoryExists(ctx, item.CategoryID); err != nil {
			return err
		}
	}

	err := r.db.Pool().QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO posts (user_id, image_url, caption, location_name, latitude, longitude, category_id)
			VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, 0))
			RETURNING id, category_id, created_at, updated_at
		)
		SELECT inserted.id, users.user_name, COALESCE(users.avatar_url, ''),
			COALESCE(categories.id, 0),
			COALESCE(categories.name, ''),
			COALESCE(categories.slug, ''),
			inserted.created_at, inserted.updated_at
		FROM inserted
		INNER JOIN users ON users.id = $1
		LEFT JOIN categories ON categories.id = inserted.category_id
	`,
		item.UserID,
		item.ImageURL.String(),
		item.Caption.String(),
		item.LocationName.String(),
		item.Latitude,
		item.Longitude,
		item.CategoryID,
	).Scan(&item.ID, &item.UserName, &item.UserAvatarURL, &item.CategoryID, &item.CategoryName, &item.CategorySlug, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return share.MapDBError(ctx, postRepoService, "posts.insert", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}

	return nil
}

func (r *PostgresRepository) UpdatePost(ctx context.Context, postID uint64, userID uint64, update post.PostUpdate) (post.Post, error) {
	if r.db == nil || r.db.Pool() == nil {
		return post.Post{}, post.ErrDependencyUnavailable
	}
	if update.CategoryID != 0 {
		if err := r.ensureCategoryExists(ctx, update.CategoryID); err != nil {
			return post.Post{}, err
		}
	}
	if err := r.ensurePostExists(ctx, postID); err != nil {
		return post.Post{}, err
	}

	item, err := scanPost(r.db.Pool().QueryRow(ctx, postSelectSQL("$2")+`
		WHERE posts.id = $1
			AND posts.user_id = $2
			AND posts.deleted_at IS NULL
			AND posts.status = 'visible'
		LIMIT 1
	`, postID, userID))
	if err == nil {
		_, err = r.db.Pool().Exec(ctx, `
			UPDATE posts
			SET caption = $3,
				location_name = $4,
				latitude = $5,
				longitude = $6,
				category_id = NULLIF($7, 0),
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		`, postID, userID, update.Caption.String(), update.LocationName.String(), update.Latitude, update.Longitude, update.CategoryID)
		if err != nil {
			return post.Post{}, share.MapDBError(ctx, postRepoService, "posts.update", err, post.ErrDependencyUnavailable, post.ErrInternal)
		}

		return r.loadPost(ctx, postID, userID)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return post.Post{}, post.ErrPostUpdateForbidden
	}

	return item, share.MapDBError(ctx, postRepoService, "posts.update_owner", err, post.ErrDependencyUnavailable, post.ErrInternal)
}

func (r *PostgresRepository) DeletePost(ctx context.Context, postID uint64, actor post.ModerationActor) error {
	return r.moderatePost(ctx, postID, actor, "", true)
}

func (r *PostgresRepository) HidePost(ctx context.Context, postID uint64, actor post.ModerationActor, reason post.ReportReason) error {
	return r.moderatePost(ctx, postID, actor, reason.String(), false)
}

func (r *PostgresRepository) ReportPost(ctx context.Context, report post.ContentReport) error {
	if r.db == nil || r.db.Pool() == nil {
		return post.ErrDependencyUnavailable
	}
	if err := r.ensurePostExists(ctx, report.PostID); err != nil {
		return err
	}

	if _, err := r.db.Pool().Exec(ctx, `
		INSERT INTO content_reports (reporter_user_id, post_id, reason)
		VALUES ($1, $2, $3)
	`, report.ReporterUserID, report.PostID, report.Reason.String()); err != nil {
		return share.MapDBError(ctx, postRepoService, "content_reports.post_insert", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}

	return nil
}

func (r *PostgresRepository) ReportComment(ctx context.Context, report post.ContentReport) error {
	if r.db == nil || r.db.Pool() == nil {
		return post.ErrDependencyUnavailable
	}
	if err := r.ensureCommentBelongsToPost(ctx, report.CommentID, report.PostID); err != nil {
		return err
	}

	if _, err := r.db.Pool().Exec(ctx, `
		INSERT INTO content_reports (reporter_user_id, post_id, comment_id, reason)
		VALUES ($1, $2, $3, $4)
	`, report.ReporterUserID, report.PostID, report.CommentID, report.Reason.String()); err != nil {
		return share.MapDBError(ctx, postRepoService, "content_reports.comment_insert", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}

	return nil
}

func (r *PostgresRepository) DeleteComment(ctx context.Context, postID uint64, commentID uint64, actor post.ModerationActor) error {
	return r.moderateComment(ctx, postID, commentID, actor, "", true)
}

func (r *PostgresRepository) HideComment(ctx context.Context, postID uint64, commentID uint64, actor post.ModerationActor, reason post.ReportReason) error {
	return r.moderateComment(ctx, postID, commentID, actor, reason.String(), false)
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

func (r *PostgresRepository) CreateSavedCollection(ctx context.Context, collection *post.SavedCollection) error {
	if r.db == nil || r.db.Pool() == nil {
		return post.ErrDependencyUnavailable
	}
	if collection == nil {
		return post.ErrInternal
	}

	err := r.db.Pool().QueryRow(ctx, `
		INSERT INTO saved_collections (user_id, name, share_slug, is_public)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`, collection.UserID, collection.Name.String(), collection.ShareSlug, collection.IsPublic).Scan(&collection.ID, &collection.CreatedAt, &collection.UpdatedAt)
	if err != nil {
		if dberr.IsUniqueViolation(err) {
			return post.ErrCollectionNameTaken
		}
		return share.MapDBError(ctx, postRepoService, "saved_collections.insert", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}

	return nil
}

func (r *PostgresRepository) ListSavedCollections(ctx context.Context, userID uint64) ([]post.SavedCollection, error) {
	if r.db == nil || r.db.Pool() == nil {
		return nil, post.ErrDependencyUnavailable
	}

	rows, err := r.db.Pool().Query(ctx, `
		SELECT id, user_id, name, share_slug, is_public, created_at, updated_at
		FROM saved_collections
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
	`, userID)
	if err != nil {
		return nil, share.MapDBError(ctx, postRepoService, "saved_collections.list", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}
	defer rows.Close()

	collections := make([]post.SavedCollection, 0)
	for rows.Next() {
		var item post.SavedCollection
		var rawName string
		if err := rows.Scan(&item.ID, &item.UserID, &rawName, &item.ShareSlug, &item.IsPublic, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, share.MapDBError(ctx, postRepoService, "saved_collections.list.scan", err, post.ErrDependencyUnavailable, post.ErrInternal)
		}
		name, err := post.NewSavedCollectionName(rawName)
		if err != nil {
			return nil, err
		}
		item.Name = name

		posts, err := r.listSavedCollectionPosts(ctx, item.ID, userID, userID)
		if err != nil {
			return nil, err
		}
		item.Posts = posts
		collections = append(collections, item)
	}
	if err := rows.Err(); err != nil {
		return nil, share.MapDBError(ctx, postRepoService, "saved_collections.list.iterate", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}

	return collections, nil
}

func (r *PostgresRepository) ListSavedPosts(ctx context.Context, userID uint64) ([]post.Post, error) {
	if r.db == nil || r.db.Pool() == nil {
		return nil, post.ErrDependencyUnavailable
	}

	rows, err := r.db.Pool().Query(ctx, `
		SELECT posts.id, posts.user_id, users.user_name, COALESCE(users.avatar_url, ''), posts.image_url, posts.caption, posts.location_name,
				COALESCE(categories.id, 0), COALESCE(categories.name, ''), COALESCE(categories.slug, ''),
				COALESCE(posts.latitude, 0), COALESCE(posts.longitude, 0),
				EXISTS (
					SELECT 1
					FROM post_likes
					WHERE post_likes.post_id = posts.id AND post_likes.user_id = $1
				),
				true,
				posts.status,
				(SELECT COUNT(*) FROM post_likes WHERE post_likes.post_id = posts.id),
				(SELECT COUNT(*) FROM post_comments WHERE post_comments.post_id = posts.id AND post_comments.deleted_at IS NULL AND post_comments.status = 'visible'),
				(SELECT COUNT(*) FROM post_saves WHERE post_saves.post_id = posts.id),
				posts.created_at, posts.updated_at
		FROM post_saves
		INNER JOIN posts ON posts.id = post_saves.post_id
		INNER JOIN users ON users.id = posts.user_id
		LEFT JOIN categories ON categories.id = posts.category_id
		WHERE post_saves.user_id = $1
			AND posts.deleted_at IS NULL
			AND posts.status = 'visible'
		ORDER BY post_saves.created_at DESC, post_saves.id DESC
		LIMIT 200
	`, userID)
	if err != nil {
		return nil, share.MapDBError(ctx, postRepoService, "saved_posts.list", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}
	defer rows.Close()

	posts := make([]post.Post, 0)
	for rows.Next() {
		item, err := scanPost(rows)
		if err != nil {
			return nil, share.MapDBError(ctx, postRepoService, "saved_posts.list.scan", err, post.ErrDependencyUnavailable, post.ErrInternal)
		}
		posts = append(posts, item)
	}
	if err := rows.Err(); err != nil {
		return nil, share.MapDBError(ctx, postRepoService, "saved_posts.list.iterate", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}

	return posts, nil
}

func (r *PostgresRepository) AddPostToSavedCollection(ctx context.Context, collectionID uint64, postID uint64, userID uint64) error {
	if r.db == nil || r.db.Pool() == nil {
		return post.ErrDependencyUnavailable
	}
	if err := r.ensureSavedCollectionOwner(ctx, collectionID, userID); err != nil {
		return err
	}
	if err := r.ensurePostExists(ctx, postID); err != nil {
		return err
	}

	tx, err := r.db.Pool().Begin(ctx)
	if err != nil {
		return share.MapDBError(ctx, postRepoService, "saved_collection_posts.begin", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO post_saves (post_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (post_id, user_id) DO NOTHING
	`, postID, userID); err != nil {
		return share.MapDBError(ctx, postRepoService, "saved_collection_posts.ensure_save", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO saved_collection_posts (collection_id, post_id, user_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (collection_id, post_id) DO NOTHING
	`, collectionID, postID, userID); err != nil {
		return share.MapDBError(ctx, postRepoService, "saved_collection_posts.insert", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}

	if err := tx.Commit(ctx); err != nil {
		return share.MapDBError(ctx, postRepoService, "saved_collection_posts.commit", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}

	return nil
}

func (r *PostgresRepository) RemovePostFromSavedCollection(ctx context.Context, collectionID uint64, postID uint64, userID uint64) error {
	if r.db == nil || r.db.Pool() == nil {
		return post.ErrDependencyUnavailable
	}
	if err := r.ensureSavedCollectionOwner(ctx, collectionID, userID); err != nil {
		return err
	}

	if _, err := r.db.Pool().Exec(ctx, `
		DELETE FROM saved_collection_posts
		WHERE collection_id = $1 AND post_id = $2 AND user_id = $3
	`, collectionID, postID, userID); err != nil {
		return share.MapDBError(ctx, postRepoService, "saved_collection_posts.delete", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}

	return nil
}

func (r *PostgresRepository) DeleteSavedCollection(ctx context.Context, collectionID uint64, userID uint64) error {
	if r.db == nil || r.db.Pool() == nil {
		return post.ErrDependencyUnavailable
	}
	if err := r.ensureSavedCollectionOwner(ctx, collectionID, userID); err != nil {
		return err
	}

	if _, err := r.db.Pool().Exec(ctx, `
		DELETE FROM saved_collections
		WHERE id = $1 AND user_id = $2
	`, collectionID, userID); err != nil {
		return share.MapDBError(ctx, postRepoService, "saved_collections.delete", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}

	return nil
}

func (r *PostgresRepository) UpdateSavedCollectionVisibility(ctx context.Context, collectionID uint64, userID uint64, isPublic bool) (post.SavedCollection, error) {
	if r.db == nil || r.db.Pool() == nil {
		return post.SavedCollection{}, post.ErrDependencyUnavailable
	}
	if err := r.ensureSavedCollectionOwner(ctx, collectionID, userID); err != nil {
		return post.SavedCollection{}, err
	}

	item, err := r.loadSavedCollection(ctx, userID, `
		UPDATE saved_collections
		SET is_public = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, name, share_slug, is_public, created_at, updated_at
	`, collectionID, userID, isPublic)
	if err != nil {
		return post.SavedCollection{}, err
	}

	return item, nil
}

func (r *PostgresRepository) GetPublicSavedCollection(ctx context.Context, shareSlug string, viewerUserID uint64) (*post.SavedCollection, error) {
	if r.db == nil || r.db.Pool() == nil {
		return nil, post.ErrDependencyUnavailable
	}

	item, err := r.loadSavedCollection(ctx, viewerUserID, `
		SELECT id, user_id, name, share_slug, is_public, created_at, updated_at
		FROM saved_collections
		WHERE share_slug = $1 AND is_public = TRUE
		LIMIT 1
	`, strings.TrimSpace(shareSlug))
	if err != nil {
		return nil, err
	}

	return &item, nil
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
			WHERE id = $1 AND post_id = $2 AND deleted_at IS NULL
		),
		COALESCE((
			SELECT user_id
			FROM post_comments
			WHERE id = $1 AND post_id = $2 AND deleted_at IS NULL
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
			WHERE id = $1 AND post_id = $2 AND user_id = $3 AND deleted_at IS NULL AND status = 'visible'
			RETURNING id, post_id, user_id, content, parent_comment_id, created_at, updated_at, status
		)
		SELECT updated.id, updated.post_id, updated.user_id, users.user_name, updated.content, updated.created_at, updated.updated_at, updated.status,
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
		&comment.Status,
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

func (r *PostgresRepository) GetPosts(ctx context.Context, filter post.PostListFilter) ([]post.Post, error) {
	if r.db == nil || r.db.Pool() == nil {
		return nil, post.ErrDependencyUnavailable
	}

	offset := filter.Offset
	hasCursor := filter.Cursor != nil
	if hasCursor {
		offset = 0
	}
	var cursorCreatedAt time.Time
	var cursorID uint64
	var cursorRank float64
	if filter.Cursor != nil {
		cursorCreatedAt = filter.Cursor.CreatedAt
		cursorID = filter.Cursor.ID
		cursorRank = filter.Cursor.Rank
	}
	rankAt := filter.RankAt
	if rankAt.IsZero() {
		rankAt = time.Now().UTC()
	}

	searchPattern := "%" + strings.TrimSpace(filter.Search) + "%"
	categorySlug := strings.TrimSpace(filter.CategorySlug)
	sort := strings.TrimSpace(filter.Sort)
	radiusDegrees := float64(filter.RadiusMeters) / 111000
	if radiusDegrees <= 0 {
		radiusDegrees = 50.0 / 111.0
	}

	rows, err := r.db.Pool().Query(ctx, `
			WITH ranked_posts AS (
				SELECT posts.id, posts.user_id, users.user_name, COALESCE(users.avatar_url, '') AS user_avatar_url, posts.image_url, posts.caption, posts.location_name,
					COALESCE(categories.id, 0) AS category_id,
					COALESCE(categories.name, '') AS category_name,
					COALESCE(categories.slug, '') AS category_slug,
					COALESCE(posts.latitude, 0) AS latitude,
					COALESCE(posts.longitude, 0) AS longitude,
					EXISTS (
						SELECT 1
						FROM post_likes
						WHERE post_likes.post_id = posts.id AND post_likes.user_id = $3
					) AS is_liked,
					EXISTS (
						SELECT 1
						FROM post_saves
						WHERE post_saves.post_id = posts.id AND post_saves.user_id = $3
					) AS is_saved,
					posts.status,
					(SELECT COUNT(*) FROM post_likes WHERE post_likes.post_id = posts.id) AS likes_count,
					(SELECT COUNT(*) FROM post_comments WHERE post_comments.post_id = posts.id AND post_comments.deleted_at IS NULL AND post_comments.status = 'visible') AS comments_count,
					(SELECT COUNT(*) FROM post_saves WHERE post_saves.post_id = posts.id) AS saves_count,
					posts.created_at,
					posts.updated_at,
					CASE
						WHEN $7 = 'nearby' THEN (
							((COALESCE(posts.latitude, 0)::double precision - $8::double precision) * (COALESCE(posts.latitude, 0)::double precision - $8::double precision))
							+ ((COALESCE(posts.longitude, 0)::double precision - $9::double precision) * (COALESCE(posts.longitude, 0)::double precision - $9::double precision))
						)
						WHEN $7 = 'popular' THEN (
							(SELECT COUNT(*) FROM post_likes WHERE post_likes.post_id = posts.id) * 3
							+ (SELECT COUNT(*) FROM post_comments WHERE post_comments.post_id = posts.id AND post_comments.deleted_at IS NULL AND post_comments.status = 'visible') * 2
							+ (SELECT COUNT(*) FROM post_saves WHERE post_saves.post_id = posts.id)
						)::double precision
						WHEN $7 = 'trending' THEN (
							(SELECT COUNT(*) FROM post_likes WHERE post_likes.post_id = posts.id) * 3
							+ (SELECT COUNT(*) FROM post_comments WHERE post_comments.post_id = posts.id AND post_comments.deleted_at IS NULL AND post_comments.status = 'visible') * 2
							+ (SELECT COUNT(*) FROM post_saves WHERE post_saves.post_id = posts.id)
						)::double precision / (1 + EXTRACT(EPOCH FROM ($15::timestamptz - posts.created_at)) / 86400)
						ELSE 0
					END AS rank_value
				FROM posts
				INNER JOIN users ON users.id = posts.user_id
				LEFT JOIN categories ON categories.id = posts.category_id
				WHERE ($4 = '%%' OR posts.caption ILIKE $4 OR posts.location_name ILIKE $4 OR users.user_name ILIKE $4
					OR categories.name ILIKE $4 OR categories.slug ILIKE $4)
					AND ($5 = '' OR categories.slug = $5)
					AND posts.deleted_at IS NULL
					AND posts.status = 'visible'
					AND (
						$6 = ''
						OR (
							$6 = 'following'
							AND EXISTS (
								SELECT 1
								FROM user_follows
								WHERE user_follows.follower_id = $3
									AND user_follows.following_id = posts.user_id
							)
						)
					)
					AND (
						$7 <> 'nearby'
						OR (
							((COALESCE(posts.latitude, 0)::double precision - $8::double precision) * (COALESCE(posts.latitude, 0)::double precision - $8::double precision))
							+ ((COALESCE(posts.longitude, 0)::double precision - $9::double precision) * (COALESCE(posts.longitude, 0)::double precision - $9::double precision))
						) <= ($10::double precision * $10::double precision)
					)
					AND (
						$3 = 0
						OR NOT EXISTS (
							SELECT 1
							FROM user_blocks
							WHERE (blocker_user_id = $3 AND blocked_user_id = posts.user_id)
								OR (blocker_user_id = posts.user_id AND blocked_user_id = $3)
						)
					)
			)
			SELECT id, user_id, user_name, user_avatar_url, image_url, caption, location_name,
				category_id, category_name, category_slug, latitude, longitude,
				is_liked, is_saved, status, likes_count, comments_count, saves_count,
				created_at, updated_at, rank_value
			FROM ranked_posts
			WHERE (
				$14 = false
				OR (
					$7 = 'nearby'
					AND (
						rank_value > $13::double precision
						OR (rank_value = $13::double precision AND (created_at, id) < ($11::timestamptz, $12::bigint))
					)
				)
				OR (
					$7 IN ('popular', 'trending')
					AND (
						rank_value < $13::double precision
						OR (rank_value = $13::double precision AND (created_at, id) < ($11::timestamptz, $12::bigint))
					)
				)
				OR (
					$7 NOT IN ('nearby', 'popular', 'trending')
					AND (created_at, id) < ($11::timestamptz, $12::bigint)
				)
			)
			ORDER BY
				CASE WHEN $7 = 'nearby' THEN rank_value END ASC,
				CASE WHEN $7 IN ('popular', 'trending') THEN rank_value END DESC,
				created_at DESC, id DESC
			LIMIT $1 OFFSET $2
		`, filter.Limit, offset, filter.ViewerUserID, searchPattern, categorySlug, strings.TrimSpace(filter.Feed), sort, filter.Latitude, filter.Longitude, radiusDegrees, cursorCreatedAt, cursorID, cursorRank, hasCursor, rankAt)
	if err != nil {
		return nil, share.MapDBError(ctx, postRepoService, "posts.get_posts", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}
	defer rows.Close()

	posts := make([]post.Post, 0)
	for rows.Next() {
		item, err := scanPostWithRank(rows)
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
			post_comments.status,
			parent.id, parent.user_id, parent_user.user_name, parent.content
		FROM post_comments
		INNER JOIN users ON users.id = post_comments.user_id
		LEFT JOIN post_comments parent ON parent.id = post_comments.parent_comment_id
		LEFT JOIN users parent_user ON parent_user.id = parent.user_id
		WHERE post_comments.post_id = $1
			AND post_comments.deleted_at IS NULL
			AND post_comments.status = 'visible'
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
			&item.Status,
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

	item, err := scanPost(r.db.Pool().QueryRow(ctx, postSelectSQL("$2")+`
		WHERE posts.id = $1
			AND posts.deleted_at IS NULL
			AND posts.status = 'visible'
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
	rows, err := r.db.Pool().Query(ctx, postSelectSQL("0")+`
		WHERE posts.location_name ILIKE $1
			AND posts.deleted_at IS NULL
			AND posts.status = 'visible'
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

func (r *PostgresRepository) moderatePost(ctx context.Context, postID uint64, actor post.ModerationActor, reason string, deleted bool) error {
	if r.db == nil || r.db.Pool() == nil {
		return post.ErrDependencyUnavailable
	}
	if actor.UserID == 0 {
		return post.ErrUserIDRequired
	}

	var result pgconn.CommandTag
	var err error
	if deleted {
		result, err = r.db.Pool().Exec(ctx, `
			UPDATE posts
			SET status = 'deleted',
				deleted_at = CURRENT_TIMESTAMP,
				deleted_by = $2,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $1
				AND deleted_at IS NULL
				AND ($3 OR user_id = $2)
		`, postID, actor.UserID, actor.IsAdmin)
	} else {
		result, err = r.db.Pool().Exec(ctx, `
			UPDATE posts
			SET status = 'hidden',
				hidden_at = CURRENT_TIMESTAMP,
				hidden_by = $2,
				moderation_reason = $4,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $1
				AND deleted_at IS NULL
				AND ($3 OR user_id = $2)
		`, postID, actor.UserID, actor.IsAdmin, reason)
	}
	if err != nil {
		return share.MapDBError(ctx, postRepoService, "posts.moderate", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}
	if result.RowsAffected() == 0 {
		if err := r.ensurePostExists(ctx, postID); err != nil {
			return err
		}
		return post.ErrPostModerationForbidden
	}

	return nil
}

func (r *PostgresRepository) moderateComment(ctx context.Context, postID uint64, commentID uint64, actor post.ModerationActor, reason string, deleted bool) error {
	if r.db == nil || r.db.Pool() == nil {
		return post.ErrDependencyUnavailable
	}
	if actor.UserID == 0 {
		return post.ErrUserIDRequired
	}

	var result pgconn.CommandTag
	var err error
	if deleted {
		result, err = r.db.Pool().Exec(ctx, `
			UPDATE post_comments
			SET status = 'deleted',
				deleted_at = CURRENT_TIMESTAMP,
				deleted_by = $3,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $2
				AND post_id = $1
				AND deleted_at IS NULL
				AND (
					$4
					OR user_id = $3
					OR EXISTS (
						SELECT 1
						FROM posts
						WHERE posts.id = post_comments.post_id AND posts.user_id = $3
					)
				)
		`, postID, commentID, actor.UserID, actor.IsAdmin)
	} else {
		result, err = r.db.Pool().Exec(ctx, `
			UPDATE post_comments
			SET status = 'hidden',
				hidden_at = CURRENT_TIMESTAMP,
				hidden_by = $3,
				moderation_reason = $5,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $2
				AND post_id = $1
				AND deleted_at IS NULL
				AND (
					$4
					OR user_id = $3
					OR EXISTS (
						SELECT 1
						FROM posts
						WHERE posts.id = post_comments.post_id AND posts.user_id = $3
					)
				)
		`, postID, commentID, actor.UserID, actor.IsAdmin, reason)
	}
	if err != nil {
		return share.MapDBError(ctx, postRepoService, "comments.moderate", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}
	if result.RowsAffected() == 0 {
		if err := r.ensureCommentBelongsToPost(ctx, commentID, postID); err != nil {
			return err
		}
		return post.ErrPostModerationForbidden
	}

	return nil
}

func (r *PostgresRepository) ensureCategoryExists(ctx context.Context, categoryID uint64) error {
	var exists bool
	err := r.db.Pool().QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM categories
			WHERE id = $1
		)
	`, categoryID).Scan(&exists)
	if err != nil {
		return share.MapDBError(ctx, postRepoService, "posts.category_exists", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}
	if !exists {
		return post.ErrCategoryNotFound
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

func (r *PostgresRepository) ensureSavedCollectionOwner(ctx context.Context, collectionID uint64, userID uint64) error {
	var exists bool
	err := r.db.Pool().QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM saved_collections
			WHERE id = $1 AND user_id = $2
		)
	`, collectionID, userID).Scan(&exists)
	if err != nil {
		return share.MapDBError(ctx, postRepoService, "saved_collections.owner", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}
	if !exists {
		return post.ErrCollectionNotFound
	}

	return nil
}

func (r *PostgresRepository) loadSavedCollection(ctx context.Context, viewerUserID uint64, query string, args ...any) (post.SavedCollection, error) {
	var item post.SavedCollection
	var rawName string
	if err := r.db.Pool().QueryRow(ctx, query, args...).Scan(
		&item.ID,
		&item.UserID,
		&rawName,
		&item.ShareSlug,
		&item.IsPublic,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return post.SavedCollection{}, post.ErrCollectionNotFound
		}
		return post.SavedCollection{}, share.MapDBError(ctx, postRepoService, "saved_collections.load", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}

	name, err := post.NewSavedCollectionName(rawName)
	if err != nil {
		return post.SavedCollection{}, err
	}
	item.Name = name

	posts, err := r.listSavedCollectionPosts(ctx, item.ID, item.UserID, viewerUserID)
	if err != nil {
		return post.SavedCollection{}, err
	}
	item.Posts = posts

	return item, nil
}

func (r *PostgresRepository) listSavedCollectionPosts(ctx context.Context, collectionID uint64, ownerUserID uint64, viewerUserID uint64) ([]post.Post, error) {
	rows, err := r.db.Pool().Query(ctx, `
		SELECT posts.id, posts.user_id, users.user_name, COALESCE(users.avatar_url, ''), posts.image_url, posts.caption, posts.location_name,
				COALESCE(categories.id, 0), COALESCE(categories.name, ''), COALESCE(categories.slug, ''),
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
				posts.status,
				(SELECT COUNT(*) FROM post_likes WHERE post_likes.post_id = posts.id),
				(SELECT COUNT(*) FROM post_comments WHERE post_comments.post_id = posts.id AND post_comments.deleted_at IS NULL AND post_comments.status = 'visible'),
				(SELECT COUNT(*) FROM post_saves WHERE post_saves.post_id = posts.id),
				posts.created_at, posts.updated_at
		FROM saved_collection_posts
		INNER JOIN posts ON posts.id = saved_collection_posts.post_id
		INNER JOIN users ON users.id = posts.user_id
		LEFT JOIN categories ON categories.id = posts.category_id
		WHERE saved_collection_posts.collection_id = $1
			AND saved_collection_posts.user_id = $2
			AND posts.deleted_at IS NULL
			AND posts.status = 'visible'
		ORDER BY saved_collection_posts.created_at DESC, saved_collection_posts.id DESC
		LIMIT 200
	`, collectionID, ownerUserID, viewerUserID)
	if err != nil {
		return nil, share.MapDBError(ctx, postRepoService, "saved_collection_posts.list", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}
	defer rows.Close()

	posts := make([]post.Post, 0)
	for rows.Next() {
		item, err := scanPost(rows)
		if err != nil {
			return nil, share.MapDBError(ctx, postRepoService, "saved_collection_posts.list.scan", err, post.ErrDependencyUnavailable, post.ErrInternal)
		}
		posts = append(posts, item)
	}
	if err := rows.Err(); err != nil {
		return nil, share.MapDBError(ctx, postRepoService, "saved_collection_posts.list.iterate", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}

	return posts, nil
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
		&item.UserAvatarURL,
		&rawImageURL,
		&rawCaption,
		&rawLocationName,
		&item.CategoryID,
		&item.CategoryName,
		&item.CategorySlug,
		&item.Latitude,
		&item.Longitude,
		&item.IsLiked,
		&item.IsSaved,
		&item.Status,
		&item.LikesCount,
		&item.CommentsCount,
		&item.SavesCount,
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

func scanPostWithRank(scanner rowScanner) (post.Post, error) {
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
		&item.UserAvatarURL,
		&rawImageURL,
		&rawCaption,
		&rawLocationName,
		&item.CategoryID,
		&item.CategoryName,
		&item.CategorySlug,
		&item.Latitude,
		&item.Longitude,
		&item.IsLiked,
		&item.IsSaved,
		&item.Status,
		&item.LikesCount,
		&item.CommentsCount,
		&item.SavesCount,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.CursorRank,
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
