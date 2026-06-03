package infra

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"falzo-be/internal/i18n"
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

func postCategoriesJSONSQL(localeParam string) string {
	return `
	COALESCE((
		SELECT jsonb_agg(
			jsonb_build_object(
				'id', categories.id,
				'name', COALESCE(category_locale.name, categories.name),
				'slug', categories.slug
			)
			ORDER BY post_categories.created_at, post_categories.category_id
		)
		FROM post_categories
		INNER JOIN categories ON categories.id = post_categories.category_id
		LEFT JOIN category_translations category_locale
			ON category_locale.category_id = categories.id
			AND category_locale.locale = ` + localeParam + `
		WHERE post_categories.post_id = posts.id
	), '[]'::jsonb)`
}

const postImageURLsJSONSQL = `COALESCE(NULLIF(posts.image_urls, '[]'::jsonb), jsonb_build_array(posts.image_url))`

func NewPostgresRepository(db database.Client) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func postSelectSQL(viewerParam string, localeParam string) string {
	return `
		SELECT posts.id, posts.user_id, users.user_name, COALESCE(users.avatar_url, ''), posts.image_url,
				` + postImageURLsJSONSQL + `::text, posts.caption, posts.location_name,
				COALESCE(categories.id, 0), COALESCE(category_locale.name, categories.name, ''), COALESCE(categories.slug, ''),
				` + postCategoriesJSONSQL(localeParam) + `::text,
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
				(SELECT COUNT(*) FROM post_trust_votes WHERE post_trust_votes.post_id = posts.id AND post_trust_votes.vote_type = 'credible'),
				(SELECT COUNT(*) FROM post_trust_votes WHERE post_trust_votes.post_id = posts.id AND post_trust_votes.vote_type = 'suspicious'),
				(SELECT COUNT(*) FROM post_trust_votes WHERE post_trust_votes.post_id = posts.id AND post_trust_votes.vote_type = 'ai_generated'),
				(SELECT COUNT(*) FROM post_trust_votes WHERE post_trust_votes.post_id = posts.id AND post_trust_votes.vote_type = 'wrong_context'),
				(SELECT COUNT(*) FROM post_trust_votes WHERE post_trust_votes.post_id = posts.id AND post_trust_votes.vote_type = 'unsure'),
				COALESCE((
					SELECT post_trust_votes.vote_type
					FROM post_trust_votes
					WHERE post_trust_votes.post_id = posts.id AND post_trust_votes.user_id = ` + viewerParam + `
					LIMIT 1
				), ''),
				posts.created_at, posts.updated_at
		FROM posts
		INNER JOIN users ON users.id = posts.user_id
		LEFT JOIN categories ON categories.id = posts.category_id
		LEFT JOIN category_translations category_locale
			ON category_locale.category_id = categories.id
			AND category_locale.locale = ` + localeParam + `
	`
}

