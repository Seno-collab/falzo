package realtime

import (
	domainchat "be/internal/domain/chat"
	domainroom "be/internal/domain/room"
	"be/internal/observability"
	"be/internal/shared/clock"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

	EventPresenceSnapshot   = "presence.snapshot"
	EventChatMessage        = "chat.message"
	EventRoomUpdated        = "room.updated"
	EventRoundStarted       = "game.round.started"
	EventVoteUpdated        = "game.vote.updated"
	EventStateUpdated       = "game.state.updated"
	EventSocialUpdated      = "social.notifications.updated"
	EventError              = "error"
	eventConnectionReplaced = "socket.connection.replaced"
	eventRoomMemberEvicted  = "room.member.evicted"

	MaxChatMessageRunes = 500
	outboundQueueSize   = 64
	presenceTTL         = 45 * time.Second
	presenceSweepEvery  = 15 * time.Second
	backplaneTimeout    = 2 * time.Second
	chatRateLimitWindow = 5 * time.Second
	chatRateLimitCount  = 10
	requestDedupeTTL    = 2 * time.Minute
	eventDedupeTTL      = 2 * time.Minute
	connectionLeaseTTL  = 45 * time.Second
	maxRequestIDLength  = 128
	userChannelPrefix   = "user:"
)

var (
	ErrHubClosed          = errors.New("realtime hub is closed")
	ErrClientNotFound     = errors.New("realtime client is not registered")
	ErrEmptyMessage       = errors.New("chat message is required")
	ErrMessageTooLong     = errors.New("chat message is too long")
	ErrRateLimited        = errors.New("chat message rate limit exceeded")
	ErrSpectator          = errors.New("eliminated players cannot send chat messages")
	ErrConnectionReplaced = errors.New("connection was replaced by a newer connection")
	ErrRoomMemberRemoved  = errors.New("room membership was removed")
	ErrRequestIDRequired  = errors.New("request id is required")
	ErrDuplicateEvent     = errors.New("duplicate socket event")
	ErrSlowConsumer       = errors.New("socket client outbound queue is full")
	ErrChatUnavailable    = errors.New("chat persistence is unavailable")
)

type Event struct {
	EventID    string    `json:"event_id"`
	Type       string    `json:"type"`
	RequestID  string    `json:"request_id,omitempty"`
	OccurredAt time.Time `json:"occurred_at"`
	Payload    any       `json:"payload"`
}

type Member struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	SeatNumber int    `json:"seat_number"`
	Host       bool   `json:"host"`
	Eliminated bool   `json:"eliminated"`
}

type PresencePlayer struct {
	Member
	Online bool `json:"online"`
}

type PresenceSnapshot struct {
	Players []PresencePlayer `json:"players"`
}

type ChatMessage = domainchat.Message

type ChatStore interface {
	SaveMessage(ctx context.Context, message domainchat.Message) error
}

type RoundStarted struct {
	Round           int                   `json:"round"`
	DealtAt         time.Time             `json:"dealt_at"`
	Phase           domainroom.RoundPhase `json:"phase"`
	PhaseDeadlineAt time.Time             `json:"phase_deadline_at"`
}

type RoomUpdated struct {
	Version int64  `json:"version"`
	Reason  string `json:"reason"`
}

type VoteUpdated struct {
	Round          int  `json:"round"`
	VotesCast      int  `json:"votes_cast"`
	EligibleVoters int  `json:"eligible_voters"`
	Completed      bool `json:"completed"`
}

