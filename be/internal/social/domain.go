package social

import (
	"context"
	"errors"
	"time"

	"falzo-be/internal/post"
)

var (
	ErrDependencyUnavailable = errors.New("social dependency unavailable")
	ErrInternal              = errors.New("social internal error")
	ErrUserIDRequired        = errors.New("user id is required")
	ErrTargetUserIDRequired  = errors.New("target user id is required")
	ErrCannotFollowSelf      = errors.New("cannot follow self")
	ErrUserNotFound          = errors.New("user not found")
)

type Repository interface {
	GetPublicProfile(ctx context.Context, userID uint64, viewerUserID uint64) (PublicProfile, error)
	Follow(ctx context.Context, followerID uint64, followingID uint64) (bool, error)
	Unfollow(ctx context.Context, followerID uint64, followingID uint64) error
	ListFollowerIDs(ctx context.Context, userID uint64) ([]uint64, error)
}

type PublicProfile struct {
	UserID         uint64          `json:"user_id"`
	UserName       string          `json:"user_name"`
	FullName       string          `json:"full_name,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	PostsCount     int             `json:"posts_count"`
	FollowersCount int             `json:"followers_count"`
	FollowingCount int             `json:"following_count"`
	IsFollowing    bool            `json:"is_following"`
	Posts          []post.PostView `json:"posts"`
}
