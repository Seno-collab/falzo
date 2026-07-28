package realtime

import (
	"be/internal/shared/clock"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	Subprotocol = "falzo.v1"

	EventPresenceSnapshot = "presence.snapshot"
	EventChatMessage      = "chat.message"
	EventRoundStarted     = "game.round.started"
	EventError            = "error"

	MaxChatMessageRunes = 500
	outboundQueueSize   = 64
	presenceTTL         = 45 * time.Second
	presenceSweepEvery  = 15 * time.Second
	backplaneTimeout    = 2 * time.Second
	chatRateLimitWindow = 5 * time.Second
	chatRateLimitCount  = 10
)

var (
	ErrHubClosed      = errors.New("realtime hub is closed")
	ErrClientNotFound = errors.New("realtime client is not registered")
	ErrEmptyMessage   = errors.New("chat message is required")
	ErrMessageTooLong = errors.New("chat message is too long")
	ErrRateLimited    = errors.New("chat message rate limit exceeded")
)

type Event struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

type Member struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	SeatNumber int    `json:"seat_number"`
	Host       bool   `json:"host"`
}

type PresencePlayer struct {
	Member
	Online bool `json:"online"`
}

type PresenceSnapshot struct {
	Players []PresencePlayer `json:"players"`
}

type ChatMessage struct {
	ID       string    `json:"id"`
	UserID   int64     `json:"user_id"`
	UserName string    `json:"username"`
	Text     string    `json:"text"`
	SentAt   time.Time `json:"sent_at"`
}

