package post

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var (
	ErrNotFound                  = errors.New("post not found")
	ErrDependencyUnavailable     = errors.New("post dependency unavailable")
	ErrInternal                  = errors.New("post internal error")
	ErrUserIDRequired            = errors.New("user id is required")
	ErrPostIDRequired            = errors.New("post id is required")
	ErrPageMustBePositive        = errors.New("page must be greater than 0")
	ErrLimitMustBePositive       = errors.New("limit must be greater than 0")
	ErrLimitTooLarge             = errors.New("limit exceeds maximum")
	ErrInvalidCursor             = errors.New("invalid cursor")
	ErrInvalidFeed               = errors.New("invalid feed")
	ErrLocationNameRequired      = errors.New("location name is required")
	ErrLatitudeOutOfRange        = errors.New("latitude must be between -90 and 90")
	ErrLongitudeOutOfRange       = errors.New("longitude must be between -180 and 180")
	ErrImageURLRequired          = errors.New("image url is required")
	ErrInvalidImageURL           = errors.New("invalid image url")
	ErrCaptionTooLong            = errors.New("caption exceeds max length")
	ErrLocationNameTooLong       = errors.New("location name exceeds max length")
	ErrCategoryNotFound          = errors.New("category not found")
	ErrTooManyCategories         = errors.New("too many categories")
	ErrCommentRequired           = errors.New("comment content is required")
	ErrCommentTooLong            = errors.New("comment content exceeds max length")
	ErrReplyCommentNotFound      = errors.New("reply comment not found")
	ErrCommentNotFound           = errors.New("comment not found")
	ErrCommentUpdateForbidden    = errors.New("comment update forbidden")
	ErrCollectionIDRequired      = errors.New("collection id is required")
	ErrCollectionNameRequired    = errors.New("collection name is required")
	ErrCollectionNameTooLong     = errors.New("collection name exceeds max length")
	ErrCollectionNameTaken       = errors.New("collection name already exists")
	ErrCollectionNotFound        = errors.New("collection not found")
	ErrCollectionSlugRequired    = errors.New("collection share slug is required")
	ErrPostUpdateForbidden       = errors.New("post update forbidden")
	ErrPostModerationForbidden   = errors.New("post moderation forbidden")
	ErrInvalidPostSort           = errors.New("invalid post sort")
	ErrNearbyCoordinatesRequired = errors.New("nearby coordinates are required")
	ErrNearbyRadiusTooLarge      = errors.New("nearby radius exceeds maximum")
	ErrReportReasonRequired      = errors.New("report reason is required")
	ErrReportReasonTooLong       = errors.New("report reason exceeds max length")
	ErrTrustVoteTypeRequired     = errors.New("trust vote type is required")
	ErrInvalidTrustVoteType      = errors.New("invalid trust vote type")
	ErrTrustVoteReasonTooLong    = errors.New("trust vote reason exceeds max length")
)

