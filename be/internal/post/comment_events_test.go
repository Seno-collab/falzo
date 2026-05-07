package post

import (
	"testing"
	"time"
)

func TestCommentEventBrokerPublishesToMatchingPostSubscribers(t *testing.T) {
	broker := NewCommentEventBroker()
	matching, unsubscribeMatching := broker.SubscribeComments(t.Context(), 7)
	defer unsubscribeMatching()
	other, unsubscribeOther := broker.SubscribeComments(t.Context(), 8)
	defer unsubscribeOther()

	comment := CommentView{ID: 12, PostID: 7, UserID: 3, Content: "Fresh comment"}
	if err := broker.PublishCommentCreated(t.Context(), comment); err != nil {
		t.Fatalf("publish comment created: %v", err)
	}

	select {
	case got := <-matching:
		if got != comment {
			t.Fatalf("expected matching subscriber to receive %+v, got %+v", comment, got)
		}
	case <-time.After(time.Second):
		t.Fatal("expected matching subscriber to receive comment event")
	}

	select {
	case got := <-other:
		t.Fatalf("expected other post subscriber to stay idle, got %+v", got)
	case <-time.After(25 * time.Millisecond):
	}
}

func TestCommentEventBrokerPublishesToMultipleSubscribersForSamePost(t *testing.T) {
	broker := NewCommentEventBroker()
	first, unsubscribeFirst := broker.SubscribeComments(t.Context(), 9)
	defer unsubscribeFirst()
	second, unsubscribeSecond := broker.SubscribeComments(t.Context(), 9)
	defer unsubscribeSecond()

	comment := CommentView{ID: 21, PostID: 9, UserID: 4, Content: "Shared post comment"}
	if err := broker.PublishCommentCreated(t.Context(), comment); err != nil {
		t.Fatalf("publish comment created: %v", err)
	}

	for label, ch := range map[string]<-chan CommentView{
		"first":  first,
		"second": second,
	} {
		select {
		case got := <-ch:
			if got != comment {
				t.Fatalf("expected %s subscriber to receive %+v, got %+v", label, comment, got)
			}
		case <-time.After(time.Second):
			t.Fatalf("expected %s subscriber to receive comment event", label)
		}
	}
}
