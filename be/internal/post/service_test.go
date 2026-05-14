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
	search           string
	categorySlug     string
	feed             string
}

func (f *fakePostRepository) Create(_ context.Context, post *Post) error {
	f.createCategoryID = post.CategoryID
	return nil
}

func (f *fakePostRepository) Like(context.Context, uint64, uint64) error { return nil }

func (f *fakePostRepository) Unlike(context.Context, uint64, uint64) error { return nil }

func (f *fakePostRepository) Save(context.Context, uint64, uint64) error { return nil }

func (f *fakePostRepository) Unsave(context.Context, uint64, uint64) error { return nil }

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

func (f *fakePostRepository) GetPosts(_ context.Context, page int, limit int, _ uint64, search string, categorySlug string, feed string) ([]Post, error) {
	f.page = page
	f.limit = limit
	f.search = search
	f.categorySlug = categorySlug
	f.feed = feed
	return nil, nil
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

	if repo.page != 2 || repo.limit != 24 || repo.search != "Kyoto" || repo.categorySlug != "heritage" || repo.feed != "following" {
		t.Fatalf("expected page=2 limit=24 search=Kyoto category=heritage feed=following, got page=%d limit=%d search=%q category=%q feed=%q", repo.page, repo.limit, repo.search, repo.categorySlug, repo.feed)
	}
}

func TestGetFollowingFeedRequiresViewer(t *testing.T) {
	service := NewService(&fakePostRepository{})

	_, err := service.GetPosts(t.Context(), ListPostsInput{Page: 1, Limit: 24, Feed: "following"})
	if err != ErrUserIDRequired {
		t.Fatalf("expected ErrUserIDRequired, got %v", err)
	}
}