func (r *PostgresRepository) loadPost(ctx context.Context, postID uint64, viewerUserID uint64) (post.Post, error) {
	item, err := scanPost(r.db.Pool().QueryRow(ctx, postSelectSQL("$2", "$3")+`
		WHERE posts.id = $1
		LIMIT 1
	`, postID, viewerUserID, i18n.LocaleFromContext(ctx)))
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
	categoryIDs, err := post.NormalizeCategoryIDs(item.CategoryID, item.CategoryIDs)
	if err != nil {
		return err
	}
	if len(categoryIDs) > 0 {
		if err := r.ensureCategoriesExist(ctx, categoryIDs); err != nil {
			return err
		}
	}

	tx, err := r.db.Pool().Begin(ctx)
	if err != nil {
		return share.MapDBError(ctx, postRepoService, "posts.insert.begin", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}
	defer tx.Rollback(ctx)

	imageURLsJSON, err := marshalImageURLs(item.ImageURL, item.ImageURLs)
	if err != nil {
		return err
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO posts (user_id, image_url, image_urls, caption, location_name, latitude, longitude, category_id)
		VALUES ($1, $2, $3::jsonb, $4, $5, $6, $7, NULLIF($8, 0))
		RETURNING id, created_at, updated_at
	`,
		item.UserID,
		item.ImageURL.String(),
		imageURLsJSON,
		item.Caption.String(),
		item.LocationName.String(),
		item.Latitude,
		item.Longitude,
		firstCategoryID(categoryIDs),
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return share.MapDBError(ctx, postRepoService, "posts.insert", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}

	if err := r.replacePostCategories(ctx, tx, item.ID, categoryIDs); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return share.MapDBError(ctx, postRepoService, "posts.insert.commit", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}

	loaded, err := r.loadPost(ctx, item.ID, item.UserID)
	if err != nil {
		return err
	}
	*item = loaded

	return nil
}

func (r *PostgresRepository) UpdatePost(ctx context.Context, postID uint64, userID uint64, update post.PostUpdate) (post.Post, error) {
	if r.db == nil || r.db.Pool() == nil {
		return post.Post{}, post.ErrDependencyUnavailable
	}
	categoryIDs, err := post.NormalizeCategoryIDs(update.CategoryID, update.CategoryIDs)
	if err != nil {
		return post.Post{}, err
	}
	if len(categoryIDs) > 0 {
		if err := r.ensureCategoriesExist(ctx, categoryIDs); err != nil {
			return post.Post{}, err
		}
	}
	if err := r.ensurePostExists(ctx, postID); err != nil {
		return post.Post{}, err
	}

	item, err := scanPost(r.db.Pool().QueryRow(ctx, postSelectSQL("$2", "$3")+`
		WHERE posts.id = $1
			AND posts.user_id = $2
			AND posts.deleted_at IS NULL
			AND posts.status = 'visible'
		LIMIT 1
	`, postID, userID, i18n.LocaleFromContext(ctx)))
	if err == nil {
		tx, err := r.db.Pool().Begin(ctx)
		if err != nil {
			return post.Post{}, share.MapDBError(ctx, postRepoService, "posts.update.begin", err, post.ErrDependencyUnavailable, post.ErrInternal)
		}
		defer tx.Rollback(ctx)

		_, err = tx.Exec(ctx, `
			UPDATE posts
			SET caption = $3,
				location_name = $4,
				latitude = $5,
				longitude = $6,
				category_id = NULLIF($7, 0),
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		`, postID, userID, update.Caption.String(), update.LocationName.String(), update.Latitude, update.Longitude, firstCategoryID(categoryIDs))
		if err != nil {
			return post.Post{}, share.MapDBError(ctx, postRepoService, "posts.update", err, post.ErrDependencyUnavailable, post.ErrInternal)
		}
		if err := r.replacePostCategories(ctx, tx, postID, categoryIDs); err != nil {
			return post.Post{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return post.Post{}, share.MapDBError(ctx, postRepoService, "posts.update.commit", err, post.ErrDependencyUnavailable, post.ErrInternal)
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

func (r *PostgresRepository) UpsertTrustVote(ctx context.Context, vote post.TrustVote) (post.PostTrustSummary, error) {
	if r.db == nil || r.db.Pool() == nil {
		return post.PostTrustSummary{}, post.ErrDependencyUnavailable
	}
	if err := r.ensurePostExists(ctx, vote.PostID); err != nil {
		return post.PostTrustSummary{}, err
	}

	if _, err := r.db.Pool().Exec(ctx, `
		INSERT INTO post_trust_votes (post_id, user_id, vote_type, reason)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (post_id, user_id) DO UPDATE
		SET vote_type = EXCLUDED.vote_type,
			reason = EXCLUDED.reason,
			updated_at = CURRENT_TIMESTAMP
	`, vote.PostID, vote.UserID, string(vote.Type), vote.Reason); err != nil {
		return post.PostTrustSummary{}, share.MapDBError(ctx, postRepoService, "post_trust_votes.upsert", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}

	return r.getPostTrustSummary(ctx, vote.PostID, vote.UserID)
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
		SELECT posts.id, posts.user_id, users.user_name, COALESCE(users.avatar_url, ''), posts.image_url,
				`+postImageURLsJSONSQL+`::text, posts.caption, posts.location_name,
				COALESCE(categories.id, 0), COALESCE(category_locale.name, categories.name, ''), COALESCE(categories.slug, ''),
				`+postCategoriesJSONSQL("$2")+`::text,
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
				(SELECT COUNT(*) FROM post_trust_votes WHERE post_trust_votes.post_id = posts.id AND post_trust_votes.vote_type = 'credible'),
				(SELECT COUNT(*) FROM post_trust_votes WHERE post_trust_votes.post_id = posts.id AND post_trust_votes.vote_type = 'suspicious'),
				(SELECT COUNT(*) FROM post_trust_votes WHERE post_trust_votes.post_id = posts.id AND post_trust_votes.vote_type = 'ai_generated'),
				(SELECT COUNT(*) FROM post_trust_votes WHERE post_trust_votes.post_id = posts.id AND post_trust_votes.vote_type = 'wrong_context'),
				(SELECT COUNT(*) FROM post_trust_votes WHERE post_trust_votes.post_id = posts.id AND post_trust_votes.vote_type = 'unsure'),
				COALESCE((
					SELECT post_trust_votes.vote_type
					FROM post_trust_votes
					WHERE post_trust_votes.post_id = posts.id AND post_trust_votes.user_id = $1
					LIMIT 1
				), ''),
				posts.created_at, posts.updated_at
		FROM post_saves
		INNER JOIN posts ON posts.id = post_saves.post_id
		INNER JOIN users ON users.id = posts.user_id
		LEFT JOIN categories ON categories.id = posts.category_id
		LEFT JOIN category_translations category_locale
			ON category_locale.category_id = categories.id
			AND category_locale.locale = $2
		WHERE post_saves.user_id = $1
			AND posts.deleted_at IS NULL
			AND posts.status = 'visible'
		ORDER BY post_saves.created_at DESC, post_saves.id DESC
		LIMIT 200
	`, userID, i18n.LocaleFromContext(ctx))
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
			WITH base_posts AS (
				SELECT posts.id, posts.user_id, users.user_name, COALESCE(users.avatar_url, '') AS user_avatar_url, posts.image_url,
					`+postImageURLsJSONSQL+`::text AS image_urls, posts.caption, posts.location_name,
					COALESCE(categories.id, 0) AS category_id,
					COALESCE(category_locale.name, categories.name, '') AS category_name,
					COALESCE(categories.slug, '') AS category_slug,
					`+postCategoriesJSONSQL("$16")+`::text AS categories,
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
					COALESCE(likes.likes_count, 0) AS likes_count,
					COALESCE(comments.comments_count, 0) AS comments_count,
					COALESCE(saves.saves_count, 0) AS saves_count,
					COALESCE(trust_votes.credible_count, 0) AS credible_count,
					COALESCE(trust_votes.suspicious_count, 0) AS suspicious_count,
					COALESCE(trust_votes.ai_generated_count, 0) AS ai_generated_count,
					COALESCE(trust_votes.wrong_context_count, 0) AS wrong_context_count,
					COALESCE(trust_votes.unsure_count, 0) AS unsure_count,
					COALESCE(viewer_trust_vote.vote_type, '') AS viewer_vote,
					posts.created_at,
					posts.updated_at,
					(
						((COALESCE(posts.latitude, 0)::double precision - $8::double precision) * (COALESCE(posts.latitude, 0)::double precision - $8::double precision))
						+ ((COALESCE(posts.longitude, 0)::double precision - $9::double precision) * (COALESCE(posts.longitude, 0)::double precision - $9::double precision))
					) AS nearby_rank
				FROM posts
				INNER JOIN users ON users.id = posts.user_id
				LEFT JOIN categories ON categories.id = posts.category_id
				LEFT JOIN category_translations category_locale
					ON category_locale.category_id = categories.id
					AND category_locale.locale = $16
				LEFT JOIN LATERAL (
					SELECT COUNT(*) AS likes_count
					FROM post_likes
					WHERE post_likes.post_id = posts.id
				) likes ON TRUE
				LEFT JOIN LATERAL (
					SELECT COUNT(*) AS comments_count
					FROM post_comments
					WHERE post_comments.post_id = posts.id
						AND post_comments.deleted_at IS NULL
						AND post_comments.status = 'visible'
				) comments ON TRUE
				LEFT JOIN LATERAL (
					SELECT COUNT(*) AS saves_count
					FROM post_saves
					WHERE post_saves.post_id = posts.id
				) saves ON TRUE
				LEFT JOIN LATERAL (
					SELECT
						COUNT(*) FILTER (WHERE vote_type = 'credible') AS credible_count,
						COUNT(*) FILTER (WHERE vote_type = 'suspicious') AS suspicious_count,
						COUNT(*) FILTER (WHERE vote_type = 'ai_generated') AS ai_generated_count,
						COUNT(*) FILTER (WHERE vote_type = 'wrong_context') AS wrong_context_count,
						COUNT(*) FILTER (WHERE vote_type = 'unsure') AS unsure_count
					FROM post_trust_votes
					WHERE post_trust_votes.post_id = posts.id
				) trust_votes ON TRUE
				LEFT JOIN LATERAL (
					SELECT post_trust_votes.vote_type
					FROM post_trust_votes
					WHERE post_trust_votes.post_id = posts.id
						AND post_trust_votes.user_id = $3
					LIMIT 1
				) viewer_trust_vote ON TRUE
				WHERE ($4 = '%%' OR posts.caption ILIKE $4 OR posts.location_name ILIKE $4 OR users.user_name ILIKE $4
					OR categories.name ILIKE $4 OR category_locale.name ILIKE $4 OR categories.slug ILIKE $4
					OR EXISTS (
						SELECT 1
						FROM post_categories search_post_categories
						INNER JOIN categories search_categories ON search_categories.id = search_post_categories.category_id
						LEFT JOIN category_translations search_category_locale
							ON search_category_locale.category_id = search_categories.id
							AND search_category_locale.locale = $16
						WHERE search_post_categories.post_id = posts.id
							AND (search_categories.name ILIKE $4 OR search_category_locale.name ILIKE $4 OR search_categories.slug ILIKE $4)
					))
					AND ($5 = '' OR categories.slug = $5 OR EXISTS (
						SELECT 1
						FROM post_categories filter_post_categories
						INNER JOIN categories filter_categories ON filter_categories.id = filter_post_categories.category_id
						WHERE filter_post_categories.post_id = posts.id
							AND filter_categories.slug = $5
					))
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
			),
			ranked_posts AS (
				SELECT id, user_id, user_name, user_avatar_url, image_url, image_urls, caption, location_name,
					category_id, category_name, category_slug, categories, latitude, longitude,
					is_liked, is_saved, status, likes_count, comments_count, saves_count,
					credible_count, suspicious_count, ai_generated_count, wrong_context_count, unsure_count, viewer_vote,
					created_at, updated_at,
					CASE
						WHEN $7 = 'nearby' THEN nearby_rank
						WHEN $7 = 'popular' THEN (likes_count * 3 + comments_count * 2 + saves_count)::double precision
						WHEN $7 = 'trending' THEN (likes_count * 3 + comments_count * 2 + saves_count)::double precision / (1 + EXTRACT(EPOCH FROM ($15::timestamptz - created_at)) / 86400)
						ELSE 0
					END AS rank_value
				FROM base_posts
			)
			SELECT id, user_id, user_name, user_avatar_url, image_url, image_urls, caption, location_name,
				category_id, category_name, category_slug, categories, latitude, longitude,
				is_liked, is_saved, status, likes_count, comments_count, saves_count,
				credible_count, suspicious_count, ai_generated_count, wrong_context_count, unsure_count, viewer_vote,
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
		`, filter.Limit, offset, filter.ViewerUserID, searchPattern, categorySlug, strings.TrimSpace(filter.Feed), sort, filter.Latitude, filter.Longitude, radiusDegrees, cursorCreatedAt, cursorID, cursorRank, hasCursor, rankAt, i18n.LocaleFromContext(ctx))
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

	item, err := scanPost(r.db.Pool().QueryRow(ctx, postSelectSQL("$2", "$3")+`
		WHERE posts.id = $1
			AND posts.deleted_at IS NULL
			AND posts.status = 'visible'
		LIMIT 1
	`, postID, viewerUserID, i18n.LocaleFromContext(ctx)))
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
	rows, err := r.db.Pool().Query(ctx, postSelectSQL("0", "$2")+`
		WHERE posts.location_name ILIKE $1
			AND posts.deleted_at IS NULL
			AND posts.status = 'visible'
		ORDER BY posts.created_at DESC, posts.id DESC
		LIMIT 100
	`, pattern, i18n.LocaleFromContext(ctx))
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

func (r *PostgresRepository) ensureCategoriesExist(ctx context.Context, categoryIDs []uint64) error {
	for _, categoryID := range categoryIDs {
		if err := r.ensureCategoryExists(ctx, categoryID); err != nil {
			return err
		}
	}

	return nil
}

type postCategoryWriter interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func (r *PostgresRepository) replacePostCategories(ctx context.Context, tx postCategoryWriter, postID uint64, categoryIDs []uint64) error {
	if _, err := tx.Exec(ctx, `
		DELETE FROM post_categories
		WHERE post_id = $1
	`, postID); err != nil {
		return share.MapDBError(ctx, postRepoService, "post_categories.delete", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}

	for _, categoryID := range categoryIDs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO post_categories (post_id, category_id)
			VALUES ($1, $2)
			ON CONFLICT (post_id, category_id) DO NOTHING
		`, postID, categoryID); err != nil {
			return share.MapDBError(ctx, postRepoService, "post_categories.insert", err, post.ErrDependencyUnavailable, post.ErrInternal)
		}
	}

	return nil
}