type StateUpdated struct {
	Round             int                   `json:"round"`
	Cycle             int                   `json:"cycle"`
	Phase             domainroom.RoundPhase `json:"phase"`
	CurrentTurnUserID *int64                `json:"current_turn_player_id,omitempty"`
	PhaseDeadlineAt   *time.Time            `json:"phase_deadline_at,omitempty"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type connectionReplaced struct {
	UserID               int64  `json:"user_id"`
	PreviousConnectionID string `json:"previous_connection_id"`
	NewConnectionID      string `json:"new_connection_id"`
}

type roomMemberEvicted struct {
	UserID int64 `json:"user_id"`
}

type Client struct {
	id            string
	roomID        string
	userID        int64
	userName      string
	events        chan Event
	done          chan struct{}
	windowAt      time.Time
	windowN       int
	stopOnce      sync.Once
	stopErr       error
	scope         string
	trackPresence bool
}

func (c *Client) Events() <-chan Event  { return c.events }
func (c *Client) Done() <-chan struct{} { return c.done }
func (c *Client) RoomID() string        { return c.roomID }
func (c *Client) UserID() int64         { return c.userID }
func (c *Client) UserName() string      { return c.userName }
func (c *Client) Scope() string         { return c.scope }
func (c *Client) CloseReason() error {
	if c.stopErr != nil {
		return c.stopErr
	}
	return ErrClientNotFound
}

func (c *Client) stop(err error) {
	c.stopOnce.Do(func() {
		c.stopErr = err
		close(c.done)
	})
}

type roomState struct {
	members       map[int64]Member
	clients       map[string]*Client
	activeByUser  map[int64]*Client
	connections   map[int64]int
	seenEvents    map[string]time.Time
	trackPresence bool
}

type Hub struct {
	mu             sync.Mutex
	clock          clock.Clock
	backplane      Backplane
	logger         *slog.Logger
	metrics        *observability.Metrics
	chatStore      ChatStore
	rooms          map[string]*roomState
	closed         bool
	started        bool
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	recentRequests map[string]time.Time
}

type HubOption func(*Hub)

func WithBackplane(backplane Backplane) HubOption {
	return func(h *Hub) { h.backplane = backplane }
}

func WithLogger(logger *slog.Logger) HubOption {
	return func(h *Hub) { h.logger = logger }
}

func WithMetrics(metrics *observability.Metrics) HubOption {
	return func(h *Hub) { h.metrics = metrics }
}

func WithChatStore(chatStore ChatStore) HubOption {
	return func(h *Hub) { h.chatStore = chatStore }
}

func NewHub(c clock.Clock, options ...HubOption) *Hub {
	hub := &Hub{
		clock:          c,
		logger:         slog.Default(),
		rooms:          make(map[string]*roomState),
		recentRequests: make(map[string]time.Time),
	}
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
			members:       make(map[int64]Member),
			clients:       make(map[string]*Client),
			activeByUser:  make(map[int64]*Client),
			connections:   make(map[int64]int),
			seenEvents:    make(map[string]time.Time),
			trackPresence: true,
		}
		h.rooms[roomID] = room
	}
	replaceMembers(room, members)

	client := &Client{
		id:            uuid.NewString(),
		roomID:        roomID,
		userID:        userID,
		userName:      userName,
		events:        make(chan Event, outboundQueueSize),
		done:          make(chan struct{}),
		scope:         "room",
		trackPresence: true,
	}
	replaced := room.activeByUser[userID]
	if replaced != nil {
		delete(room.clients, replaced.id)
		delete(room.connections, replaced.userID)
		replaced.stop(ErrConnectionReplaced)
		if h.metrics != nil {
			h.metrics.RealtimeConnections.Dec()
		}
	}
	room.clients[client.id] = client
	room.activeByUser[userID] = client
	room.connections[userID] = 1
	if h.backplane == nil {
		h.broadcastPresenceLocked(room)
	}
	membersSnapshot := membersForRoom(room)
	h.mu.Unlock()
	if h.metrics != nil {
		h.metrics.RealtimeConnections.Inc()
		h.metrics.RealtimeConnectionsTotal.WithLabelValues("room", "accepted").Inc()
	}

	if h.backplane != nil {
		if replaced != nil {
			h.removePresence(replaced)
		}
		previousID, err := h.claimConnection(client)
		if err != nil {
			h.Unregister(client)
			return nil, fmt.Errorf("claim active socket connection: %w", err)
		} else if previousID != "" && previousID != client.id {
			h.publishConnectionReplaced(roomID, userID, previousID, client.id)
		}
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

func (h *Hub) RegisterUser(userID int64, userName string) (*Client, error) {
	channelID := userChannelID(userID)
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return nil, ErrHubClosed
	}
	room := h.rooms[channelID]
	if room == nil {
		room = &roomState{
			members: make(map[int64]Member), clients: make(map[string]*Client),
			activeByUser: make(map[int64]*Client), connections: make(map[int64]int),
			seenEvents: make(map[string]time.Time), trackPresence: false,
		}
		h.rooms[channelID] = room
	}
	client := &Client{
		id: uuid.NewString(), roomID: channelID, userID: userID, userName: userName,
		events: make(chan Event, outboundQueueSize), done: make(chan struct{}),
		scope: "user", trackPresence: false,
	}
	replaced := room.activeByUser[userID]
	if replaced != nil {
		delete(room.clients, replaced.id)
		delete(room.connections, replaced.userID)
		replaced.stop(ErrConnectionReplaced)
		if h.metrics != nil {
			h.metrics.RealtimeConnections.Dec()
		}
	}
	room.clients[client.id] = client
	room.activeByUser[userID] = client
	room.connections[userID] = 1
	h.mu.Unlock()
	if h.metrics != nil {
		h.metrics.RealtimeConnections.Inc()
		h.metrics.RealtimeConnectionsTotal.WithLabelValues("user", "accepted").Inc()
	}
	if h.backplane != nil {
		previousID, err := h.claimConnection(client)
		if err != nil {
			h.Unregister(client)
			return nil, fmt.Errorf("claim active user socket connection: %w", err)
		}
		if previousID != "" && previousID != client.id {
			h.publishConnectionReplaced(channelID, userID, previousID, client.id)
		}
	}
	return client, nil
}

func (h *Hub) PublishUser(userID int64, eventType string, payload any) {
	h.Publish(userChannelID(userID), eventType, payload)
}

func (h *Hub) Unregister(client *Client) {
	if client == nil {
		return
	}

	h.mu.Lock()

	room := h.rooms[client.roomID]
	registered := room != nil && room.clients[client.id] == client
	if !registered {
		h.mu.Unlock()
		if h.backplane != nil {
			if client.trackPresence {
				h.removePresence(client)
			}
			h.releaseConnection(client)
		}
		return
	}

	delete(room.clients, client.id)
	if room.activeByUser[client.userID] == client {
		delete(room.activeByUser, client.userID)
	}
	delete(room.connections, client.userID)
	client.stop(ErrClientNotFound)
	if h.metrics != nil {
		h.metrics.RealtimeConnections.Dec()
	}

	if len(room.clients) == 0 {
		delete(h.rooms, client.roomID)
	} else if h.backplane == nil {
		h.broadcastPresenceLocked(room)
	}
	h.mu.Unlock()

	if h.backplane != nil {
		if client.trackPresence {
			h.removePresence(client)
		}
		h.releaseConnection(client)
		if client.trackPresence {
			if err := h.publishPresenceSync(client.roomID, nil); err != nil {
				h.logger.Warn("redis presence sync publish failed", slog.Any("error", err))
				h.broadcastLocalPresence(client.roomID)
			}
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

func (h *Hub) EvictRoomMember(roomID string, userID int64) {
	event := h.newEvent(eventRoomMemberEvicted, "", roomMemberEvicted{UserID: userID})
	if h.backplane != nil {
		if err := h.publish(roomID, event); err == nil {
			return
		} else {
			h.logger.Warn("redis room member eviction publish failed", slog.Any("error", err))
		}
	}
	h.evictLocalRoomMember(roomID, userID)
}

func (h *Hub) PublishChat(client *Client, text string) error {
	return h.PublishChatForRequest(client, "", text)
}

func (h *Hub) PublishChatForRequest(client *Client, requestID, text string) error {
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
	if member, exists := room.members[client.userID]; !exists || member.Eliminated {
		h.mu.Unlock()
		return ErrSpectator
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
		RoomID:   client.roomID,
		UserID:   client.userID,
		UserName: client.userName,
		Text:     text,
		SentAt:   now.UTC(),
	}
	h.mu.Unlock()
	if h.chatStore != nil {
		ctx, cancel := context.WithTimeout(context.Background(), backplaneTimeout)
		err := h.chatStore.SaveMessage(ctx, message)
		cancel()
		if err != nil {
			h.logger.Error("persist chat message", slog.Any("error", err))
			return ErrChatUnavailable
		}
	}
	h.mu.Lock()
	room = h.rooms[client.roomID]
	if room == nil || room.clients[client.id] != client {
		h.mu.Unlock()
		return ErrClientNotFound
	}
	if h.backplane == nil {
		h.broadcastLocked(room, h.newEvent(EventChatMessage, requestID, message))
		h.mu.Unlock()
		return nil
	}
	h.mu.Unlock()
	event := h.newEvent(EventChatMessage, requestID, message)
	if err := h.publish(client.roomID, event); err != nil {
		h.logger.Warn("redis chat publish failed", slog.Any("error", err))
		h.broadcastEvent(client.roomID, event)
	}
	return nil
}

func (h *Hub) Publish(roomID, eventType string, payload any) {
	h.PublishForRequest(roomID, eventType, "", payload)
}

func (h *Hub) PublishForRequest(roomID, eventType, requestID string, payload any) {
	event := h.newEvent(eventType, requestID, payload)
	if h.backplane != nil {
		if err := h.publish(roomID, event); err == nil {
			return
		} else {
			h.logger.Warn("redis realtime publish failed", slog.Any("error", err))
		}
	}
	h.broadcastEvent(roomID, event)
}

func (h *Hub) Heartbeat(client *Client) error {
	if h.backplane == nil || client == nil {
		return nil
	}
	owner, err := h.refreshConnection(client)
	if err != nil {
		return err
	}
	if !owner {
		client.stop(ErrConnectionReplaced)
		return ErrConnectionReplaced
	}
	if client.trackPresence {
		if err := h.touchPresence(client); err != nil {
			h.logger.Warn("redis presence heartbeat failed", slog.Any("error", err))
		}
	}
	return nil
}

func (h *Hub) SendError(client *Client, code, message string) {
	h.SendRequestError(client, "", code, message)
}

func (h *Hub) SendRequestError(client *Client, requestID, code, message string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room := h.rooms[client.roomID]
	if room == nil || room.clients[client.id] != client {
		return
	}
	h.enqueueLocked(client, h.newEvent(EventError, requestID, ErrorPayload{Code: code, Message: message}))
}

func (h *Hub) Send(client *Client, eventType, requestID string, payload any) {
	h.mu.Lock()
	defer h.mu.Unlock()
	room := h.rooms[client.roomID]
	if room == nil || room.clients[client.id] != client {
		return
	}
	h.enqueueLocked(client, h.newEvent(eventType, requestID, payload))
}

func (h *Hub) ClaimRequest(client *Client, requestID string) error {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" || len(requestID) > maxRequestIDLength {
		return ErrRequestIDRequired
	}
	now := h.clock.Now()
	key := client.roomID + ":" + fmt.Sprintf("%d", client.userID) + ":" + requestID

	h.mu.Lock()
	room := h.rooms[client.roomID]
	if room == nil || room.clients[client.id] != client {
		h.mu.Unlock()
		return ErrClientNotFound
	}
	pruneDedupeMap(h.recentRequests, now)
	if expiresAt, exists := h.recentRequests[key]; exists && expiresAt.After(now) {
		h.mu.Unlock()
		return ErrDuplicateEvent
	}
	h.mu.Unlock()

	if h.backplane != nil {
		owner, err := h.refreshConnection(client)
		if err != nil {
			return err
		}
		if !owner {
			client.stop(ErrConnectionReplaced)
			return ErrConnectionReplaced
		}
		ctx, cancel := h.backplaneContext()
		claimed, err := h.backplane.ClaimRequest(ctx, client.roomID, client.userID, requestID, now.Add(requestDedupeTTL))
		cancel()
		if err != nil {
			return err
		}
		if !claimed {
			return ErrDuplicateEvent
		}
	}

	h.mu.Lock()
	h.recentRequests[key] = now.Add(requestDedupeTTL)
	h.mu.Unlock()
	return nil
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
	for _, room := range h.rooms {
		for _, client := range room.clients {
			client.stop(ErrHubClosed)
			clients = append(clients, client)
			if h.metrics != nil {
				h.metrics.RealtimeConnections.Dec()
			}
		}
	}
	clear(h.rooms)
	h.mu.Unlock()
	if h.backplane != nil {
		for _, client := range clients {
			if client.trackPresence {
				h.removePresence(client)
			}
			h.releaseConnection(client)
		}
		for _, client := range clients {
			if client.trackPresence {
				_ = h.publishPresenceSync(client.roomID, nil)
			}
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
	if message.Type == eventRoomMemberEvicted {
		var payload roomMemberEvicted
		if err := json.Unmarshal(message.Payload, &payload); err != nil {
			h.logger.Warn("invalid room member eviction payload", slog.Any("error", err))
			return
		}
		h.evictLocalRoomMember(message.RoomID, payload.UserID)
		return
	}
	if message.Type == eventConnectionReplaced {
		var payload connectionReplaced
		if err := json.Unmarshal(message.Payload, &payload); err != nil {
			h.logger.Warn("invalid connection replacement payload", slog.Any("error", err))
			return
		}
		h.mu.Lock()
		var replaced *Client
		if room := h.rooms[message.RoomID]; room != nil {
			candidate := room.activeByUser[payload.UserID]
			if candidate != nil && candidate.id == payload.PreviousConnectionID && candidate.id != payload.NewConnectionID {
				replaced = candidate
				candidate.stop(ErrConnectionReplaced)
			}
		}
		h.mu.Unlock()
		if replaced != nil {
			h.Unregister(replaced)
		}
		return
	}
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
	h.broadcastEvent(message.RoomID, Event{
		EventID:    message.EventID,
		Type:       message.Type,
		RequestID:  message.RequestID,
		OccurredAt: message.OccurredAt,
		Payload:    message.Payload,
	})
}

func (h *Hub) evictLocalRoomMember(roomID string, userID int64) {
	h.mu.Lock()
	room := h.rooms[roomID]
	if room == nil {
		h.mu.Unlock()
		return
	}
	delete(room.members, userID)
	client := room.activeByUser[userID]
	if client != nil {
		client.stop(ErrRoomMemberRemoved)
	}
	h.mu.Unlock()

	if client != nil {
		h.Unregister(client)
	}
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
			for roomID, room := range h.rooms {
				if room.trackPresence {
					roomIDs = append(roomIDs, roomID)
				}
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

func (h *Hub) claimConnection(client *Client) (string, error) {
	ctx, cancel := h.backplaneContext()
	defer cancel()
	return h.backplane.ClaimConnection(
		ctx, client.roomID, client.userID, client.id, h.clock.Now().Add(connectionLeaseTTL),
	)
}

func (h *Hub) refreshConnection(client *Client) (bool, error) {
	ctx, cancel := h.backplaneContext()
	defer cancel()
	return h.backplane.RefreshConnection(
		ctx, client.roomID, client.userID, client.id, h.clock.Now().Add(connectionLeaseTTL),
	)
}

func (h *Hub) releaseConnection(client *Client) {
	ctx, cancel := h.backplaneContext()
	err := h.backplane.ReleaseConnection(ctx, client.roomID, client.userID, client.id)
	cancel()
	if err != nil {
		h.logger.Warn("redis connection release failed", slog.Any("error", err))
	}
}

func (h *Hub) removePresence(client *Client) {
	ctx, cancel := h.backplaneContext()
	err := h.backplane.RemovePresence(ctx, client.roomID, client.id, client.userID)
	cancel()
	if err != nil {
		h.logger.Warn("redis presence remove failed", slog.Any("error", err))
	}
}

func (h *Hub) publishConnectionReplaced(roomID string, userID int64, previousID, newID string) {
	event := h.newEvent(eventConnectionReplaced, "", connectionReplaced{
		UserID:               userID,
		PreviousConnectionID: previousID,
		NewConnectionID:      newID,
	})
	if err := h.publish(roomID, event); err != nil {
		h.logger.Warn("redis connection replacement publish failed", slog.Any("error", err))
	}
}

func (h *Hub) publish(roomID string, event Event) error {
	ctx, cancel := h.backplaneContext()
	defer cancel()
	return h.backplane.Publish(ctx, roomID, event)
}

func (h *Hub) publishPresenceSync(roomID string, members []Member) error {
	return h.publish(roomID, h.newEvent(EventPresenceSync, "", PresenceSync{Members: members}))
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

func userChannelID(userID int64) string {
	return userChannelPrefix + fmt.Sprintf("%d", userID)
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
	h.broadcastLocked(room, h.newEvent(EventPresenceSnapshot, "", PresenceSnapshot{Players: players}))
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
	if event.EventID != "" {
		now := h.clock.Now()
		pruneDedupeMap(room.seenEvents, now)
		if expiresAt, exists := room.seenEvents[event.EventID]; exists && expiresAt.After(now) {
			return
		}
		room.seenEvents[event.EventID] = now.Add(eventDedupeTTL)
	}
	for _, client := range room.clients {
		h.enqueueLocked(client, event)
	}
}

func (h *Hub) newEvent(eventType, requestID string, payload any) Event {
	return Event{
		EventID:    uuid.NewString(),
		Type:       eventType,
		RequestID:  requestID,
		OccurredAt: h.clock.Now().UTC(),
		Payload:    payload,
	}
}

func pruneDedupeMap(entries map[string]time.Time, now time.Time) {
	for key, expiresAt := range entries {
		if !expiresAt.After(now) {
			delete(entries, key)
		}
	}
}

func (h *Hub) enqueueLocked(client *Client, event Event) {
	select {
	case client.events <- event:
	default:
		if h.metrics != nil {
			h.metrics.RealtimeQueueDropped.WithLabelValues(event.Type).Inc()
		}
		// A client that missed an ordered state event must reconnect and sync a
		// fresh authoritative snapshot rather than continue with stale state.
		client.stop(ErrSlowConsumer)
	}
}
