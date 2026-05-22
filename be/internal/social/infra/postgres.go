package infra

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"falzo-be/internal/post"
	"falzo-be/internal/share"
	"falzo-be/internal/social"
	"falzo-be/pkg/database"

	"github.com/jackc/pgx/v5"
)

const socialRepoService = "social"

type PostgresRepository struct {
	db database.Client
}

func NewPostgresRepository(db database.Client) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetPublicProfile(ctx context.Context, userID uint64, viewerUserID uint64) (social.PublicProfile, error) {
	if r.db == nil || r.db.Pool() == nil {
		return social.PublicProfile{}, social.ErrDependencyUnavailable
	}

	var profile social.PublicProfile
	var postsCount int64
	var followersCount int64
	var followingCount int64
	err := r.db.Pool().QueryRow(ctx, `
		SELECT users.id, users.user_name, COALESCE(users.full_name, ''), COALESCE(users.avatar_url, ''), users.created_at,
			(SELECT COUNT(*) FROM posts WHERE posts.user_id = users.id),
			(SELECT COUNT(*) FROM user_follows WHERE user_follows.following_id = users.id),
			(SELECT COUNT(*) FROM user_follows WHERE user_follows.follower_id = users.id),
			EXISTS (
				SELECT 1
				FROM user_follows
				WHERE follower_id = $2 AND following_id = users.id
			)
		FROM users
		WHERE users.id = $1 AND users.is_active = TRUE
		LIMIT 1
	`, userID, viewerUserID).Scan(
		&profile.UserID,
		&profile.UserName,
		&profile.FullName,
		&profile.AvatarURL,
		&profile.CreatedAt,
		&postsCount,
		&followersCount,
		&followingCount,
		&profile.IsFollowing,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return social.PublicProfile{}, social.ErrUserNotFound
		}
		return social.PublicProfile{}, mapDBError(ctx, "social.profile", err)
	}
	profile.PostsCount = int(postsCount)
	profile.FollowersCount = int(followersCount)
	profile.FollowingCount = int(followingCount)
	profile.AvatarURLAlias = profile.AvatarURL

	posts, err := r.getUserPosts(ctx, userID, viewerUserID)
	if err != nil {
		return social.PublicProfile{}, err
	}
	profile.Posts = posts

	return profile, nil
}

func (r *PostgresRepository) Follow(ctx context.Context, followerID uint64, followingID uint64) (bool, error) {
	if r.db == nil || r.db.Pool() == nil {
		return false, social.ErrDependencyUnavailable
	}
	if err := r.ensureUserExists(ctx, followingID); err != nil {
		return false, err
	}

	result, err := r.db.Pool().Exec(ctx, `
		INSERT INTO user_follows (follower_id, following_id)
		VALUES ($1, $2)
		ON CONFLICT (follower_id, following_id) DO NOTHING
	`, followerID, followingID)
	if err != nil {
		return false, mapDBError(ctx, "social.follow", err)
	}

	return result.RowsAffected() > 0, nil
}

func (r *PostgresRepository) Unfollow(ctx context.Context, followerID uint64, followingID uint64) error {
	if r.db == nil || r.db.Pool() == nil {
		return social.ErrDependencyUnavailable
	}

	if _, err := r.db.Pool().Exec(ctx, `
		DELETE FROM user_follows
		WHERE follower_id = $1 AND following_id = $2
	`, followerID, followingID); err != nil {
		return mapDBError(ctx, "social.unfollow", err)
	}

	return nil
}