type Repository interface {
	Create(ctx context.Context, post *Post) error
	UpdatePost(ctx context.Context, postID uint64, userID uint64, update PostUpdate) (Post, error)
	DeletePost(ctx context.Context, postID uint64, actor ModerationActor) error
	HidePost(ctx context.Context, postID uint64, actor ModerationActor, reason ReportReason) error
	ReportPost(ctx context.Context, report ContentReport) error
	ReportComment(ctx context.Context, report ContentReport) error
	UpsertTrustVote(ctx context.Context, vote TrustVote) (PostTrustSummary, error)
	DeleteComment(ctx context.Context, postID uint64, commentID uint64, actor ModerationActor) error
	HideComment(ctx context.Context, postID uint64, commentID uint64, actor ModerationActor, reason ReportReason) error
	Comment(ctx context.Context, comment *Comment) error
	UpdateComment(ctx context.Context, postID uint64, commentID uint64, userID uint64, content Content) (Comment, error)
	Like(ctx context.Context, postID uint64, userID uint64) error
	Unlike(ctx context.Context, postID uint64, userID uint64) error
	Save(ctx context.Context, postID uint64, userID uint64) error
	Unsave(ctx context.Context, postID uint64, userID uint64) error
	CreateSavedCollection(ctx context.Context, collection *SavedCollection) error
	ListSavedCollections(ctx context.Context, userID uint64) ([]SavedCollection, error)
	ListSavedPosts(ctx context.Context, userID uint64) ([]Post, error)
	AddPostToSavedCollection(ctx context.Context, collectionID uint64, postID uint64, userID uint64) error
	RemovePostFromSavedCollection(ctx context.Context, collectionID uint64, postID uint64, userID uint64) error
	DeleteSavedCollection(ctx context.Context, collectionID uint64, userID uint64) error
	UpdateSavedCollectionVisibility(ctx context.Context, collectionID uint64, userID uint64, isPublic bool) (SavedCollection, error)
	GetPublicSavedCollection(ctx context.Context, shareSlug string, viewerUserID uint64) (*SavedCollection, error)
	GetPosts(ctx context.Context, filter PostListFilter) ([]Post, error)
	GetPostDetail(ctx context.Context, postID uint64, viewerUserID uint64) (*Post, error)
	GetPostsByLocation(ctx context.Context, locationName LocationName) ([]Post, error)
	GetComments(ctx context.Context, postID uint64, page int, limit int) ([]Comment, error)
}

type Post struct {
	ID            uint64           `json:"id"`
	UserID        uint64           `json:"user_id"`
	UserName      string           `json:"user_name"`
	UserAvatarURL string           `json:"user_avatar_url,omitempty"`
	CategoryID    uint64           `json:"category_id,omitempty"`
	CategoryName  string           `json:"category_name,omitempty"`
	CategorySlug  string           `json:"category_slug,omitempty"`
	CategoryIDs   []uint64         `json:"-"`
	Categories    []PostCategory   `json:"categories,omitempty"`
	ImageURL      ImageURL         `json:"-"`
	Caption       Caption          `json:"-"`
	LocationName  LocationName     `json:"-"`
	Latitude      float64          `json:"latitude"`
	Longitude     float64          `json:"longitude"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"-"`
	CursorRank    float64          `json:"-"`
	IsLiked       bool             `json:"is_liked"`
	IsSaved       bool             `json:"is_saved"`
	Status        string           `json:"status"`
	LikesCount    int              `json:"likes_count"`
	CommentsCount int              `json:"comments_count"`
	SavesCount    int              `json:"saves_count"`
	TrustSummary  PostTrustSummary `json:"trust_summary"`
}

type PostView struct {
	ID            uint64           `json:"id"`
	UserID        uint64           `json:"user_id"`
	UserName      string           `json:"user_name"`
	UserAvatarURL string           `json:"user_avatar_url,omitempty"`
	CategoryID    uint64           `json:"category_id,omitempty"`
	CategoryName  string           `json:"category_name,omitempty"`
	CategorySlug  string           `json:"category_slug,omitempty"`
	Categories    []PostCategory   `json:"categories,omitempty"`
	ImageURL      string           `json:"image_url"`
	Caption       string           `json:"caption"`
	LocationName  string           `json:"location_name"`
	Latitude      float64          `json:"latitude"`
	Longitude     float64          `json:"longitude"`
	CreatedAt     time.Time        `json:"created_at"`
	IsLiked       bool             `json:"is_liked"`
	IsSaved       bool             `json:"is_saved"`
	Status        string           `json:"status"`
	LikesCount    int              `json:"likes_count"`
	CommentsCount int              `json:"comments_count"`
	SavesCount    int              `json:"saves_count"`
	TrustSummary  PostTrustSummary `json:"trust_summary"`
}

