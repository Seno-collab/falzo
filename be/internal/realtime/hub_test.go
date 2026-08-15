package realtime

import (
	domainchat "be/internal/domain/chat"
	"context"
	"errors"
	"testing"
	"time"
)

type recordingChatStore struct{ messages []domainchat.Message }

func (s *recordingChatStore) SaveMessage(_ context.Context, message domainchat.Message) error {
	s.messages = append(s.messages, message)
	return nil
}

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time { return c.now }

func TestPublishChatRejectsEliminatedMember(t *testing.T) {
	hub := NewHub(fixedClock{now: time.Date(2026, time.August, 14, 12, 0, 0, 0, time.UTC)})
	client, err := hub.Register("room-1", 7, "spectator", []Member{{
		ID:         7,
		Name:       "spectator",
		SeatNumber: 1,
		Eliminated: true,
	}})
	if err != nil {
		t.Fatalf("register client: %v", err)
	}
	defer hub.Unregister(client)

	if err := hub.PublishChat(client, "hello"); !errors.Is(err, ErrSpectator) {
		t.Fatalf("PublishChat() error = %v, want %v", err, ErrSpectator)
	}
}

func TestUpdateMembersRevokesChatAfterElimination(t *testing.T) {
	hub := NewHub(fixedClock{now: time.Date(2026, time.August, 14, 12, 0, 0, 0, time.UTC)})
	members := []Member{{ID: 9, Name: "player", SeatNumber: 1}}
	client, err := hub.Register("room-2", 9, "player", members)
	if err != nil {
		t.Fatalf("register client: %v", err)
	}
	defer hub.Unregister(client)

	if err := hub.PublishChat(client, "before vote"); err != nil {
		t.Fatalf("PublishChat() before elimination error = %v", err)
	}
	hub.UpdateMembers("room-2", []Member{{
		ID:         9,
		Name:       "player",
		SeatNumber: 1,
		Eliminated: true,
	}})
	if err := hub.PublishChat(client, "after vote"); !errors.Is(err, ErrSpectator) {
		t.Fatalf("PublishChat() after elimination error = %v, want %v", err, ErrSpectator)
	}
}

func TestRegisterReplacesExistingConnectionForSameRoomAndUser(t *testing.T) {
	hub := NewHub(fixedClock{now: time.Date(2026, time.August, 14, 12, 0, 0, 0, time.UTC)})
	members := []Member{{ID: 12, Name: "player", SeatNumber: 1}}
	first, err := hub.Register("room-replace", 12, "player", members)
	if err != nil {
		t.Fatalf("register first client: %v", err)
	}
	second, err := hub.Register("room-replace", 12, "player", members)
	if err != nil {
		t.Fatalf("register second client: %v", err)
	}
	defer hub.Unregister(first)
	defer hub.Unregister(second)

	select {
	case <-first.Done():
		if !errors.Is(first.CloseReason(), ErrConnectionReplaced) {
			t.Fatalf("first close reason = %v, want %v", first.CloseReason(), ErrConnectionReplaced)
		}
	case <-time.After(time.Second):
		t.Fatal("first connection was not replaced")
	}
	if err := hub.PublishChat(first, "stale connection"); !errors.Is(err, ErrClientNotFound) {
		t.Fatalf("stale PublishChat() error = %v, want %v", err, ErrClientNotFound)
	}
	if err := hub.PublishChat(second, "active connection"); err != nil {
		t.Fatalf("active PublishChat() error = %v", err)
	}
}

func TestEvictRoomMemberStopsTheirConnectionAndRevokesMembership(t *testing.T) {
	hub := NewHub(fixedClock{now: time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)})
	members := []Member{
		{ID: 61, Name: "host", SeatNumber: 1, Host: true},
		{ID: 62, Name: "member", SeatNumber: 2},
	}
	client, err := hub.Register("room-kick", 62, "member", members)
	if err != nil {
		t.Fatal(err)
	}
	defer hub.Unregister(client)

	hub.EvictRoomMember("room-kick", 62)

	select {
	case <-client.Done():
		if !errors.Is(client.CloseReason(), ErrRoomMemberRemoved) {
			t.Fatalf("close reason = %v, want %v", client.CloseReason(), ErrRoomMemberRemoved)
		}
	case <-time.After(time.Second):
		t.Fatal("evicted room member connection stayed open")
	}
	if err := hub.PublishChat(client, "still here"); !errors.Is(err, ErrClientNotFound) {
		t.Fatalf("PublishChat() error = %v, want %v", err, ErrClientNotFound)
	}
}

