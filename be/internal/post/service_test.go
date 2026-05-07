package post

import (
	"context"
	"testing"
)

type fakePostRepository struct {
	page  int
	limit int
}

func (f *fakePostRepository) Create(context.Context, *Post) error { return nil }

func (f *fakePostRepository) Like(context.Context, uint64, uint64) error { return nil }

func (f *fakePostRepository) Unlike(context.Context, uint64, uint64) error { return nil }

func (f *fakePostRepository) Save(context.Context, uint64, uint64) error { return nil }

func (f *fakePostRepository) Unsave(context.Context, uint64, uint64) error { return nil }

func (f *fakePostRepository) Comment(_ context.Context, comment *Comment) error {
	comment.ID = 10
	return nil
}

func (f *fakePostRepository) GetPosts(_ context.Context, page int, limit int, _ uint64) ([]Post, error) {
	f.page = page
	f.limit = limit
	return nil, nil
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

	_, err := service.GetPosts(t.Context(), ListPostsInput{Page: 2, Limit: 24})
	if err != nil {
		t.Fatalf("get posts: %v", err)
	}

	if repo.page != 2 || repo.limit != 24 {
		t.Fatalf("expected page=2 limit=24, got page=%d limit=%d", repo.page, repo.limit)
	}
}