func (p Post) View() PostView {
	categories := normalizedPostCategories(p)

	return PostView{
		ID:            p.ID,
		UserID:        p.UserID,
		UserName:      p.UserName,
		UserAvatarURL: p.UserAvatarURL,
		CategoryID:    p.CategoryID,
		CategoryName:  p.CategoryName,
		CategorySlug:  p.CategorySlug,
		Categories:    categories,
		ImageURL:      p.ImageURL.String(),
		Caption:       p.Caption.String(),
		LocationName:  p.LocationName.String(),
		Latitude:      p.Latitude,
		Longitude:     p.Longitude,
		CreatedAt:     p.CreatedAt,
		IsLiked:       p.IsLiked,
		IsSaved:       p.IsSaved,
		Status:        p.Status,
		LikesCount:    p.LikesCount,
		CommentsCount: p.CommentsCount,
		SavesCount:    p.SavesCount,
		TrustSummary:  p.TrustSummary.WithStatus(),
	}
}

type NewPostInput struct {
	UserID       uint64
	CategoryID   uint64
	CategoryIDs  []uint64
	ImageURL     string
	Caption      string
	LocationName string
	Latitude     float64
	Longitude    float64
}

type PostUpdate struct {
	Caption      Caption
	LocationName LocationName
	Latitude     float64
	Longitude    float64
	CategoryID   uint64
	CategoryIDs  []uint64
}

type PostCategory struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type ModerationActor struct {
	UserID  uint64
	IsAdmin bool
}

type ContentReport struct {
	ReporterUserID uint64
	PostID         uint64
	CommentID      uint64
	Reason         ReportReason
}

type TrustVoteType string

const (
	TrustVoteCredible     TrustVoteType = "credible"
	TrustVoteSuspicious   TrustVoteType = "suspicious"
	TrustVoteAIGenerated  TrustVoteType = "ai_generated"
	TrustVoteWrongContext TrustVoteType = "wrong_context"
	TrustVoteUnsure       TrustVoteType = "unsure"
)

type TrustVote struct {
	PostID uint64
	UserID uint64
	Type   TrustVoteType
	Reason string
}

type PostTrustSummary struct {
	Status            string `json:"status"`
	TotalCount        int    `json:"total_count"`
	CredibleCount     int    `json:"credible_count"`
	SuspiciousCount   int    `json:"suspicious_count"`
	AIGeneratedCount  int    `json:"ai_generated_count"`
	WrongContextCount int    `json:"wrong_context_count"`
	UnsureCount       int    `json:"unsure_count"`
	ViewerVote        string `json:"viewer_vote,omitempty"`
}

func NewTrustVote(postID uint64, userID uint64, rawType string, rawReason string) (TrustVote, error) {
	if postID == 0 {
		return TrustVote{}, ErrPostIDRequired
	}
	if userID == 0 {
		return TrustVote{}, ErrUserIDRequired
	}

	voteType, err := NewTrustVoteType(rawType)
	if err != nil {
		return TrustVote{}, err
	}

	reason := strings.TrimSpace(rawReason)
	if len(reason) > 500 {
		return TrustVote{}, ErrTrustVoteReasonTooLong
	}

	return TrustVote{
		PostID: postID,
		UserID: userID,
		Type:   voteType,
		Reason: reason,
	}, nil
}

func NewTrustVoteType(raw string) (TrustVoteType, error) {
	value := TrustVoteType(strings.TrimSpace(raw))
	if value == "" {
		return "", ErrTrustVoteTypeRequired
	}

	switch value {
	case TrustVoteCredible, TrustVoteSuspicious, TrustVoteAIGenerated, TrustVoteWrongContext, TrustVoteUnsure:
		return value, nil
	default:
		return "", ErrInvalidTrustVoteType
	}
}