type RoundStarted struct {
	Round   int       `json:"round"`
	DealtAt time.Time `json:"dealt_at"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Client struct {
	id       string
	roomID   string
	userID   int64
	userName string
	events   chan Event
	done     chan struct{}
	windowAt time.Time
	windowN  int
}

func (c *Client) Events() <-chan Event  { return c.events }
func (c *Client) Done() <-chan struct{} { return c.done }
func (c *Client) RoomID() string        { return c.roomID }
func (c *Client) UserID() int64         { return c.userID }
func (c *Client) UserName() string      { return c.userName }

type roomState struct {
	members     map[int64]Member
	clients     map[string]*Client
	connections map[int64]int
}

type Hub struct {
	mu        sync.Mutex
	clock     clock.Clock
	backplane Backplane
	logger    *slog.Logger
	rooms     map[string]*roomState
	closed    bool
	started   bool
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

type HubOption func(*Hub)

func WithBackplane(backplane Backplane) HubOption {
	return func(h *Hub) { h.backplane = backplane }
}

func WithLogger(logger *slog.Logger) HubOption {
	return func(h *Hub) { h.logger = logger }
}

func NewHub(c clock.Clock, options ...HubOption) *Hub {
	hub := &Hub{clock: c, logger: slog.Default(), rooms: make(map[string]*roomState)}
	for _, option := range options {
		option(hub)
	}
	if hub.logger == nil {
		hub.logger = slog.Default()
	}
	return hub
}

func (h *Hub) Start(parent context.Context) error {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return ErrHubClosed
	}
	if h.started || h.backplane == nil {
		h.started = true
		h.mu.Unlock()
		return nil
	}
	ctx, cancel := context.WithCancel(parent)
	h.cancel = cancel
	h.started = true
	h.mu.Unlock()

	messages, err := h.backplane.Subscribe(ctx)
	if err != nil {
		cancel()
		h.mu.Lock()
		h.started = false
		h.cancel = nil
		h.mu.Unlock()
		return err
	}
	h.wg.Add(2)
	go h.consumeBackplane(ctx, messages)
	go h.sweepPresence(ctx)
	return nil
}

func (h *Hub) Register(roomID string, userID int64, userName string, members []Member) (*Client, error) {
	h.mu.Lock()

	if h.closed {
		h.mu.Unlock()
		return nil, ErrHubClosed
	}

	room := h.rooms[roomID]
	if room == nil {
		room = &roomState{
			members:     make(map[int64]Member),
			clients:     make(map[string]*Client),
			connections: make(map[int64]int),
		}
		h.rooms[roomID] = room
	}
	replaceMembers(room, members)

	client := &Client{
		id:       uuid.NewString(),
		roomID:   roomID,
		userID:   userID,
		userName: userName,
		events:   make(chan Event, outboundQueueSize),
		done:     make(chan struct{}),
	}
	room.clients[client.id] = client
	room.connections[userID]++
	if h.backplane == nil {
		h.broadcastPresenceLocked(room)
	}
	membersSnapshot := membersForRoom(room)
	h.mu.Unlock()

	if h.backplane != nil {
		if err := h.touchPresence(client); err != nil {
			h.logger.Warn("redis presence touch failed", slog.Any("error", err))
			h.broadcastLocalPresence(roomID)
		} else if err := h.publishPresenceSync(roomID, membersSnapshot); err != nil {
			h.logger.Warn("redis presence sync publish failed", slog.Any("error", err))
			h.refreshPresence(roomID)
		}
	}
	return client, nil
}

func (h *Hub) Unregister(client *Client) {
	if client == nil {
		return
	}

	h.mu.Lock()

	room := h.rooms[client.roomID]
	if room == nil || room.clients[client.id] != client {
		h.mu.Unlock()
		return
	}

	delete(room.clients, client.id)
	room.connections[client.userID]--
	if room.connections[client.userID] <= 0 {
		delete(room.connections, client.userID)
	}
	close(client.done)

	if len(room.clients) == 0 {
		delete(h.rooms, client.roomID)
	} else if h.backplane == nil {
		h.broadcastPresenceLocked(room)
	}
	h.mu.Unlock()

	if h.backplane != nil {
		ctx, cancel := h.backplaneContext()
		err := h.backplane.RemovePresence(ctx, client.roomID, client.id, client.userID)
		cancel()
		if err != nil {
			h.logger.Warn("redis presence remove failed", slog.Any("error", err))
		}
		if err := h.publishPresenceSync(client.roomID, nil); err != nil {
			h.logger.Warn("redis presence sync publish failed", slog.Any("error", err))
			h.broadcastLocalPresence(client.roomID)
		}
	}
}

func (h *Hub) UpdateMembers(roomID string, members []Member) {
	h.mu.Lock()

	room := h.rooms[roomID]
	if room == nil {
		h.mu.Unlock()
		return
	}
	replaceMembers(room, members)
	if h.backplane == nil {
		h.broadcastPresenceLocked(room)
	}
	h.mu.Unlock()
	if h.backplane != nil {
		if err := h.publishPresenceSync(roomID, members); err != nil {
			h.logger.Warn("redis presence sync publish failed", slog.Any("error", err))
			h.broadcastLocalPresence(roomID)
		}
	}
}

func (h *Hub) PublishChat(client *Client, text string) error {
	text = strings.TrimSpace(text)
	switch {
	case text == "":
		return ErrEmptyMessage
	case utf8.RuneCountInString(text) > MaxChatMessageRunes:
		return ErrMessageTooLong
	}

	h.mu.Lock()

	room := h.rooms[client.roomID]
	if room == nil || room.clients[client.id] != client {
		h.mu.Unlock()
		return ErrClientNotFound
	}
	now := h.clock.Now()
	if client.windowAt.IsZero() || now.Sub(client.windowAt) >= chatRateLimitWindow {
		client.windowAt = now
		client.windowN = 0
	}
	if client.windowN >= chatRateLimitCount {
		h.mu.Unlock()
		return ErrRateLimited
	}
	client.windowN++
	message := ChatMessage{
		ID:       uuid.NewString(),
		UserID:   client.userID,
		UserName: client.userName,
		Text:     text,
		SentAt:   now.UTC(),
	}
	if h.backplane == nil {
		h.broadcastLocked(room, Event{Type: EventChatMessage, Payload: message})
		h.mu.Unlock()
		return nil
	}
	h.mu.Unlock()
	if err := h.publish(client.roomID, EventChatMessage, message); err != nil {
		h.logger.Warn("redis chat publish failed", slog.Any("error", err))
		h.broadcastEvent(client.roomID, Event{Type: EventChatMessage, Payload: message})
	}
	return nil
}

func (h *Hub) Publish(roomID, eventType string, payload any) {
	if h.backplane != nil {
		if err := h.publish(roomID, eventType, payload); err == nil {
			return
		} else {
			h.logger.Warn("redis realtime publish failed", slog.Any("error", err))
		}
	}
	h.broadcastEvent(roomID, Event{Type: eventType, Payload: payload})
}

func (h *Hub) Heartbeat(client *Client) {
	if h.backplane == nil || client == nil {
		return
	}
	if err := h.touchPresence(client); err != nil {
		h.logger.Warn("redis presence heartbeat failed", slog.Any("error", err))
	}
}

func (h *Hub) SendError(client *Client, code, message string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room := h.rooms[client.roomID]
	if room == nil || room.clients[client.id] != client {
		return
	}
	enqueue(client, Event{Type: EventError, Payload: ErrorPayload{Code: code, Message: message}})
}

func (h *Hub) Close() {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return
	}
	h.closed = true
	if h.cancel != nil {
		h.cancel()
	}
	clients := make([]*Client, 0)
	roomIDs := make(map[string]struct{})
	for _, room := range h.rooms {
		for _, client := range room.clients {
			close(client.done)
			clients = append(clients, client)
			roomIDs[client.roomID] = struct{}{}
		}
	}
	clear(h.rooms)
	h.mu.Unlock()
	if h.backplane != nil {
		for _, client := range clients {
			ctx, cancel := h.backplaneContext()
			_ = h.backplane.RemovePresence(ctx, client.roomID, client.id, client.userID)
			cancel()
		}
		for roomID := range roomIDs {
			_ = h.publishPresenceSync(roomID, nil)
		}
		_ = h.backplane.Close()
	}
	h.wg.Wait()
}

func (h *Hub) consumeBackplane(ctx context.Context, messages <-chan BackplaneMessage) {
	defer h.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case message, ok := <-messages:
			if !ok {
				return
			}
			h.handleBackplaneMessage(message)
		}
	}
}

func (h *Hub) handleBackplaneMessage(message BackplaneMessage) {
	if message.Type == EventPresenceSync {
		var payload PresenceSync
		if err := json.Unmarshal(message.Payload, &payload); err != nil {
			h.logger.Warn("invalid presence sync payload", slog.Any("error", err))
			return
		}
		h.mu.Lock()
		if room := h.rooms[message.RoomID]; room != nil && len(payload.Members) > 0 {
			replaceMembers(room, payload.Members)
		}
		h.mu.Unlock()
		h.refreshPresence(message.RoomID)
		return
	}
	h.broadcastEvent(message.RoomID, Event{Type: message.Type, Payload: message.Payload})
}

func (h *Hub) sweepPresence(ctx context.Context) {
	defer h.wg.Done()
	ticker := time.NewTicker(presenceSweepEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.mu.Lock()
			roomIDs := make([]string, 0, len(h.rooms))
			for roomID := range h.rooms {
				roomIDs = append(roomIDs, roomID)
			}
			h.mu.Unlock()
			for _, roomID := range roomIDs {
				h.refreshPresence(roomID)
			}
		}
	}
}

func (h *Hub) touchPresence(client *Client) error {
	ctx, cancel := h.backplaneContext()
	defer cancel()
	return h.backplane.TouchPresence(
		ctx,
		client.roomID,
		client.id,
		client.userID,
		h.clock.Now().Add(presenceTTL),
	)
}

func (h *Hub) publish(roomID, eventType string, payload any) error {
	ctx, cancel := h.backplaneContext()
	defer cancel()
	return h.backplane.Publish(ctx, roomID, eventType, payload)
}

func (h *Hub) publishPresenceSync(roomID string, members []Member) error {
	return h.publish(roomID, EventPresenceSync, PresenceSync{Members: members})
}

func (h *Hub) refreshPresence(roomID string) {
	h.mu.Lock()
	_, exists := h.rooms[roomID]
	h.mu.Unlock()
	if !exists {
		return
	}

	ctx, cancel := h.backplaneContext()
	onlineUsers, err := h.backplane.OnlineUsers(ctx, roomID, h.clock.Now())
	cancel()
	if err != nil {
		h.logger.Warn("redis presence read failed", slog.Any("error", err))
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	if room := h.rooms[roomID]; room != nil {
		h.broadcastPresenceWithOnlineLocked(room, onlineUsers)
	}
}

func (h *Hub) broadcastLocalPresence(roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if room := h.rooms[roomID]; room != nil {
		h.broadcastPresenceLocked(room)
	}
}

func (h *Hub) broadcastEvent(roomID string, event Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if room := h.rooms[roomID]; room != nil {
		h.broadcastLocked(room, event)
	}
}

func (h *Hub) backplaneContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), backplaneTimeout)
}

func replaceMembers(room *roomState, members []Member) {
	room.members = make(map[int64]Member, len(members))
	for _, member := range members {
		room.members[member.ID] = member
	}
}

func (h *Hub) broadcastPresenceLocked(room *roomState) {
	onlineUsers := make(map[int64]bool, len(room.connections))
	for userID := range room.connections {
		onlineUsers[userID] = true
	}
	h.broadcastPresenceWithOnlineLocked(room, onlineUsers)
}

func (h *Hub) broadcastPresenceWithOnlineLocked(room *roomState, onlineUsers map[int64]bool) {
	players := make([]PresencePlayer, 0, len(room.members))
	for _, member := range room.members {
		players = append(players, PresencePlayer{
			Member: member,
			Online: onlineUsers[member.ID],
		})
	}
	sort.Slice(players, func(i, j int) bool {
		return players[i].SeatNumber < players[j].SeatNumber
	})
	h.broadcastLocked(room, Event{
		Type:    EventPresenceSnapshot,
		Payload: PresenceSnapshot{Players: players},
	})
}

func membersForRoom(room *roomState) []Member {
	members := make([]Member, 0, len(room.members))
	for _, member := range room.members {
		members = append(members, member)
	}
	sort.Slice(members, func(i, j int) bool {
		return members[i].SeatNumber < members[j].SeatNumber
	})
	return members
}

func (h *Hub) broadcastLocked(room *roomState, event Event) {
	for _, client := range room.clients {
		enqueue(client, event)
	}
}

func enqueue(client *Client, event Event) {
	select {
	case client.events <- event:
	default:
		// A bounded queue keeps a slow connection from blocking the whole room.
	}
}
