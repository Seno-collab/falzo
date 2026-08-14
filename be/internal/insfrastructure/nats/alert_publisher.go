package nats

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"be/internal/alerting"
	"be/internal/observability"

	natsgo "github.com/nats-io/nats.go"
)

const alertQueueSize = 256

var ErrAlertQueueFull = errors.New("NATS alert queue is full")
var ErrAlertPublisherClosed = errors.New("NATS alert publisher is closed")

type AlertPublisher struct {
	nc        *natsgo.Conn
	js        natsgo.JetStreamContext
	subject   string
	queue     chan alerting.Event
	done      chan struct{}
	metrics   *observability.Metrics
	closeOnce sync.Once
	queueMu   sync.RWMutex
	closed    bool
}

func NewAlertPublisher(
	url string,
	stream string,
	subject string,
	metrics *observability.Metrics,
) (*AlertPublisher, error) {
	nc, err := natsgo.Connect(
		url,
		natsgo.Name("falzo-api-error-publisher"),
		natsgo.Timeout(5*time.Second),
		natsgo.MaxReconnects(-1),
		natsgo.ReconnectWait(time.Second),
		natsgo.DrainTimeout(3*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("connect NATS: %w", err)
	}
	js, err := nc.JetStream(natsgo.PublishAsyncMaxPending(alertQueueSize))
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("create JetStream context: %w", err)
	}
	if err := ensureAlertStream(js, stream, subject); err != nil {
		nc.Close()
		return nil, err
	}

	publisher := &AlertPublisher{
		nc: nc, js: js, subject: subject,
		queue: make(chan alerting.Event, alertQueueSize),
		done:  make(chan struct{}), metrics: metrics,
	}
	go publisher.run()
	return publisher, nil
}

func (p *AlertPublisher) Notify(_ context.Context, event alerting.Event) error {
	p.queueMu.RLock()
	defer p.queueMu.RUnlock()
	if p.closed {
		return ErrAlertPublisherClosed
	}
	select {
	case p.queue <- event:
		p.observe("enqueue", "success")
		p.setQueueDepth()
		return nil
	default:
		p.observe("enqueue", "dropped")
		return ErrAlertQueueFull
	}
}

func (p *AlertPublisher) Close(ctx context.Context) error {
	p.closeOnce.Do(func() {
		p.queueMu.Lock()
		p.closed = true
		close(p.queue)
		p.queueMu.Unlock()
	})
	select {
	case <-p.done:
	case <-ctx.Done():
		p.nc.Close()
		return ctx.Err()
	}
	if err := p.nc.Drain(); err != nil {
		p.nc.Close()
		return fmt.Errorf("drain NATS: %w", err)
	}
	return nil
}

func (p *AlertPublisher) run() {
	defer close(p.done)
	for event := range p.queue {
		p.setQueueDepth()
		payload, err := json.Marshal(event)
		if err != nil {
			p.observe("marshal", "error")
			continue
		}
		if err := p.publishWithRetry(payload); err != nil {
			p.observe("publish", "error")
			continue
		}
		p.observe("publish", "success")
	}
	p.setQueueDepth()
}

func (p *AlertPublisher) publishWithRetry(payload []byte) error {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_, lastErr = p.js.Publish(p.subject, payload, natsgo.Context(ctx))
		cancel()
		if lastErr == nil {
			return nil
		}
		time.Sleep(time.Duration(attempt+1) * 250 * time.Millisecond)
	}
	return lastErr
}

func (p *AlertPublisher) observe(stage, result string) {
	if p.metrics != nil {
		p.metrics.AlertNotifications.WithLabelValues(stage, result).Inc()
	}
}

func (p *AlertPublisher) setQueueDepth() {
	if p.metrics != nil {
		p.metrics.AlertQueueDepth.Set(float64(len(p.queue)))
	}
}

func ensureAlertStream(js natsgo.JetStreamContext, stream, subject string) error {
	info, err := js.StreamInfo(stream)
	if err == nil {
		for _, configuredSubject := range info.Config.Subjects {
			if configuredSubject == subject {
				return nil
			}
		}
		config := info.Config
		config.Subjects = append(config.Subjects, subject)
		if _, err := js.UpdateStream(&config); err != nil {
			return fmt.Errorf("add alert subject to NATS stream: %w", err)
		}
		return nil
	}
	if !errors.Is(err, natsgo.ErrStreamNotFound) {
		return fmt.Errorf("inspect NATS alert stream: %w", err)
	}
	_, err = js.AddStream(&natsgo.StreamConfig{
		Name:      stream,
		Subjects:  []string{subject},
		Retention: natsgo.LimitsPolicy,
		Storage:   natsgo.FileStorage,
		MaxAge:    7 * 24 * time.Hour,
		MaxMsgs:   100_000,
		Discard:   natsgo.DiscardOld,
		Replicas:  1,
	})
	if err != nil {
		// API and bot can start together. If the other process won the stream
		// creation race, accept the stream after verifying its subject.
		if info, lookupErr := js.StreamInfo(stream); lookupErr == nil {
			for _, configuredSubject := range info.Config.Subjects {
				if configuredSubject == subject {
					return nil
				}
			}
		}
		return fmt.Errorf("create NATS alert stream: %w", err)
	}
	return nil
}