func firstCategoryID(categoryIDs []uint64) uint64 {
	if len(categoryIDs) == 0 {
		return 0
	}

	return categoryIDs[0]
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

func (r *PostgresRepository) getPostTrustSummary(ctx context.Context, postID uint64, viewerUserID uint64) (post.PostTrustSummary, error) {
	var summary post.PostTrustSummary
	if err := r.db.Pool().QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE vote_type = 'credible'),
			COUNT(*) FILTER (WHERE vote_type = 'suspicious'),
			COUNT(*) FILTER (WHERE vote_type = 'ai_generated'),
			COUNT(*) FILTER (WHERE vote_type = 'wrong_context'),
			COUNT(*) FILTER (WHERE vote_type = 'unsure'),
			COALESCE((
				SELECT viewer_vote.vote_type
				FROM post_trust_votes viewer_vote
				WHERE viewer_vote.post_id = $1 AND viewer_vote.user_id = $2
				LIMIT 1
			), '')
		FROM post_trust_votes
		WHERE post_id = $1
	`, postID, viewerUserID).Scan(
		&summary.CredibleCount,
		&summary.SuspiciousCount,
		&summary.AIGeneratedCount,
		&summary.WrongContextCount,
		&summary.UnsureCount,
		&summary.ViewerVote,
	); err != nil {
		return post.PostTrustSummary{}, share.MapDBError(ctx, postRepoService, "post_trust_votes.summary", err, post.ErrDependencyUnavailable, post.ErrInternal)
	}

	return summary.WithStatus(), nil
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
		SELECT posts.id, posts.user_id, users.user_name, COALESCE(users.avatar_url, ''), posts.image_url,
				`+postImageURLsJSONSQL+`::text, posts.caption, posts.location_name,
				COALESCE(categories.id, 0), COALESCE(category_locale.name, categories.name, ''), COALESCE(categories.slug, ''),
				`+postCategoriesJSONSQL("$4")+`::text,
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
				(SELECT COUNT(*) FROM post_trust_votes WHERE post_trust_votes.post_id = posts.id AND post_trust_votes.vote_type = 'credible'),
				(SELECT COUNT(*) FROM post_trust_votes WHERE post_trust_votes.post_id = posts.id AND post_trust_votes.vote_type = 'suspicious'),
				(SELECT COUNT(*) FROM post_trust_votes WHERE post_trust_votes.post_id = posts.id AND post_trust_votes.vote_type = 'ai_generated'),
				(SELECT COUNT(*) FROM post_trust_votes WHERE post_trust_votes.post_id = posts.id AND post_trust_votes.vote_type = 'wrong_context'),
				(SELECT COUNT(*) FROM post_trust_votes WHERE post_trust_votes.post_id = posts.id AND post_trust_votes.vote_type = 'unsure'),
				COALESCE((
					SELECT post_trust_votes.vote_type
					FROM post_trust_votes
					WHERE post_trust_votes.post_id = posts.id AND post_trust_votes.user_id = $3
					LIMIT 1
				), ''),
				posts.created_at, posts.updated_at
		FROM saved_collection_posts
		INNER JOIN posts ON posts.id = saved_collection_posts.post_id
		INNER JOIN users ON users.id = posts.user_id
		LEFT JOIN categories ON categories.id = posts.category_id
		LEFT JOIN category_translations category_locale
			ON category_locale.category_id = categories.id
			AND category_locale.locale = $4
		WHERE saved_collection_posts.collection_id = $1
			AND saved_collection_posts.user_id = $2
			AND posts.deleted_at IS NULL
			AND posts.status = 'visible'
		ORDER BY saved_collection_posts.created_at DESC, saved_collection_posts.id DESC
		LIMIT 200
	`, collectionID, ownerUserID, viewerUserID, i18n.LocaleFromContext(ctx))
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

