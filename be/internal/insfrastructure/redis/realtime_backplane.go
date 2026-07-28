package redis

import (
	"be/internal/realtime"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const (
	realtimeSubscriptionBuffer = 256
	presenceKeyTTL             = 2 * time.Minute
)

type RealtimeBackplane struct {
	client *goredis.Client
	prefix string
	logger *slog.Logger
	mu     sync.Mutex
	pubsub *goredis.PubSub
	closed bool
}

func NewRealtimeBackplane(client *goredis.Client, prefix string, logger *slog.Logger) *RealtimeBackplane {
	if logger == nil {
		logger = slog.Default()
	}
	return &RealtimeBackplane{client: client, prefix: prefix, logger: logger}
}

func (b *RealtimeBackplane) Subscribe(ctx context.Context) (<-chan realtime.BackplaneMessage, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil, fmt.Errorf("realtime backplane is closed")
	}
	if b.pubsub != nil {
		return nil, fmt.Errorf("realtime backplane is already subscribed")
	}

	pubsub := b.client.Subscribe(ctx, b.channel())
	if _, err := pubsub.Receive(ctx); err != nil {
		_ = pubsub.Close()
		return nil, fmt.Errorf("subscribe realtime channel: %w", err)
	}
	b.pubsub = pubsub

	messages := make(chan realtime.BackplaneMessage, realtimeSubscriptionBuffer)
	go b.consume(ctx, pubsub.Channel(goredis.WithChannelSize(realtimeSubscriptionBuffer)), messages)
	return messages, nil
}

func (b *RealtimeBackplane) Publish(ctx context.Context, roomID, eventType string, payload any) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal realtime payload: %w", err)
	}
	messageJSON, err := json.Marshal(realtime.BackplaneMessage{
		RoomID:  roomID,
		Type:    eventType,
		Payload: payloadJSON,
	})
	if err != nil {
		return fmt.Errorf("marshal realtime message: %w", err)
	}
	return b.client.Publish(ctx, b.channel(), messageJSON).Err()
}

func (b *RealtimeBackplane) TouchPresence(
	ctx context.Context,
	roomID,
	connectionID string,
	userID int64,
	expiresAt time.Time,
) error {
	_, err := b.client.Pipelined(ctx, func(pipe goredis.Pipeliner) error {
		pipe.ZAdd(ctx, b.presenceKey(roomID), goredis.Z{
			Score:  float64(expiresAt.UnixMilli()),
			Member: presenceMember(connectionID, userID),
		})
		pipe.Expire(ctx, b.presenceKey(roomID), presenceKeyTTL)
		return nil
	})
	return err
}

func (b *RealtimeBackplane) RemovePresence(ctx context.Context, roomID, connectionID string, userID int64) error {
	return b.client.ZRem(ctx, b.presenceKey(roomID), presenceMember(connectionID, userID)).Err()
}

func (b *RealtimeBackplane) OnlineUsers(ctx context.Context, roomID string, now time.Time) (map[int64]bool, error) {
	var active *goredis.StringSliceCmd
	_, err := b.client.Pipelined(ctx, func(pipe goredis.Pipeliner) error {
		pipe.ZRemRangeByScore(ctx, b.presenceKey(roomID), "-inf", strconv.FormatInt(now.UnixMilli(), 10))
		active = pipe.ZRangeByScore(ctx, b.presenceKey(roomID), &goredis.ZRangeBy{
			Min: strconv.FormatInt(now.UnixMilli()+1, 10),
			Max: "+inf",
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	users := make(map[int64]bool)
	for _, member := range active.Val() {
		separator := strings.LastIndexByte(member, ':')
		if separator < 0 {
			continue
		}
		userID, err := strconv.ParseInt(member[separator+1:], 10, 64)
		if err == nil {
			users[userID] = true
		}
	}
	return users, nil
}

func (b *RealtimeBackplane) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return nil
	}
	b.closed = true
	if b.pubsub != nil {
		return b.pubsub.Close()
	}
	return nil
}

func (b *RealtimeBackplane) consume(
	ctx context.Context,
	source <-chan *goredis.Message,
	destination chan<- realtime.BackplaneMessage,
) {
	defer close(destination)
	for {
		select {
		case <-ctx.Done():
			return
		case message, ok := <-source:
			if !ok {
				return
			}
			var decoded realtime.BackplaneMessage
			if err := json.Unmarshal([]byte(message.Payload), &decoded); err != nil {
				b.logger.Warn("invalid redis realtime message", slog.Any("error", err))
				continue
			}
			select {
			case destination <- decoded:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (b *RealtimeBackplane) channel() string {
	return b.prefix + ":realtime:events"
}

func (b *RealtimeBackplane) presenceKey(roomID string) string {
	return b.prefix + ":realtime:presence:" + roomID
}

func presenceMember(connectionID string, userID int64) string {
	return connectionID + ":" + strconv.FormatInt(userID, 10)
}
