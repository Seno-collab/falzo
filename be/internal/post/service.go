package post

import (
	"context"
	"strings"
)

const maxPostListLimit = 50
const feedFollowing = "following"
const (
	postSortNewest   = "newest"
	postSortPopular  = "popular"
	postSortTrending = "trending"
	postSortNearby   = "nearby"
)

type Service struct {
	posts Repository
}

func NewService(posts Repository) *Service {
	return &Service{posts: posts}
}

type CreatePostInput = NewPostInput

type PostActionInput struct {
	PostID uint64
	UserID uint64
}

type ModerationInput struct {
	PostID    uint64
	CommentID uint64
	Actor     ModerationActor
	Reason    string
}

type ReportInput struct {
	PostID         uint64
	CommentID      uint64
	ReporterUserID uint64
	Reason         string
}

type CommentPostInput = NewCommentInput
type CreateSavedCollectionInput = NewSavedCollectionInput

type UpdatePostInput struct {
	PostID       uint64
	UserID       uint64
	CategoryID   uint64
	Caption      string
	LocationName string
	Latitude     float64
	Longitude    float64
}

type UpdateCommentInput struct {
	PostID    uint64
	CommentID uint64
	UserID    uint64
	Content   string
}

type ListPostsInput struct {
	Page         int
	Limit        int
	ViewerUserID uint64
	Search       string
	CategorySlug string
	Feed         string
	Sort         string
	Latitude     float64
	Longitude    float64
	RadiusMeters int
}

type ListCommentsInput struct {
	PostID uint64
	Page   int
	Limit  int
}

type GetPostDetailInput struct {
	PostID       uint64
	ViewerUserID uint64
}

type GetPostsByLocationInput struct {
	LocationName string
}

type SavedCollectionPostInput struct {
	CollectionID uint64
	PostID       uint64
	UserID       uint64
}

type SavedCollectionInput struct {
	CollectionID uint64
	UserID       uint64
}

func (s *Service) CreatePost(ctx context.Context, input CreatePostInput) (PostView, error) {
	if s.posts == nil {
		return PostView{}, ErrDependencyUnavailable
	}

	post, err := NewPost(NewPostInput(input))
	if err != nil {
		return PostView{}, err
	}

	if err := s.posts.Create(ctx, &post); err != nil {
		return PostView{}, err
	}

	return post.View(), nil
}

func (s *Service) UpdatePost(ctx context.Context, input UpdatePostInput) (PostView, error) {
	if s.posts == nil {
		return PostView{}, ErrDependencyUnavailable
	}
	if input.PostID == 0 {
		return PostView{}, ErrPostIDRequired
	}
	if input.UserID == 0 {
		return PostView{}, ErrUserIDRequired
	}

	update, err := NewPostUpdate(NewPostInput{
		CategoryID:   input.CategoryID,
		Caption:      input.Caption,
		LocationName: input.LocationName,
		Latitude:     input.Latitude,
		Longitude:    input.Longitude,
	})
	if err != nil {
		return PostView{}, err
	}

	item, err := s.posts.UpdatePost(ctx, input.PostID, input.UserID, update)
	if err != nil {
		return PostView{}, err
	}

	return item.View(), nil
}

func (s *Service) DeletePost(ctx context.Context, input ModerationInput) error {
	if err := s.validatePostModerationInput(input); err != nil {
		return err
	}

	return s.posts.DeletePost(ctx, input.PostID, input.Actor)
}

func (s *Service) HidePost(ctx context.Context, input ModerationInput) error {
	if err := s.validatePostModerationInput(input); err != nil {
		return err
	}

	reason, err := NewReportReason(input.Reason)
	if err != nil {
		return err
	}

	return s.posts.HidePost(ctx, input.PostID, input.Actor, reason)
}

func (s *Service) ReportPost(ctx context.Context, input ReportInput) error {
	report, err := s.newContentReport(input, false)
	if err != nil {
		return err
	}

	return s.posts.ReportPost(ctx, report)
}