func applyScannedCategories(item *post.Post, rawCategories string) error {
	if strings.TrimSpace(rawCategories) != "" {
		var categories []post.PostCategory
		if err := json.Unmarshal([]byte(rawCategories), &categories); err != nil {
			return err
		}
		item.Categories = categories
		item.CategoryIDs = make([]uint64, 0, len(categories))
		for _, category := range categories {
			if category.ID == 0 {
				continue
			}
			item.CategoryIDs = append(item.CategoryIDs, category.ID)
		}
	}
	if len(item.Categories) == 0 && item.CategoryID != 0 {
		item.Categories = []post.PostCategory{{
			ID:   item.CategoryID,
			Name: item.CategoryName,
			Slug: item.CategorySlug,
		}}
		item.CategoryIDs = []uint64{item.CategoryID}
	}
	if item.CategoryID == 0 && len(item.Categories) > 0 {
		item.CategoryID = item.Categories[0].ID
		item.CategoryName = item.Categories[0].Name
		item.CategorySlug = item.Categories[0].Slug
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
		rawImageURLs    string
		rawCaption      string
		rawLocationName string
		rawCategories   string
	)

	err := scanner.Scan(
		&item.ID,
		&item.UserID,
		&item.UserName,
		&item.UserAvatarURL,
		&rawImageURL,
		&rawImageURLs,
		&rawCaption,
		&rawLocationName,
		&item.CategoryID,
		&item.CategoryName,
		&item.CategorySlug,
		&rawCategories,
		&item.Latitude,
		&item.Longitude,
		&item.IsLiked,
		&item.IsSaved,
		&item.Status,
		&item.LikesCount,
		&item.CommentsCount,
		&item.SavesCount,
		&item.TrustSummary.CredibleCount,
		&item.TrustSummary.SuspiciousCount,
		&item.TrustSummary.AIGeneratedCount,
		&item.TrustSummary.WrongContextCount,
		&item.TrustSummary.UnsureCount,
		&item.TrustSummary.ViewerVote,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return post.Post{}, err
	}
	if err := applyScannedCategories(&item, rawCategories); err != nil {
		return post.Post{}, err
	}

	item.ImageURL, item.ImageURLs, err = post.NewPostImageURLs(rawImageURL, parseScannedImageURLs(rawImageURLs))
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
	item.TrustSummary = item.TrustSummary.WithStatus()

	return item, nil
}

