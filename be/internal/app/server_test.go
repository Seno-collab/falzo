package app

import (
	"context"
	"testing"

	"falzo-be/internal/auth"
	"falzo-be/internal/post"
)

type fakePublicProfileInvalidator struct {
	userID uint64
	called bool
}

func (f *fakePublicProfileInvalidator) InvalidatePublicProfile(_ context.Context, userID uint64) {
	f.userID = userID
	f.called = true
}

type fakePublicFeedInvalidator struct {
	called bool
}

func (f *fakePublicFeedInvalidator) InvalidatePublicFeed(_ context.Context) {
	f.called = true
}

type fakePostEventPublisher struct {
	event  post.UserAvatarUpdatedEvent
	called bool
}

func (f *fakePostEventPublisher) PublishPostCreated(context.Context, post.PostView) error {
	return nil
}

func (f *fakePostEventPublisher) PublishPostDeleted(context.Context, uint64) error {
	return nil
}

func (f *fakePostEventPublisher) PublishUserAvatarUpdated(_ context.Context, event post.UserAvatarUpdatedEvent) error {
	f.event = event
	f.called = true
	return nil
}

func TestAuthAvatarEventPublisherInvalidatesCachesAndPublishesPostEvent(t *testing.T) {
	profiles := &fakePublicProfileInvalidator{}
	postFeeds := &fakePublicFeedInvalidator{}
	posts := &fakePostEventPublisher{}
	publisher := authAvatarEventPublisher{
		posts:     posts,
		postFeeds: postFeeds,
		profiles:  profiles,
	}

	err := publisher.PublishAvatarUpdated(context.Background(), auth.AvatarUpdatedEvent{
		UserID:         42,
		AvatarURL:      "https://cdn.example.com/avatar.png",
		AvatarURLAlias: "https://cdn.example.com/avatar.png",
	})
	if err != nil {
		t.Fatalf("publish avatar updated: %v", err)
	}
	if !profiles.called || profiles.userID != 42 {
		t.Fatalf("expected profile cache invalidation for user 42, got called=%v userID=%d", profiles.called, profiles.userID)
	}
	if !postFeeds.called {
		t.Fatal("expected public post feed cache invalidation")
	}
	if !posts.called {
		t.Fatal("expected post avatar update event to be published")
	}
	if posts.event.UserID != 42 || posts.event.AvatarURL != "https://cdn.example.com/avatar.png" {
		t.Fatalf("unexpected post avatar event: %+v", posts.event)
	}
}
