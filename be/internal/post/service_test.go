package post

import (
	"context"
	"testing"
	"time"
)

type fakePostRepository struct {
	commentReplyToID uint64
	createCategoryID uint64
	updateUserID     uint64
	page             int
	limit            int
	offset           int
	cursor           *PostCursor
	rankAt           time.Time
	search           string
	categorySlug     string
	feed             string
	sort             string
	radiusMeters     int
	posts            []Post
	trustVote        TrustVote
}

func (f *fakePostRepository) Create(_ context.Context, post *Post) error {
	f.createCategoryID = post.CategoryID
	return nil
}

func (f *fakePostRepository) UpdatePost(_ context.Context, postID uint64, userID uint64, update PostUpdate) (Post, error) {
	return Post{ID: postID, UserID: userID, Caption: update.Caption, LocationName: update.LocationName, Latitude: update.Latitude, Longitude: update.Longitude, Status: "visible"}, nil
}

func (f *fakePostRepository) DeletePost(context.Context, uint64, ModerationActor) error { return nil }

func (f *fakePostRepository) HidePost(context.Context, uint64, ModerationActor, ReportReason) error {
	return nil
}

func (f *fakePostRepository) ReportPost(context.Context, ContentReport) error { return nil }

func (f *fakePostRepository) ReportComment(context.Context, ContentReport) error { return nil }

func (f *fakePostRepository) UpsertTrustVote(_ context.Context, vote TrustVote) (PostTrustSummary, error) {
	f.trustVote = vote
	return PostTrustSummary{
		CredibleCount: 1,
		ViewerVote:    string(vote.Type),
	}.WithStatus(), nil
}

func (f *fakePostRepository) DeleteComment(context.Context, uint64, uint64, ModerationActor) error {
	return nil
}

func (f *fakePostRepository) HideComment(context.Context, uint64, uint64, ModerationActor, ReportReason) error {
	return nil
}

func (f *fakePostRepository) Like(context.Context, uint64, uint64) error { return nil }

func (f *fakePostRepository) Unlike(context.Context, uint64, uint64) error { return nil }

func (f *fakePostRepository) Save(context.Context, uint64, uint64) error { return nil }

func (f *fakePostRepository) Unsave(context.Context, uint64, uint64) error { return nil }

func (f *fakePostRepository) CreateSavedCollection(_ context.Context, collection *SavedCollection) error {
	collection.ID = 4
	return nil
}

func (f *fakePostRepository) ListSavedCollections(context.Context, uint64) ([]SavedCollection, error) {
	return nil, nil
}

func (f *fakePostRepository) ListSavedPosts(context.Context, uint64) ([]Post, error) {
	return nil, nil
}

func (f *fakePostRepository) AddPostToSavedCollection(context.Context, uint64, uint64, uint64) error {
	return nil
}

func (f *fakePostRepository) RemovePostFromSavedCollection(context.Context, uint64, uint64, uint64) error {
	return nil
}

func (f *fakePostRepository) DeleteSavedCollection(context.Context, uint64, uint64) error {
	return nil
}

func (f *fakePostRepository) UpdateSavedCollectionVisibility(_ context.Context, collectionID uint64, userID uint64, isPublic bool) (SavedCollection, error) {
	name, _ := NewSavedCollectionName("Shared route")
	return SavedCollection{
		ID:        collectionID,
		UserID:    userID,
		Name:      name,
		ShareSlug: "shared-route-test",
		IsPublic:  isPublic,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}, nil
}

func (f *fakePostRepository) GetPublicSavedCollection(_ context.Context, shareSlug string, _ uint64) (*SavedCollection, error) {
	name, _ := NewSavedCollectionName("Shared route")
	return &SavedCollection{
		ID:        9,
		UserID:    2,
		Name:      name,
		ShareSlug: shareSlug,
		IsPublic:  true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}, nil
}

func (f *fakePostRepository) Comment(_ context.Context, comment *Comment) error {
	f.commentReplyToID = comment.ReplyToCommentID
	comment.ID = 10
	return nil
}