func TestClaimRequestRejectsDuplicateAfterReconnect(t *testing.T) {
	hub := NewHub(fixedClock{now: time.Date(2026, time.August, 14, 12, 0, 0, 0, time.UTC)})
	members := []Member{{ID: 21, Name: "player", SeatNumber: 1}}
	first, err := hub.Register("room-dedupe", 21, "player", members)
	if err != nil {
		t.Fatal(err)
	}
	if err := hub.ClaimRequest(first, "request-1"); err != nil {
		t.Fatalf("first ClaimRequest() error = %v", err)
	}
	hub.Unregister(first)

	second, err := hub.Register("room-dedupe", 21, "player", members)
	if err != nil {
		t.Fatal(err)
	}
	defer hub.Unregister(second)
	if err := hub.ClaimRequest(second, "request-1"); !errors.Is(err, ErrDuplicateEvent) {
		t.Fatalf("duplicate ClaimRequest() error = %v, want %v", err, ErrDuplicateEvent)
	}
}

func TestBroadcastDropsDuplicateServerEventID(t *testing.T) {
	hub := NewHub(fixedClock{now: time.Date(2026, time.August, 14, 12, 0, 0, 0, time.UTC)})
	client, err := hub.Register("room-events", 31, "player", []Member{{ID: 31, Name: "player", SeatNumber: 1}})
	if err != nil {
		t.Fatal(err)
	}
	defer hub.Unregister(client)
	drainEvents(client.Events())

	event := hub.newEvent(EventRoomUpdated, "request-2", RoomUpdated{Version: 2, Reason: "test"})
	hub.broadcastEvent("room-events", event)
	hub.broadcastEvent("room-events", event)

	select {
	case received := <-client.Events():
		if received.EventID != event.EventID {
			t.Fatalf("event id = %q, want %q", received.EventID, event.EventID)
		}
	case <-time.After(time.Second):
		t.Fatal("expected first event")
	}
	select {
	case duplicate := <-client.Events():
		t.Fatalf("unexpected duplicate event: %#v", duplicate)
	default:
	}
}

func TestPublishChatPersistsBeforeBroadcast(t *testing.T) {
	store := &recordingChatStore{}
	hub := NewHub(fixedClock{now: time.Date(2026, time.August, 14, 12, 0, 0, 0, time.UTC)}, WithChatStore(store))
	client, err := hub.Register("room-chat", 41, "player", []Member{{ID: 41, Name: "player", SeatNumber: 1}})
	if err != nil {
		t.Fatal(err)
	}
	defer hub.Unregister(client)
	drainEvents(client.Events())
	if err := hub.PublishChat(client, "durable message"); err != nil {
		t.Fatal(err)
	}
	if len(store.messages) != 1 || store.messages[0].Text != "durable message" {
		t.Fatalf("stored messages = %#v", store.messages)
	}
}

func TestFullOutboundQueueStopsSlowConsumer(t *testing.T) {
	hub := NewHub(fixedClock{now: time.Now()})
	client, err := hub.RegisterUser(52, "slow")
	if err != nil {
		t.Fatal(err)
	}
	defer hub.Unregister(client)
	for index := 0; index <= outboundQueueSize; index++ {
		hub.Send(client, EventSocialUpdated, "", map[string]int{"index": index})
	}
	select {
	case <-client.Done():
		if !errors.Is(client.CloseReason(), ErrSlowConsumer) {
			t.Fatalf("close reason = %v, want %v", client.CloseReason(), ErrSlowConsumer)
		}
	case <-time.After(time.Second):
		t.Fatal("slow consumer was not stopped")
	}
}

func TestPublishUserDeliversSocialEvent(t *testing.T) {
	hub := NewHub(fixedClock{now: time.Now()})
	client, err := hub.RegisterUser(61, "friend")
	if err != nil {
		t.Fatal(err)
	}
	defer hub.Unregister(client)
	hub.PublishUser(61, EventSocialUpdated, map[string]string{"reason": "friend_request_received"})
	select {
	case event := <-client.Events():
		if event.Type != EventSocialUpdated {
			t.Fatalf("event type = %q, want %q", event.Type, EventSocialUpdated)
		}
	case <-time.After(time.Second):
		t.Fatal("social event was not delivered")
	}
}

func drainEvents(events <-chan Event) {
	for {
		select {
		case <-events:
		default:
			return
		}
	}
}