func (r *PostgresRepository) Block(ctx context.Context, blockerID uint64, blockedID uint64) error {
	if r.db == nil || r.db.Pool() == nil {
		return social.ErrDependencyUnavailable
	}
	if err := r.ensureUserExists(ctx, blockedID); err != nil {
		return err
	}

	tx, err := r.db.Pool().Begin(ctx)
	if err != nil {
		return mapDBError(ctx, "social.block.begin", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO user_blocks (blocker_user_id, blocked_user_id)
		VALUES ($1, $2)
		ON CONFLICT (blocker_user_id, blocked_user_id) DO NOTHING
	`, blockerID, blockedID); err != nil {
		return mapDBError(ctx, "social.block", err)
	}

	if _, err := tx.Exec(ctx, `
		DELETE FROM user_follows
		WHERE (follower_id = $1 AND following_id = $2)
			OR (follower_id = $2 AND following_id = $1)
	`, blockerID, blockedID); err != nil {
		return mapDBError(ctx, "social.block.unfollow", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return mapDBError(ctx, "social.block.commit", err)
	}

	return nil
}

func (r *PostgresRepository) Unblock(ctx context.Context, blockerID uint64, blockedID uint64) error {
	if r.db == nil || r.db.Pool() == nil {
		return social.ErrDependencyUnavailable
	}

	if _, err := r.db.Pool().Exec(ctx, `
		DELETE FROM user_blocks
		WHERE blocker_user_id = $1 AND blocked_user_id = $2
	`, blockerID, blockedID); err != nil {
		return mapDBError(ctx, "social.unblock", err)
	}

	return nil
}

func (r *PostgresRepository) ListFollowerIDs(ctx context.Context, userID uint64) ([]uint64, error) {
	if r.db == nil || r.db.Pool() == nil {
		return nil, social.ErrDependencyUnavailable
	}

	rows, err := r.db.Pool().Query(ctx, `
		SELECT follower_id
		FROM user_follows
		WHERE following_id = $1
		ORDER BY created_at DESC, id DESC
	`, userID)
	if err != nil {
		return nil, mapDBError(ctx, "social.followers", err)
	}
	defer rows.Close()

	ids := make([]uint64, 0)
	for rows.Next() {
		var id uint64
		if err := rows.Scan(&id); err != nil {
			return nil, mapDBError(ctx, "social.followers.scan", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, mapDBError(ctx, "social.followers.rows", err)
	}

	return ids, nil
}

func (r *PostgresRepository) ensureUserExists(ctx context.Context, userID uint64) error {
	var exists bool
	err := r.db.Pool().QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE id = $1 AND is_active = TRUE
		)
	`, userID).Scan(&exists)
	if err != nil {
		return mapDBError(ctx, "social.user_exists", err)
	}
	if !exists {
		return social.ErrUserNotFound
	}

	return nil
}

func (r *PostgresRepository) getUserPosts(ctx context.Context, userID uint64, viewerUserID uint64) ([]post.PostView, error) {
	rows, err := r.db.Pool().Query(ctx, `
		SELECT posts.id, posts.user_id, users.user_name, posts.image_url, posts.caption, COALESCE(posts.location_name, ''),
			COALESCE(categories.id, 0), COALESCE(categories.name, ''), COALESCE(categories.slug, ''),
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
			posts.status,
			(SELECT COUNT(*) FROM post_likes WHERE post_likes.post_id = posts.id),
			(SELECT COUNT(*) FROM post_comments WHERE post_comments.post_id = posts.id AND post_comments.deleted_at IS NULL AND post_comments.status = 'visible'),
			(SELECT COUNT(*) FROM post_saves WHERE post_saves.post_id = posts.id),
			posts.created_at, posts.updated_at
		FROM posts
		INNER JOIN users ON users.id = posts.user_id
		LEFT JOIN categories ON categories.id = posts.category_id
		WHERE posts.user_id = $1
			AND posts.deleted_at IS NULL
			AND posts.status = 'visible'
		ORDER BY posts.created_at DESC, posts.id DESC
		LIMIT 60
	`, userID, viewerUserID)
	if err != nil {
		return nil, mapDBError(ctx, "social.user_posts", err)
	}
	defer rows.Close()

	items := make([]post.PostView, 0)
	for rows.Next() {
		var item post.PostView
		var updatedAt time.Time
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.UserName,
			&item.ImageURL,
			&item.Caption,
			&item.LocationName,
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
			&updatedAt,
		); err != nil {
			return nil, mapDBError(ctx, "social.user_posts.scan", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, mapDBError(ctx, "social.user_posts.rows", err)
	}

	return items, nil
}

func mapDBError(ctx context.Context, operation string, err error) error {
	return share.MapDBError(ctx, socialRepoService, operation, err, social.ErrDependencyUnavailable, social.ErrInternal)
}

var _ social.Repository = (*PostgresRepository)(nil)
