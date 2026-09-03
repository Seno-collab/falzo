package nats

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"be/internal/alerting"

	natsgo "github.com/nats-io/nats.go"
)

type AlertConsumer struct {
	nc  *natsgo.Conn
	sub *natsgo.Subscription
}

func NewAlertConsumer(url, stream, subject, durable string) (*AlertConsumer, error) {
	nc, err := natsgo.Connect(
		url,
		natsgo.Name("falzo-telegram-error-bot"),
		natsgo.Timeout(5*time.Second),
		natsgo.MaxReconnects(-1),
		natsgo.ReconnectWait(time.Second),
		natsgo.DrainTimeout(3*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("connect NATS: %w", err)
	}
	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("create JetStream context: %w", err)
	}
	if err := ensureAlertStream(js, stream, subject); err != nil {
		nc.Close()
		return nil, err
	}
	sub, err := js.PullSubscribe(
		subject,
		durable,
		natsgo.BindStream(stream),
		natsgo.ManualAck(),
		natsgo.AckExplicit(),
		natsgo.AckWait(30*time.Second),
		natsgo.MaxDeliver(10),
	)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("create durable NATS alert consumer: %w", err)
	}
	return &AlertConsumer{nc: nc, sub: sub}, nil
}

func (c *AlertConsumer) Run(ctx context.Context, handle func(context.Context, alerting.Event) error) error {
	for {
		if err := ctx.Err(); err != nil {
			return nil
		}
		messages, err := c.sub.Fetch(10, natsgo.MaxWait(time.Second))
		if err != nil {
			if errors.Is(err, natsgo.ErrTimeout) {
				continue
			}
			return fmt.Errorf("fetch NATS alerts: %w", err)
		}
		for _, message := range messages {
			var event alerting.Event
			if err := json.Unmarshal(message.Data, &event); err != nil || event.SchemaVersion != alerting.SchemaVersion {
				_ = message.Term()
				continue
			}
			handleCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			err := handle(handleCtx, event)
			cancel()
			if err != nil {
				_ = message.NakWithDelay(15 * time.Second)
				continue
			}
			if err := message.Ack(); err != nil {
				return fmt.Errorf("ack NATS alert: %w", err)
			}
		}
	}
}

func (c *AlertConsumer) Close() error {
	if err := c.nc.Drain(); err != nil {
		c.nc.Close()
		return fmt.Errorf("drain NATS consumer: %w", err)
	}
	return nil
}