func (s PostTrustSummary) WithStatus() PostTrustSummary {
	if s.Status != "" {
		return s
	}

	s.TotalCount = s.CredibleCount + s.SuspiciousCount + s.AIGeneratedCount + s.WrongContextCount + s.UnsureCount
	concernCount := s.SuspiciousCount + s.AIGeneratedCount + s.WrongContextCount
	switch {
	case s.TotalCount == 0:
		s.Status = "unreviewed"
	case concernCount >= 3 && concernCount > s.CredibleCount:
		s.Status = "community_suspicious"
	case s.CredibleCount >= 3 && s.CredibleCount >= concernCount*2:
		s.Status = "community_trusted"
	case s.CredibleCount > 0 && concernCount > 0:
		s.Status = "disputed"
	default:
		s.Status = "needs_more_context"
	}

	return s
}

type PostListFilter struct {
	Page         int
	Limit        int
	Offset       int
	Cursor       *PostCursor
	RankAt       time.Time
	ViewerUserID uint64
	Search       string
	CategorySlug string
	Feed         string
	Sort         string
	Latitude     float64
	Longitude    float64
	RadiusMeters int
}

type PostCursor struct {
	Sort      string
	Rank      float64
	RankAt    time.Time
	CreatedAt time.Time
	ID        uint64
}

type PostListPage struct {
	Items      []PostView `json:"items"`
	NextCursor string     `json:"next_cursor,omitempty"`
	HasMore    bool       `json:"has_more"`
}

