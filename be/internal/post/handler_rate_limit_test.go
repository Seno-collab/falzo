package post

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"falzo-be/internal/auth"
	httpMiddleware "falzo-be/pkg/http/middleware"
)

type fakePostHandlerService struct {
	commentCalls int
	reportCalls  int
}

func (f *fakePostHandlerService) CreatePost(context.Context, CreatePostInput) (PostView, error) {
	return PostView{}, nil
}

func (f *fakePostHandlerService) UpdatePost(context.Context, UpdatePostInput) (PostView, error) {
	return PostView{}, nil
}

func (f *fakePostHandlerService) DeletePost(context.Context, ModerationInput) error { return nil }

func (f *fakePostHandlerService) HidePost(context.Context, ModerationInput) error { return nil }

func (f *fakePostHandlerService) ReportPost(context.Context, ReportInput) error {
	f.reportCalls++
	return nil
}

func (f *fakePostHandlerService) UpsertTrustVote(context.Context, TrustVoteInput) (PostTrustSummary, error) {
	return PostTrustSummary{}, nil
}

func (f *fakePostHandlerService) LikePost(context.Context, PostActionInput) error { return nil }

func (f *fakePostHandlerService) UnlikePost(context.Context, PostActionInput) error { return nil }

func (f *fakePostHandlerService) SavePost(context.Context, PostActionInput) error { return nil }

func (f *fakePostHandlerService) UnsavePost(context.Context, PostActionInput) error { return nil }

func (f *fakePostHandlerService) CreateSavedCollection(context.Context, CreateSavedCollectionInput) (SavedCollectionView, error) {
	return SavedCollectionView{}, nil
}

func (f *fakePostHandlerService) ListSavedCollections(context.Context, SavedCollectionInput) ([]SavedCollectionView, error) {
	return nil, nil
}

func (f *fakePostHandlerService) ListSavedPosts(context.Context, SavedCollectionInput) ([]PostView, error) {
	return nil, nil
}

func (f *fakePostHandlerService) AddPostToSavedCollection(context.Context, SavedCollectionPostInput) error {
	return nil
}

func (f *fakePostHandlerService) RemovePostFromSavedCollection(context.Context, SavedCollectionPostInput) error {
	return nil
}

func (f *fakePostHandlerService) DeleteSavedCollection(context.Context, SavedCollectionInput) error {
	return nil
}

func (f *fakePostHandlerService) UpdateSavedCollectionVisibility(context.Context, UpdateSavedCollectionVisibilityInput) (SavedCollectionView, error) {
	return SavedCollectionView{}, nil
}

func (f *fakePostHandlerService) GetPublicSavedCollection(context.Context, PublicSavedCollectionInput) (*SavedCollectionView, error) {
	return nil, nil
}

func (f *fakePostHandlerService) CommentPost(context.Context, CommentPostInput) (CommentView, error) {
	f.commentCalls++
	now := time.Now().UTC()
	return CommentView{ID: 1, PostID: 4, UserID: 7, Content: "hello", CreatedAt: now, UpdatedAt: now, Status: "visible"}, nil
}

func (f *fakePostHandlerService) UpdateComment(context.Context, UpdateCommentInput) (CommentView, error) {
	f.commentCalls++
	now := time.Now().UTC()
	return CommentView{ID: 2, PostID: 4, UserID: 7, Content: "updated", CreatedAt: now, UpdatedAt: now, Status: "visible"}, nil
}

func (f *fakePostHandlerService) DeleteComment(context.Context, ModerationInput) error { return nil }

func (f *fakePostHandlerService) HideComment(context.Context, ModerationInput) error { return nil }

func (f *fakePostHandlerService) ReportComment(context.Context, ReportInput) error {
	f.reportCalls++
	return nil
}

func (f *fakePostHandlerService) GetPosts(context.Context, ListPostsInput) (PostListPage, error) {
	return PostListPage{}, nil
}

func (f *fakePostHandlerService) GetPostDetail(context.Context, GetPostDetailInput) (*PostView, error) {
	return nil, nil
}

func (f *fakePostHandlerService) GetPostsByLocation(context.Context, GetPostsByLocationInput) ([]PostView, error) {
	return nil, nil
}

func (f *fakePostHandlerService) GetComments(context.Context, ListCommentsInput) ([]CommentView, error) {
	return nil, nil
}

type fakePostAuthService struct{}

func (fakePostAuthService) Authenticate(context.Context, string) (*auth.AuthenticatedUser, error) {
	return &auth.AuthenticatedUser{UserID: 7, Username: "tester"}, nil
}

func TestCommentRoutesCanBeRateLimited(t *testing.T) {
	service := &fakePostHandlerService{}
	limit := httpMiddleware.NewKeyedRateLimiter(1, time.Minute, func(r *http.Request) string {
		principal, ok := auth.AuthenticatedUserFromContext(r.Context())
		if !ok || principal == nil {
			return ""
		}
		return "user:" + principal.Username
	})
	handler := NewHandler(service, fakePostAuthService{}, WithCommentMiddlewares(limit))

	for i, expectedStatus := range []int{http.StatusCreated, http.StatusTooManyRequests} {
		req := httptest.NewRequest(http.MethodPost, "/4/comments", strings.NewReader(`{"content":"hello"}`))
		req.Header.Set("Authorization", "Bearer signed-token")
		rec := httptest.NewRecorder()

		handler.Routes().ServeHTTP(rec, req)

		if rec.Code != expectedStatus {
			t.Fatalf("request %d: expected status %d, got %d body=%s", i+1, expectedStatus, rec.Code, rec.Body.String())
		}
	}

	if service.commentCalls != 1 {
		t.Fatalf("expected service to receive one comment call, got %d", service.commentCalls)
	}
}

func TestReportRoutesCanBeRateLimited(t *testing.T) {
	service := &fakePostHandlerService{}
	limit := httpMiddleware.NewKeyedRateLimiter(1, time.Hour, func(r *http.Request) string {
		principal, ok := auth.AuthenticatedUserFromContext(r.Context())
		if !ok || principal == nil {
			return ""
		}
		return "user:" + principal.Username
	})
	handler := NewHandler(service, fakePostAuthService{}, WithReportMiddlewares(limit))

	for i, expectedStatus := range []int{http.StatusCreated, http.StatusTooManyRequests} {
		req := httptest.NewRequest(http.MethodPost, "/4/report", strings.NewReader(`{"reason":"spam"}`))
		req.Header.Set("Authorization", "Bearer signed-token")
		rec := httptest.NewRecorder()

		handler.Routes().ServeHTTP(rec, req)

		if rec.Code != expectedStatus {
			t.Fatalf("request %d: expected status %d, got %d body=%s", i+1, expectedStatus, rec.Code, rec.Body.String())
		}
	}

	if service.reportCalls != 1 {
		t.Fatalf("expected service to receive one report call, got %d", service.reportCalls)
	}
}
