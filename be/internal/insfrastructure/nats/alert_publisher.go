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

const (
	alertConnectTimeout   = 2 * time.Second
	alertOperationTimeout = 3 * time.Second
	alertPublishAttempts  = 3
)

var ErrAlertQueueFull = errors.New("NATS alert queue is full")
var ErrAlertPublisherClosed = errors.New("NATS alert publisher is closed")

type AlertPublisher struct {
	url       string
	stream    string
	nc        *natsgo.Conn
	js        natsgo.JetStreamContext
	subject   string
	queue     chan alerting.Event
	done      chan struct{}
	cancel    context.CancelFunc
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
) *AlertPublisher {
	ctx, cancel := context.WithCancel(context.Background())
	publisher := &AlertPublisher{
		url: url, stream: stream, subject: subject,
		queue: make(chan alerting.Event, alertQueueSize),
		done:  make(chan struct{}), cancel: cancel, metrics: metrics,
	}
	// Connecting and publishing happen only on this worker. NATS latency or
	// downtime must never delay API startup or an application request.
	go publisher.run(ctx)
	return publisher
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
		p.cancel()
		p.queueMu.Unlock()
	})
	select {
	case <-p.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *AlertPublisher) run(ctx context.Context) {
	defer close(p.done)
	defer p.closeConnection()
	for {
		if ctx.Err() != nil {
			p.setQueueDepth()
			return
		}
		select {
		case <-ctx.Done():
			p.setQueueDepth()
			return
		case event := <-p.queue:
			p.setQueueDepth()
			payload, err := json.Marshal(event)
			if err != nil {
				p.observe("marshal", "error")
				continue
			}
			if err := p.publishWithRetry(ctx, payload); err != nil {
				p.observe("publish", "error")
				continue
			}
			p.observe("publish", "success")
		}
	}
}

func (p *AlertPublisher) publishWithRetry(ctx context.Context, payload []byte) error {
	var lastErr error
	for attempt := 0; attempt < alertPublishAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := p.ensureConnection(ctx); err != nil {
			lastErr = err
			p.observe("connect", "error")
		} else {
			publishCtx, cancel := context.WithTimeout(ctx, alertOperationTimeout)
			_, lastErr = p.js.Publish(p.subject, payload, natsgo.Context(publishCtx))
			cancel()
		}
		if lastErr == nil {
			return nil
		}
		if attempt+1 < alertPublishAttempts {
			timer := time.NewTimer(time.Duration(attempt+1) * 250 * time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}
	}
	return lastErr
}

func (p *AlertPublisher) ensureConnection(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if p.nc != nil && !p.nc.IsClosed() {
		return nil
	}
	p.closeConnection()
	nc, err := natsgo.Connect(
		p.url,
		natsgo.Name("falzo-api-error-publisher"),
		natsgo.Timeout(alertConnectTimeout),
		natsgo.MaxReconnects(-1),
		natsgo.ReconnectWait(time.Second),
		natsgo.DrainTimeout(alertOperationTimeout),
	)
	if err != nil {
		return fmt.Errorf("connect NATS: %w", err)
	}
	if err := ctx.Err(); err != nil {
		nc.Close()
		return err
	}
	js, err := nc.JetStream(
		natsgo.MaxWait(alertOperationTimeout),
		natsgo.PublishAsyncMaxPending(alertQueueSize),
	)
	if err != nil {
		nc.Close()
		return fmt.Errorf("create JetStream context: %w", err)
	}
	if err := ensureAlertStream(js, p.stream, p.subject); err != nil {
		nc.Close()
		return err
	}
	p.nc = nc
	p.js = js
	p.observe("connect", "success")
	return nil
}

func (p *AlertPublisher) closeConnection() {
	if p.nc != nil {
		p.nc.Close()
		p.nc = nil
		p.js = nil
	}
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