type SavedCollection struct {
	ID        uint64              `json:"id"`
	UserID    uint64              `json:"user_id"`
	Name      SavedCollectionName `json:"-"`
	ShareSlug string              `json:"share_slug"`
	IsPublic  bool                `json:"is_public"`
	Posts     []Post              `json:"-"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

type SavedCollectionView struct {
	ID        uint64     `json:"id"`
	UserID    uint64     `json:"user_id"`
	Name      string     `json:"name"`
	ShareSlug string     `json:"share_slug"`
	IsPublic  bool       `json:"is_public"`
	Posts     []PostView `json:"posts"`
	PostCount int        `json:"post_count"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (c SavedCollection) View() SavedCollectionView {
	posts := make([]PostView, 0, len(c.Posts))
	for _, item := range c.Posts {
		posts = append(posts, item.View())
	}

	return SavedCollectionView{
		ID:        c.ID,
		UserID:    c.UserID,
		Name:      c.Name.String(),
		ShareSlug: c.ShareSlug,
		IsPublic:  c.IsPublic,
		Posts:     posts,
		PostCount: len(posts),
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

type NewSavedCollectionInput struct {
	UserID   uint64
	Name     string
	IsPublic bool
}

func NewSavedCollection(input NewSavedCollectionInput) (SavedCollection, error) {
	if input.UserID == 0 {
		return SavedCollection{}, ErrUserIDRequired
	}

	name, err := NewSavedCollectionName(input.Name)
	if err != nil {
		return SavedCollection{}, err
	}

	now := time.Now().UTC()
	return SavedCollection{
		UserID:    input.UserID,
		Name:      name,
		ShareSlug: GenerateSavedCollectionShareSlug(name.String()),
		IsPublic:  input.IsPublic,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

type Comment struct {
	ID               uint64    `json:"id"`
	PostID           uint64    `json:"post_id"`
	UserID           uint64    `json:"user_id"`
	UserName         string    `json:"user_name"`
	Content          Content   `json:"-"`
	ReplyToCommentID uint64    `json:"reply_to_comment_id,omitempty"`
	ReplyToUserID    uint64    `json:"reply_to_user_id,omitempty"`
	ReplyToUserName  string    `json:"reply_to_user_name,omitempty"`
	ReplyToContent   string    `json:"reply_to_content,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Status           string    `json:"status"`
}

type CommentView struct {
	ID               uint64    `json:"id"`
	PostID           uint64    `json:"post_id"`
	UserID           uint64    `json:"user_id"`
	UserName         string    `json:"user_name"`
	Content          string    `json:"content"`
	ReplyToCommentID uint64    `json:"reply_to_comment_id,omitempty"`
	ReplyToUserID    uint64    `json:"reply_to_user_id,omitempty"`
	ReplyToUserName  string    `json:"reply_to_user_name,omitempty"`
	ReplyToContent   string    `json:"reply_to_content,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Status           string    `json:"status"`
}

func (c Comment) View() CommentView {
	return CommentView{
		ID:               c.ID,
		PostID:           c.PostID,
		UserID:           c.UserID,
		UserName:         c.UserName,
		Content:          c.Content.String(),
		ReplyToCommentID: c.ReplyToCommentID,
		ReplyToUserID:    c.ReplyToUserID,
		ReplyToUserName:  c.ReplyToUserName,
		ReplyToContent:   c.ReplyToContent,
		CreatedAt:        c.CreatedAt,
		UpdatedAt:        c.UpdatedAt,
		Status:           c.Status,
	}
}

type NewCommentInput struct {
	PostID           uint64
	UserID           uint64
	Content          string
	ReplyToCommentID uint64
}

func NewComment(input NewCommentInput) (Comment, error) {
	if input.PostID == 0 {
		return Comment{}, ErrPostIDRequired
	}
	if input.UserID == 0 {
		return Comment{}, ErrUserIDRequired
	}

	content, err := NewContent(input.Content)
	if err != nil {
		return Comment{}, err
	}

	return Comment{
		PostID:           input.PostID,
		UserID:           input.UserID,
		Content:          content,
		ReplyToCommentID: input.ReplyToCommentID,
		CreatedAt:        time.Now().UTC(),
		Status:           "visible",
	}, nil
}

func NewPost(input NewPostInput) (Post, error) {
	if input.UserID == 0 {
		return Post{}, ErrUserIDRequired
	}
	if input.Latitude < -90 || input.Latitude > 90 {
		return Post{}, ErrLatitudeOutOfRange
	}
	if input.Longitude < -180 || input.Longitude > 180 {
		return Post{}, ErrLongitudeOutOfRange
	}

	imageURL, err := NewImageURL(input.ImageURL)
	if err != nil {
		return Post{}, err
	}
	caption, err := NewCaption(input.Caption)
	if err != nil {
		return Post{}, err
	}
	locationName, err := NewLocationName(input.LocationName)
	if err != nil {
		return Post{}, err
	}

	categoryIDs, err := NormalizeCategoryIDs(input.CategoryID, input.CategoryIDs)
	if err != nil {
		return Post{}, err
	}

	now := time.Now().UTC()
	return Post{
		UserID:       input.UserID,
		CategoryID:   firstCategoryID(categoryIDs),
		CategoryIDs:  categoryIDs,
		ImageURL:     imageURL,
		Caption:      caption,
		LocationName: locationName,
		Latitude:     input.Latitude,
		Longitude:    input.Longitude,
		CreatedAt:    now,
		UpdatedAt:    now,
		Status:       "visible",
	}, nil
}

func NewPostUpdate(input NewPostInput) (PostUpdate, error) {
	if input.Latitude < -90 || input.Latitude > 90 {
		return PostUpdate{}, ErrLatitudeOutOfRange
	}
	if input.Longitude < -180 || input.Longitude > 180 {
		return PostUpdate{}, ErrLongitudeOutOfRange
	}

	caption, err := NewCaption(input.Caption)
	if err != nil {
		return PostUpdate{}, err
	}
	locationName, err := NewLocationName(input.LocationName)
	if err != nil {
		return PostUpdate{}, err
	}

	categoryIDs, err := NormalizeCategoryIDs(input.CategoryID, input.CategoryIDs)
	if err != nil {
		return PostUpdate{}, err
	}

	return PostUpdate{
		Caption:      caption,
		LocationName: locationName,
		Latitude:     input.Latitude,
		Longitude:    input.Longitude,
		CategoryID:   firstCategoryID(categoryIDs),
		CategoryIDs:  categoryIDs,
	}, nil
}

func NormalizeCategoryIDs(categoryID uint64, categoryIDs []uint64) ([]uint64, error) {
	seen := make(map[uint64]struct{}, len(categoryIDs)+1)
	normalized := make([]uint64, 0, len(categoryIDs)+1)

	if categoryID != 0 {
		normalized = append(normalized, categoryID)
		seen[categoryID] = struct{}{}
	}
	for _, id := range categoryIDs {
		if id == 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		normalized = append(normalized, id)
		seen[id] = struct{}{}
	}
	if len(normalized) > 20 {
		return nil, ErrTooManyCategories
	}

	return normalized, nil
}

func firstCategoryID(categoryIDs []uint64) uint64 {
	if len(categoryIDs) == 0 {
		return 0
	}

	return categoryIDs[0]
}

func normalizedPostCategories(p Post) []PostCategory {
	if len(p.Categories) > 0 {
		return p.Categories
	}
	if p.CategoryID == 0 {
		return nil
	}

	return []PostCategory{{
		ID:   p.CategoryID,
		Name: p.CategoryName,
		Slug: p.CategorySlug,
	}}
}

type ImageURL string

func NewImageURL(raw string) (ImageURL, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", ErrImageURLRequired
	}

	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed == nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", ErrInvalidImageURL
	}

	return ImageURL(value), nil
}

func (i ImageURL) String() string {
	return string(i)
}

type Caption string

func NewCaption(raw string) (Caption, error) {
	value := strings.TrimSpace(raw)
	if len(value) > 2000 {
		return "", ErrCaptionTooLong
	}

	return Caption(value), nil
}

func (c Caption) String() string {
	return string(c)
}

type LocationName string

func NewLocationName(raw string) (LocationName, error) {
	value := strings.TrimSpace(raw)
	if len(value) > 255 {
		return "", ErrLocationNameTooLong
	}

	return LocationName(value), nil
}

func (l LocationName) String() string {
	return string(l)
}

type SavedCollectionName string

func NewSavedCollectionName(raw string) (SavedCollectionName, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", ErrCollectionNameRequired
	}
	if len(value) > 120 {
		return "", ErrCollectionNameTooLong
	}

	return SavedCollectionName(value), nil
}

func (n SavedCollectionName) String() string {
	return string(n)
}

func GenerateSavedCollectionShareSlug(name string) string {
	base := slugifySavedCollectionName(name)
	if base == "" {
		base = "collection"
	}

	return base + "-" + randomSlugSuffix()
}

func slugifySavedCollectionName(name string) string {
	var builder strings.Builder
	lastDash := false
	for _, value := range strings.ToLower(strings.TrimSpace(name)) {
		switch {
		case value <= 127 && (unicode.IsLetter(value) || unicode.IsDigit(value)):
			builder.WriteRune(value)
			lastDash = false
		case !lastDash && builder.Len() > 0:
			builder.WriteByte('-')
			lastDash = true
		}
	}

	return strings.Trim(builder.String(), "-")
}

func randomSlugSuffix() string {
	var value [4]byte
	if _, err := rand.Read(value[:]); err == nil {
		return hex.EncodeToString(value[:])
	}

	return strconv.FormatInt(time.Now().UTC().UnixNano(), 36)
}

type Content string

func NewContent(raw string) (Content, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", ErrCommentRequired
	}
	if len(value) > 1000 {
		return "", ErrCommentTooLong
	}

	return Content(value), nil
}

func (c Content) String() string {
	return string(c)
}

type ReportReason string

func NewReportReason(raw string) (ReportReason, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", ErrReportReasonRequired
	}
	if len(value) > 500 {
		return "", ErrReportReasonTooLong
	}

	return ReportReason(value), nil
}

func (r ReportReason) String() string {
	return string(r)
}
