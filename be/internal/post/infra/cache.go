package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"falzo-be/internal/post"
	pkgcache "falzo-be/pkg/cache"

	goredis "github.com/redis/go-redis/v9"
)

type CachedPostRepository struct {
	next         post.Repository
	redis        *goredis.Client
	firstPageTTL time.Duration
}

type cachedPost struct {
	ID            uint64                `json:"id"`
	UserID        uint64                `json:"user_id"`
	UserName      string                `json:"user_name"`
	UserAvatarURL string                `json:"user_avatar_url"`
	CategoryID    uint64                `json:"category_id"`
	CategoryName  string                `json:"category_name"`
	CategorySlug  string                `json:"category_slug"`
	Categories    []post.PostCategory   `json:"categories"`
	ImageURL      string                `json:"image_url"`
	ImageURLs     []string              `json:"image_urls"`
	Caption       string                `json:"caption"`
	LocationName  string                `json:"location_name"`
	Latitude      float64               `json:"latitude"`
	Longitude     float64               `json:"longitude"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
	IsLiked       bool                  `json:"is_liked"`
	IsSaved       bool                  `json:"is_saved"`
	Status        string                `json:"status"`
	LikesCount    int                   `json:"likes_count"`
	CommentsCount int                   `json:"comments_count"`
	SavesCount    int                   `json:"saves_count"`
	TrustSummary  post.PostTrustSummary `json:"trust_summary"`
}

func NewCachedPostRepository(next post.Repository, cache pkgcache.Client, firstPageTTL time.Duration) post.Repository {
	if next == nil || cache == nil || cache.Client() == nil || firstPageTTL <= 0 {
		return next
	}

	return &CachedPostRepository{next: next, redis: cache.Client(), firstPageTTL: firstPageTTL}
}

func (r *CachedPostRepository) Create(ctx context.Context, item *post.Post) error {
	if err := r.next.Create(ctx, item); err != nil {
		return err
	}
	r.invalidatePublicFeed(ctx)
	return nil
}

func (r *CachedPostRepository) UpdatePost(ctx context.Context, postID uint64, userID uint64, update post.PostUpdate) (post.Post, error) {
	item, err := r.next.UpdatePost(ctx, postID, userID, update)
	if err != nil {
		return post.Post{}, err
	}
	r.invalidatePublicFeed(ctx)
	return item, nil
}

func (r *CachedPostRepository) DeletePost(ctx context.Context, postID uint64, actor post.ModerationActor) error {
	if err := r.next.DeletePost(ctx, postID, actor); err != nil {
		return err
	}
	r.invalidatePublicFeed(ctx)
	return nil
}

func (r *CachedPostRepository) HidePost(ctx context.Context, postID uint64, actor post.ModerationActor, reason post.ReportReason) error {
	if err := r.next.HidePost(ctx, postID, actor, reason); err != nil {
		return err
	}
	r.invalidatePublicFeed(ctx)
	return nil
}

func (r *CachedPostRepository) ReportPost(ctx context.Context, report post.ContentReport) error {
	return r.next.ReportPost(ctx, report)
}

func (r *CachedPostRepository) ReportComment(ctx context.Context, report post.ContentReport) error {
	return r.next.ReportComment(ctx, report)
}

func (r *CachedPostRepository) UpsertTrustVote(ctx context.Context, vote post.TrustVote) (post.PostTrustSummary, error) {
	summary, err := r.next.UpsertTrustVote(ctx, vote)
	if err != nil {
		return post.PostTrustSummary{}, err
	}
	r.invalidatePublicFeed(ctx)
	return summary, nil
}

func (r *CachedPostRepository) DeleteComment(ctx context.Context, postID uint64, commentID uint64, actor post.ModerationActor) error {
	if err := r.next.DeleteComment(ctx, postID, commentID, actor); err != nil {
		return err
	}
	r.invalidatePublicFeed(ctx)
	return nil
}

func (r *CachedPostRepository) HideComment(ctx context.Context, postID uint64, commentID uint64, actor post.ModerationActor, reason post.ReportReason) error {
	if err := r.next.HideComment(ctx, postID, commentID, actor, reason); err != nil {
		return err
	}
	r.invalidatePublicFeed(ctx)
	return nil
}

func (r *CachedPostRepository) Comment(ctx context.Context, comment *post.Comment) error {
	if err := r.next.Comment(ctx, comment); err != nil {
		return err
	}
	r.invalidatePublicFeed(ctx)
	return nil
}

func (r *CachedPostRepository) UpdateComment(ctx context.Context, postID uint64, commentID uint64, userID uint64, content post.Content) (post.Comment, error) {
	comment, err := r.next.UpdateComment(ctx, postID, commentID, userID, content)
	if err != nil {
		return post.Comment{}, err
	}
	r.invalidatePublicFeed(ctx)
	return comment, nil
}

func (r *CachedPostRepository) Like(ctx context.Context, postID uint64, userID uint64) error {
	if err := r.next.Like(ctx, postID, userID); err != nil {
		return err
	}
	r.invalidatePublicFeed(ctx)
	return nil
}

func (r *CachedPostRepository) Unlike(ctx context.Context, postID uint64, userID uint64) error {
	if err := r.next.Unlike(ctx, postID, userID); err != nil {
		return err
	}
	r.invalidatePublicFeed(ctx)
	return nil
}

func (r *CachedPostRepository) Save(ctx context.Context, postID uint64, userID uint64) error {
	if err := r.next.Save(ctx, postID, userID); err != nil {
		return err
	}
	r.invalidatePublicFeed(ctx)
	return nil
}

func (r *CachedPostRepository) Unsave(ctx context.Context, postID uint64, userID uint64) error {
	if err := r.next.Unsave(ctx, postID, userID); err != nil {
		return err
	}
	r.invalidatePublicFeed(ctx)
	return nil
}

func (r *CachedPostRepository) CreateSavedCollection(ctx context.Context, collection *post.SavedCollection) error {
	return r.next.CreateSavedCollection(ctx, collection)
}

func (r *CachedPostRepository) ListSavedCollections(ctx context.Context, userID uint64) ([]post.SavedCollection, error) {
	return r.next.ListSavedCollections(ctx, userID)
}

func (r *CachedPostRepository) ListSavedPosts(ctx context.Context, userID uint64) ([]post.Post, error) {
	return r.next.ListSavedPosts(ctx, userID)
}

func (r *CachedPostRepository) AddPostToSavedCollection(ctx context.Context, collectionID uint64, postID uint64, userID uint64) error {
	return r.next.AddPostToSavedCollection(ctx, collectionID, postID, userID)
}

func (r *CachedPostRepository) RemovePostFromSavedCollection(ctx context.Context, collectionID uint64, postID uint64, userID uint64) error {
	return r.next.RemovePostFromSavedCollection(ctx, collectionID, postID, userID)
}

func (r *CachedPostRepository) DeleteSavedCollection(ctx context.Context, collectionID uint64, userID uint64) error {
	return r.next.DeleteSavedCollection(ctx, collectionID, userID)
}

func (r *CachedPostRepository) UpdateSavedCollectionVisibility(ctx context.Context, collectionID uint64, userID uint64, isPublic bool) (post.SavedCollection, error) {
	return r.next.UpdateSavedCollectionVisibility(ctx, collectionID, userID, isPublic)
}

func (r *CachedPostRepository) GetPublicSavedCollection(ctx context.Context, shareSlug string, viewerUserID uint64) (*post.SavedCollection, error) {
	return r.next.GetPublicSavedCollection(ctx, shareSlug, viewerUserID)
}

func (r *CachedPostRepository) GetPosts(ctx context.Context, filter post.PostListFilter) ([]post.Post, error) {
	if !isCacheablePublicFirstPage(filter) {
		return r.next.GetPosts(ctx, filter)
	}

	key := feedFirstPageCacheKey(filter)
	value, err := r.redis.Get(ctx, key).Result()
	if err == nil {
		items, decodeErr := decodePosts([]byte(value))
		if decodeErr == nil {
			return items, nil
		}
		_ = r.redis.Del(ctx, key).Err()
	}

	items, err := r.next.GetPosts(ctx, filter)
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(encodePosts(items))
	if err == nil {
		_ = r.redis.Set(ctx, key, payload, r.firstPageTTL).Err()
	}
	return items, nil
}

func (r *CachedPostRepository) GetPostDetail(ctx context.Context, postID uint64, viewerUserID uint64) (*post.Post, error) {
	return r.next.GetPostDetail(ctx, postID, viewerUserID)
}

func (r *CachedPostRepository) GetPostsByLocation(ctx context.Context, locationName post.LocationName) ([]post.Post, error) {
	return r.next.GetPostsByLocation(ctx, locationName)
}

func (r *CachedPostRepository) GetComments(ctx context.Context, postID uint64, page int, limit int) ([]post.Comment, error) {
	return r.next.GetComments(ctx, postID, page, limit)
}

func (r *CachedPostRepository) invalidatePublicFeed(ctx context.Context) {
	var cursor uint64
	for {
		keys, nextCursor, err := r.redis.Scan(ctx, cursor, feedFirstPageCachePattern(), 100).Result()
		if err != nil {
			return
		}
		if len(keys) > 0 {
			_ = r.redis.Del(ctx, keys...).Err()
		}
		if nextCursor == 0 {
			return
		}
		cursor = nextCursor
	}
}

func isCacheablePublicFirstPage(filter post.PostListFilter) bool {
	return filter.ViewerUserID == 0 &&
		filter.Page == 1 &&
		filter.Cursor == nil &&
		strings.TrimSpace(filter.Feed) == "" &&
		filter.Limit > 0
}

func feedFirstPageCacheKey(filter post.PostListFilter) string {
	return fmt.Sprintf(
		"posts:feed:first:v1:limit=%d:search=%s:category=%s:sort=%s:lat=%.5f:lng=%.5f:radius=%d",
		filter.Limit,
		strings.TrimSpace(filter.Search),
		strings.TrimSpace(filter.CategorySlug),
		strings.TrimSpace(filter.Sort),
		filter.Latitude,
		filter.Longitude,
		filter.RadiusMeters,
	)
}

func feedFirstPageCachePattern() string {
	return "posts:feed:first:v1:*"
}

func encodePosts(items []post.Post) []cachedPost {
	cached := make([]cachedPost, 0, len(items))
	for _, item := range items {
		cached = append(cached, cachedPost{
			ID:            item.ID,
			UserID:        item.UserID,
			UserName:      item.UserName,
			UserAvatarURL: item.UserAvatarURL,
			CategoryID:    item.CategoryID,
			CategoryName:  item.CategoryName,
			CategorySlug:  item.CategorySlug,
			Categories:    item.Categories,
			ImageURL:      item.ImageURL.String(),
			ImageURLs:     cachedImageURLStrings(item.ImageURL, item.ImageURLs),
			Caption:       item.Caption.String(),
			LocationName:  item.LocationName.String(),
			Latitude:      item.Latitude,
			Longitude:     item.Longitude,
			CreatedAt:     item.CreatedAt,
			UpdatedAt:     item.UpdatedAt,
			IsLiked:       item.IsLiked,
			IsSaved:       item.IsSaved,
			Status:        item.Status,
			LikesCount:    item.LikesCount,
			CommentsCount: item.CommentsCount,
			SavesCount:    item.SavesCount,
			TrustSummary:  item.TrustSummary,
		})
	}
	return cached
}

func decodePosts(payload []byte) ([]post.Post, error) {
	var cached []cachedPost
	if err := json.Unmarshal(payload, &cached); err != nil {
		return nil, err
	}

	items := make([]post.Post, 0, len(cached))
	for _, item := range cached {
		imageURL, imageURLs, err := post.NewPostImageURLs(item.ImageURL, item.ImageURLs)
		if err != nil {
			return nil, err
		}
		caption, err := post.NewCaption(item.Caption)
		if err != nil {
			return nil, err
		}
		locationName, err := post.NewLocationName(item.LocationName)
		if err != nil {
			return nil, err
		}
		items = append(items, post.Post{
			ID:            item.ID,
			UserID:        item.UserID,
			UserName:      item.UserName,
			UserAvatarURL: item.UserAvatarURL,
			CategoryID:    item.CategoryID,
			CategoryName:  item.CategoryName,
			CategorySlug:  item.CategorySlug,
			Categories:    item.Categories,
			ImageURL:      imageURL,
			ImageURLs:     imageURLs,
			Caption:       caption,
			LocationName:  locationName,
			Latitude:      item.Latitude,
			Longitude:     item.Longitude,
			CreatedAt:     item.CreatedAt,
			UpdatedAt:     item.UpdatedAt,
			IsLiked:       item.IsLiked,
			IsSaved:       item.IsSaved,
			Status:        item.Status,
			LikesCount:    item.LikesCount,
			CommentsCount: item.CommentsCount,
			SavesCount:    item.SavesCount,
			TrustSummary:  item.TrustSummary.WithStatus(),
		})
	}
	return items, nil
}

func cachedImageURLStrings(primary post.ImageURL, imageURLs []post.ImageURL) []string {
	source := imageURLs
	if len(source) == 0 && primary != "" {
		source = []post.ImageURL{primary}
	}

	values := make([]string, 0, len(source))
	for _, imageURL := range source {
		if imageURL == "" {
			continue
		}
		values = append(values, imageURL.String())
	}

	return values
}

var _ post.Repository = (*CachedPostRepository)(nil)