func (s *Service) ReportComment(ctx context.Context, input ReportInput) error {
	report, err := s.newContentReport(input, true)
	if err != nil {
		return err
	}

	return s.posts.ReportComment(ctx, report)
}

func (s *Service) LikePost(ctx context.Context, input PostActionInput) error {
	return s.handlePostAction(ctx, input, func(ctx context.Context, postID uint64, userID uint64) error {
		return s.posts.Like(ctx, postID, userID)
	})
}

func (s *Service) UnlikePost(ctx context.Context, input PostActionInput) error {
	return s.handlePostAction(ctx, input, func(ctx context.Context, postID uint64, userID uint64) error {
		return s.posts.Unlike(ctx, postID, userID)
	})
}

func (s *Service) SavePost(ctx context.Context, input PostActionInput) error {
	return s.handlePostAction(ctx, input, func(ctx context.Context, postID uint64, userID uint64) error {
		return s.posts.Save(ctx, postID, userID)
	})
}

func (s *Service) UnsavePost(ctx context.Context, input PostActionInput) error {
	return s.handlePostAction(ctx, input, func(ctx context.Context, postID uint64, userID uint64) error {
		return s.posts.Unsave(ctx, postID, userID)
	})
}

func (s *Service) CreateSavedCollection(ctx context.Context, input CreateSavedCollectionInput) (SavedCollectionView, error) {
	if s.posts == nil {
		return SavedCollectionView{}, ErrDependencyUnavailable
	}

	collection, err := NewSavedCollection(NewSavedCollectionInput(input))
	if err != nil {
		return SavedCollectionView{}, err
	}

	if err := s.posts.CreateSavedCollection(ctx, &collection); err != nil {
		return SavedCollectionView{}, err
	}

	return collection.View(), nil
}