func (f *fakePostRepository) UpdateComment(_ context.Context, postID uint64, commentID uint64, userID uint64, content Content) (Comment, error) {
	f.updateUserID = userID
	return Comment{
		ID:        commentID,
		PostID:    postID,
		UserID:    userID,
		UserName:  "tester",
		Content:   content,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (f *fakePostRepository) GetPosts(_ context.Context, filter PostListFilter) ([]Post, error) {
	f.page = filter.Page
	f.limit = filter.Limit
	f.offset = filter.Offset
	f.cursor = filter.Cursor
	f.rankAt = filter.RankAt
	f.search = filter.Search
	f.categorySlug = filter.CategorySlug
	f.feed = filter.Feed
	f.sort = filter.Sort
	f.radiusMeters = filter.RadiusMeters
	return f.posts, nil
}

func TestCreatePostPassesCategoryToRepository(t *testing.T) {
	repo := &fakePostRepository{}
	service := NewService(repo)

	_, err := service.CreatePost(t.Context(), CreatePostInput{
		UserID:       2,
		CategoryID:   4,
		ImageURL:     "https://example.com/image.jpg",
		Caption:      "Category post",
		LocationName: "Da Nang",
		Latitude:     16.0471,
		Longitude:    108.2068,
	})
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	if repo.createCategoryID != 4 {
		t.Fatalf("expected category id 4, got %d", repo.createCategoryID)
	}
}

func TestCommentPostPassesReplyTargetToRepository(t *testing.T) {
	repo := &fakePostRepository{}
	service := NewService(repo)

	comment, err := service.CommentPost(t.Context(), CommentPostInput{
		PostID:           1,
		UserID:           2,
		Content:          "Nice image",
		ReplyToCommentID: 9,
	})
	if err != nil {
		t.Fatalf("comment post reply: %v", err)
	}

	if comment.ReplyToCommentID != 9 || repo.commentReplyToID != 9 {
		t.Fatalf("expected reply target 9, got view=%d repo=%d", comment.ReplyToCommentID, repo.commentReplyToID)
	}
}

func TestUpdateComment(t *testing.T) {
	repo := &fakePostRepository{}
	service := NewService(repo)

	comment, err := service.UpdateComment(t.Context(), UpdateCommentInput{
		PostID:    1,
		CommentID: 10,
		UserID:    2,
		Content:   "Updated message",
	})
	if err != nil {
		t.Fatalf("update comment: %v", err)
	}

	if comment.ID != 10 || comment.Content != "Updated message" {
		t.Fatalf("expected updated comment view, got %+v", comment)
	}
	if repo.updateUserID != 2 {
		t.Fatalf("expected update to use authenticated user 2, got %d", repo.updateUserID)
	}
}

func (f *fakePostRepository) GetPostDetail(context.Context, uint64, uint64) (*Post, error) {
	return nil, nil
}

func (f *fakePostRepository) GetPostsByLocation(context.Context, LocationName) ([]Post, error) {
	return nil, nil
}

func (f *fakePostRepository) GetComments(context.Context, uint64, int, int) ([]Comment, error) {
	return nil, nil
}

func TestCommentPost(t *testing.T) {
	service := NewService(&fakePostRepository{})

	comment, err := service.CommentPost(t.Context(), CommentPostInput{
		PostID:  1,
		UserID:  2,
		Content: "Nice image",
	})
	if err != nil {
		t.Fatalf("comment post: %v", err)
	}

	if comment.ID != 10 || comment.Content != "Nice image" {
		t.Fatalf("expected persisted comment view, got %+v", comment)
	}
}

func TestUpsertTrustVoteValidatesAndReturnsSummary(t *testing.T) {
	repo := &fakePostRepository{}
	service := NewService(repo)

	summary, err := service.UpsertTrustVote(t.Context(), TrustVoteInput{
		PostID: 1,
		UserID: 2,
		Type:   "credible",
		Reason: "I visited this place.",
	})
	if err != nil {
		t.Fatalf("upsert trust vote: %v", err)
	}

	if repo.trustVote.PostID != 1 || repo.trustVote.UserID != 2 || repo.trustVote.Type != TrustVoteCredible {
		t.Fatalf("expected trusted vote to reach repo, got %+v", repo.trustVote)
	}
	if summary.ViewerVote != "credible" || summary.Status == "" {
		t.Fatalf("expected summary with viewer vote and status, got %+v", summary)
	}
}

func TestUpsertTrustVoteRejectsInvalidType(t *testing.T) {
	service := NewService(&fakePostRepository{})

	_, err := service.UpsertTrustVote(t.Context(), TrustVoteInput{
		PostID: 1,
		UserID: 2,
		Type:   "fake",
	})
	if err != ErrInvalidTrustVoteType {
		t.Fatalf("expected ErrInvalidTrustVoteType, got %v", err)
	}
}

func TestGetPostsRejectsOversizedLimit(t *testing.T) {
	service := NewService(&fakePostRepository{})

	_, err := service.GetPosts(t.Context(), ListPostsInput{Page: 1, Limit: maxPostListLimit + 1})
	if err != ErrLimitTooLarge {
		t.Fatalf("expected ErrLimitTooLarge, got %v", err)
	}
}

func TestGetPostsPassesPaginationToRepository(t *testing.T) {
	repo := &fakePostRepository{}
	service := NewService(repo)

	_, err := service.GetPosts(t.Context(), ListPostsInput{Page: 2, Limit: 24, Search: " Kyoto ", CategorySlug: " heritage ", Feed: "following", ViewerUserID: 7})
	if err != nil {
		t.Fatalf("get posts: %v", err)
	}

	if repo.page != 2 || repo.limit != 25 || repo.offset != 24 || repo.search != "Kyoto" || repo.categorySlug != "heritage" || repo.feed != "following" {
		t.Fatalf("expected page=2 repository limit=25 offset=24 search=Kyoto category=heritage feed=following, got page=%d limit=%d offset=%d search=%q category=%q feed=%q", repo.page, repo.limit, repo.offset, repo.search, repo.categorySlug, repo.feed)
	}
}

func TestGetPostsReturnsCursorPage(t *testing.T) {
	imageURL, _ := NewImageURL("https://example.com/image.jpg")
	caption, _ := NewCaption("A scenic post")
	locationName, _ := NewLocationName("Da Nang")
	now := time.Now().UTC()
	repo := &fakePostRepository{
		posts: []Post{
			{ID: 3, ImageURL: imageURL, Caption: caption, LocationName: locationName, CreatedAt: now.Add(-1 * time.Minute)},
			{ID: 2, ImageURL: imageURL, Caption: caption, LocationName: locationName, CreatedAt: now.Add(-2 * time.Minute)},
			{ID: 1, ImageURL: imageURL, Caption: caption, LocationName: locationName, CreatedAt: now.Add(-3 * time.Minute)},
		},
	}
	service := NewService(repo)

	page, err := service.GetPosts(t.Context(), ListPostsInput{Page: 1, Limit: 2, Sort: "newest"})
	if err != nil {
		t.Fatalf("get posts: %v", err)
	}

	if len(page.Items) != 2 || !page.HasMore || page.NextCursor == "" {
		t.Fatalf("expected cursor page with 2 items and next cursor, got %+v", page)
	}
	cursor, err := decodePostCursor(page.NextCursor)
	if err != nil {
		t.Fatalf("decode next cursor: %v", err)
	}
	if cursor.ID != 2 || cursor.Sort != "newest" {
		t.Fatalf("expected cursor for last visible post, got %+v", cursor)
	}
	if cursor.RankAt.IsZero() {
		t.Fatal("expected next cursor to keep rank_at")
	}
}

func TestGetPostsPassesDecodedCursorToRepository(t *testing.T) {
	cursorValue := encodePostCursor(PostCursor{
		Sort:      "newest",
		RankAt:    time.Now().UTC().Add(-1 * time.Minute),
		CreatedAt: time.Now().UTC(),
		ID:        12,
	})
	repo := &fakePostRepository{}
	service := NewService(repo)

	_, err := service.GetPosts(t.Context(), ListPostsInput{Page: 1, Limit: 24, Sort: "newest", Cursor: cursorValue})
	if err != nil {
		t.Fatalf("get posts with cursor: %v", err)
	}

	if repo.cursor == nil || repo.cursor.ID != 12 {
		t.Fatalf("expected decoded cursor to reach repository, got %+v", repo.cursor)
	}
	if repo.offset != 0 {
		t.Fatalf("expected cursor request to use offset 0, got %d", repo.offset)
	}
	if !repo.rankAt.Equal(repo.cursor.RankAt) {
		t.Fatalf("expected repository rank_at to come from cursor, got %v want %v", repo.rankAt, repo.cursor.RankAt)
	}
}

func TestGetFollowingFeedRequiresViewer(t *testing.T) {
	service := NewService(&fakePostRepository{})

	_, err := service.GetPosts(t.Context(), ListPostsInput{Page: 1, Limit: 24, Feed: "following"})
	if err != ErrUserIDRequired {
		t.Fatalf("expected ErrUserIDRequired, got %v", err)
	}
}

func TestGetPostsRejectsNearbyRadiusAbove1000Km(t *testing.T) {
	service := NewService(&fakePostRepository{})

	_, err := service.GetPosts(t.Context(), ListPostsInput{
		Page:         1,
		Limit:        24,
		Sort:         "nearby",
		Latitude:     10.7769,
		Longitude:    106.7009,
		RadiusMeters: maxNearbyRadiusMeters + 1,
	})
	if err != ErrNearbyRadiusTooLarge {
		t.Fatalf("expected ErrNearbyRadiusTooLarge, got %v", err)
	}
}

func TestCreateSavedCollectionTrimsName(t *testing.T) {
	service := NewService(&fakePostRepository{})

	collection, err := service.CreateSavedCollection(t.Context(), CreateSavedCollectionInput{
		UserID: 2,
		Name:   " Đà Lạt trip ",
	})
	if err != nil {
		t.Fatalf("create saved collection: %v", err)
	}

	if collection.ID != 4 || collection.Name != "Đà Lạt trip" {
		t.Fatalf("expected trimmed collection view, got %+v", collection)
	}
}

func TestCreateSavedCollectionRequiresName(t *testing.T) {
	service := NewService(&fakePostRepository{})

	_, err := service.CreateSavedCollection(t.Context(), CreateSavedCollectionInput{UserID: 2})
	if err != ErrCollectionNameRequired {
		t.Fatalf("expected ErrCollectionNameRequired, got %v", err)
	}
}

func TestAddPostToSavedCollectionValidatesIDs(t *testing.T) {
	service := NewService(&fakePostRepository{})

	err := service.AddPostToSavedCollection(t.Context(), SavedCollectionPostInput{
		UserID: 2,
		PostID: 3,
	})
	if err != ErrCollectionIDRequired {
		t.Fatalf("expected ErrCollectionIDRequired, got %v", err)
	}
}

func TestUpdateSavedCollectionVisibility(t *testing.T) {
	service := NewService(&fakePostRepository{})

	collection, err := service.UpdateSavedCollectionVisibility(t.Context(), UpdateSavedCollectionVisibilityInput{
		CollectionID: 4,
		UserID:       2,
		IsPublic:     true,
	})
	if err != nil {
		t.Fatalf("update saved collection visibility: %v", err)
	}

	if !collection.IsPublic || collection.ShareSlug == "" {
		t.Fatalf("expected public collection with share slug, got %+v", collection)
	}
}

func TestGetPublicSavedCollectionRequiresSlug(t *testing.T) {
	service := NewService(&fakePostRepository{})

	_, err := service.GetPublicSavedCollection(t.Context(), PublicSavedCollectionInput{})
	if err != ErrCollectionSlugRequired {
		t.Fatalf("expected ErrCollectionSlugRequired, got %v", err)
	}
}