func scanPostWithRank(scanner rowScanner) (post.Post, error) {
	var (
		item            post.Post
		rawImageURL     string
		rawImageURLs    string
		rawCaption      string
		rawLocationName string
		rawCategories   string
	)

	err := scanner.Scan(
		&item.ID,
		&item.UserID,
		&item.UserName,
		&item.UserAvatarURL,
		&rawImageURL,
		&rawImageURLs,
		&rawCaption,
		&rawLocationName,
		&item.CategoryID,
		&item.CategoryName,
		&item.CategorySlug,
		&rawCategories,
		&item.Latitude,
		&item.Longitude,
		&item.IsLiked,
		&item.IsSaved,
		&item.Status,
		&item.LikesCount,
		&item.CommentsCount,
		&item.SavesCount,
		&item.TrustSummary.CredibleCount,
		&item.TrustSummary.SuspiciousCount,
		&item.TrustSummary.AIGeneratedCount,
		&item.TrustSummary.WrongContextCount,
		&item.TrustSummary.UnsureCount,
		&item.TrustSummary.ViewerVote,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.CursorRank,
	)
	if err != nil {
		return post.Post{}, err
	}
	if err := applyScannedCategories(&item, rawCategories); err != nil {
		return post.Post{}, err
	}

	item.ImageURL, item.ImageURLs, err = post.NewPostImageURLs(rawImageURL, parseScannedImageURLs(rawImageURLs))
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
	item.TrustSummary = item.TrustSummary.WithStatus()

	return item, nil
}

func marshalImageURLs(primary post.ImageURL, imageURLs []post.ImageURL) (string, error) {
	values := make([]string, 0, len(imageURLs))
	source := imageURLs
	if len(source) == 0 && primary != "" {
		source = []post.ImageURL{primary}
	}
	for _, imageURL := range source {
		if imageURL == "" {
			continue
		}
		values = append(values, imageURL.String())
	}

	payload, err := json.Marshal(values)
	if err != nil {
		return "", post.ErrInternal
	}

	return string(payload), nil
}

func parseScannedImageURLs(raw string) []string {
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil
	}

	return values
}

var _ post.Repository = (*PostgresRepository)(nil)