func (s *Service) ListSavedCollections(ctx context.Context, input SavedCollectionInput) ([]SavedCollectionView, error) {
	if s.posts == nil {
		return nil, ErrDependencyUnavailable
	}
	if input.UserID == 0 {
		return nil, ErrUserIDRequired
	}

	items, err := s.posts.ListSavedCollections(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	collections := make([]SavedCollectionView, 0, len(items))
	for _, item := range items {
		collections = append(collections, item.View())
	}

	return collections, nil
}

func (s *Service) ListSavedPosts(ctx context.Context, input SavedCollectionInput) ([]PostView, error) {
	if s.posts == nil {
		return nil, ErrDependencyUnavailable
	}
	if input.UserID == 0 {
		return nil, ErrUserIDRequired
	}

	items, err := s.posts.ListSavedPosts(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	posts := make([]PostView, 0, len(items))
	for _, item := range items {
		posts = append(posts, item.View())
	}

	return posts, nil
}

func (s *Service) AddPostToSavedCollection(ctx context.Context, input SavedCollectionPostInput) error {
	if err := s.validateSavedCollectionPostInput(input); err != nil {
		return err
	}

	return s.posts.AddPostToSavedCollection(ctx, input.CollectionID, input.PostID, input.UserID)
}

func (s *Service) RemovePostFromSavedCollection(ctx context.Context, input SavedCollectionPostInput) error {
	if err := s.validateSavedCollectionPostInput(input); err != nil {
		return err
	}

	return s.posts.RemovePostFromSavedCollection(ctx, input.CollectionID, input.PostID, input.UserID)
}

func (s *Service) DeleteSavedCollection(ctx context.Context, input SavedCollectionInput) error {
	if s.posts == nil {
		return ErrDependencyUnavailable
	}
	if input.UserID == 0 {
		return ErrUserIDRequired
	}
	if input.CollectionID == 0 {
		return ErrCollectionIDRequired
	}

	return s.posts.DeleteSavedCollection(ctx, input.CollectionID, input.UserID)
}

func (s *Service) validateSavedCollectionPostInput(input SavedCollectionPostInput) error {
	if s.posts == nil {
		return ErrDependencyUnavailable
	}
	if input.UserID == 0 {
		return ErrUserIDRequired
	}
	if input.CollectionID == 0 {
		return ErrCollectionIDRequired
	}
	if input.PostID == 0 {
		return ErrPostIDRequired
	}

	return nil
}

func (s *Service) handlePostAction(ctx context.Context, input PostActionInput, action func(context.Context, uint64, uint64) error) error {
	if s.posts == nil {
		return ErrDependencyUnavailable
	}
	if action == nil {
		return ErrDependencyUnavailable
	}
	if input.PostID == 0 {
		return ErrPostIDRequired
	}
	if input.UserID == 0 {
		return ErrUserIDRequired
	}

	return action(ctx, input.PostID, input.UserID)
}

func (s *Service) CommentPost(ctx context.Context, input CommentPostInput) (CommentView, error) {
	if s.posts == nil {
		return CommentView{}, ErrDependencyUnavailable
	}

	comment, err := NewComment(NewCommentInput(input))
	if err != nil {
		return CommentView{}, err
	}

	if err := s.posts.Comment(ctx, &comment); err != nil {
		return CommentView{}, err
	}

	return comment.View(), nil
}

func (s *Service) UpdateComment(ctx context.Context, input UpdateCommentInput) (CommentView, error) {
	if s.posts == nil {
		return CommentView{}, ErrDependencyUnavailable
	}
	if input.PostID == 0 {
		return CommentView{}, ErrPostIDRequired
	}
	if input.CommentID == 0 {
		return CommentView{}, ErrCommentNotFound
	}
	if input.UserID == 0 {
		return CommentView{}, ErrUserIDRequired
	}

	content, err := NewContent(input.Content)
	if err != nil {
		return CommentView{}, err
	}

	comment, err := s.posts.UpdateComment(ctx, input.PostID, input.CommentID, input.UserID, content)
	if err != nil {
		return CommentView{}, err
	}

	return comment.View(), nil
}

func (s *Service) DeleteComment(ctx context.Context, input ModerationInput) error {
	if err := s.validateCommentModerationInput(input); err != nil {
		return err
	}

	return s.posts.DeleteComment(ctx, input.PostID, input.CommentID, input.Actor)
}

func (s *Service) HideComment(ctx context.Context, input ModerationInput) error {
	if err := s.validateCommentModerationInput(input); err != nil {
		return err
	}

	reason, err := NewReportReason(input.Reason)
	if err != nil {
		return err
	}

	return s.posts.HideComment(ctx, input.PostID, input.CommentID, input.Actor, reason)
}

func (s *Service) GetPosts(ctx context.Context, input ListPostsInput) ([]PostView, error) {
	if s.posts == nil {
		return nil, ErrDependencyUnavailable
	}
	if input.Page <= 0 {
		return nil, ErrPageMustBePositive
	}
	if input.Limit <= 0 {
		return nil, ErrLimitMustBePositive
	}
	if input.Limit > maxPostListLimit {
		return nil, ErrLimitTooLarge
	}
	feed := strings.TrimSpace(input.Feed)
	if feed != "" && feed != feedFollowing {
		return nil, ErrInvalidFeed
	}
	if feed == feedFollowing && input.ViewerUserID == 0 {
		return nil, ErrUserIDRequired
	}
	sort := strings.TrimSpace(input.Sort)
	if sort == "" {
		sort = postSortNewest
	}
	if sort != postSortNewest && sort != postSortPopular && sort != postSortTrending && sort != postSortNearby {
		return nil, ErrInvalidPostSort
	}
	if sort == postSortNearby && (input.Latitude < -90 || input.Latitude > 90 || input.Longitude < -180 || input.Longitude > 180) {
		return nil, ErrNearbyCoordinatesRequired
	}

	items, err := s.posts.GetPosts(ctx, PostListFilter{
		Page:         input.Page,
		Limit:        input.Limit,
		ViewerUserID: input.ViewerUserID,
		Search:       strings.TrimSpace(input.Search),
		CategorySlug: strings.TrimSpace(input.CategorySlug),
		Feed:         feed,
		Sort:         sort,
		Latitude:     input.Latitude,
		Longitude:    input.Longitude,
		RadiusMeters: input.RadiusMeters,
	})
	if err != nil {
		return nil, err
	}

	posts := make([]PostView, 0, len(items))
	for _, item := range items {
		posts = append(posts, item.View())
	}

	return posts, nil
}

func (s *Service) validatePostModerationInput(input ModerationInput) error {
	if s.posts == nil {
		return ErrDependencyUnavailable
	}
	if input.Actor.UserID == 0 {
		return ErrUserIDRequired
	}
	if input.PostID == 0 {
		return ErrPostIDRequired
	}

	return nil
}

func (s *Service) validateCommentModerationInput(input ModerationInput) error {
	if err := s.validatePostModerationInput(input); err != nil {
		return err
	}
	if input.CommentID == 0 {
		return ErrCommentNotFound
	}

	return nil
}

func (s *Service) newContentReport(input ReportInput, requireComment bool) (ContentReport, error) {
	if s.posts == nil {
		return ContentReport{}, ErrDependencyUnavailable
	}
	if input.ReporterUserID == 0 {
		return ContentReport{}, ErrUserIDRequired
	}
	if input.PostID == 0 {
		return ContentReport{}, ErrPostIDRequired
	}
	if requireComment && input.CommentID == 0 {
		return ContentReport{}, ErrCommentNotFound
	}

	reason, err := NewReportReason(input.Reason)
	if err != nil {
		return ContentReport{}, err
	}

	return ContentReport{
		ReporterUserID: input.ReporterUserID,
		PostID:         input.PostID,
		CommentID:      input.CommentID,
		Reason:         reason,
	}, nil
}

func (s *Service) GetPostDetail(ctx context.Context, input GetPostDetailInput) (*PostView, error) {
	if s.posts == nil {
		return nil, ErrDependencyUnavailable
	}
	if input.PostID == 0 {
		return nil, ErrPostIDRequired
	}

	item, err := s.posts.GetPostDetail(ctx, input.PostID, input.ViewerUserID)
	if err != nil {
		return nil, err
	}

	mapped := item.View()
	return &mapped, nil
}

func (s *Service) GetComments(ctx context.Context, input ListCommentsInput) ([]CommentView, error) {
	if s.posts == nil {
		return nil, ErrDependencyUnavailable
	}
	if input.PostID == 0 {
		return nil, ErrPostIDRequired
	}
	if input.Page <= 0 {
		return nil, ErrPageMustBePositive
	}
	if input.Limit <= 0 {
		return nil, ErrLimitMustBePositive
	}
	if input.Limit > maxPostListLimit {
		return nil, ErrLimitTooLarge
	}

	items, err := s.posts.GetComments(ctx, input.PostID, input.Page, input.Limit)
	if err != nil {
		return nil, err
	}

	comments := make([]CommentView, 0, len(items))
	for _, item := range items {
		comments = append(comments, item.View())
	}

	return comments, nil
}

func (s *Service) GetPostsByLocation(ctx context.Context, input GetPostsByLocationInput) ([]PostView, error) {
	if s.posts == nil {
		return nil, ErrDependencyUnavailable
	}
	if strings.TrimSpace(input.LocationName) == "" {
		return nil, ErrLocationNameRequired
	}

	locationName, err := NewLocationName(input.LocationName)
	if err != nil {
		return nil, err
	}

	items, err := s.posts.GetPostsByLocation(ctx, locationName)
	if err != nil {
		return nil, err
	}

	posts := make([]PostView, 0, len(items))
	for _, item := range items {
		posts = append(posts, item.View())
	}

	return posts, nil
}
