package roomapp

import (
	roomports "be/internal/application/ports/room"
	domainroom "be/internal/domain/room"
	"be/internal/observability"
	"be/internal/shared/clock"
	"context"
	"log/slog"
	"time"
)

const (
	phaseSchedulerInterval = 500 * time.Millisecond
	phaseSchedulerBatch    = 100
)

type PhaseTransitionHandler func(context.Context, domainroom.PhaseTransition)

// PhaseScheduler is the authoritative clock for deadline-driven game changes.
// Repository row locks make RunOnce safe across multiple backend replicas.
type PhaseScheduler struct {
	repository   roomports.Repository
	clock        clock.Clock
	metrics      *observability.Metrics
	logger       *slog.Logger
	onTransition PhaseTransitionHandler
}

func NewPhaseScheduler(
	repository roomports.Repository,
	c clock.Clock,
	metrics *observability.Metrics,
	logger *slog.Logger,
	onTransition PhaseTransitionHandler,
) *PhaseScheduler {
	if logger == nil {
		logger = slog.Default()
	}
	return &PhaseScheduler{
		repository:   repository,
		clock:        c,
		metrics:      metrics,
		logger:       logger,
		onTransition: onTransition,
	}
}

func (s *PhaseScheduler) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(phaseSchedulerInterval)
		defer ticker.Stop()
		s.runOnce(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.runOnce(ctx)
			}
		}
	}()
}

func (s *PhaseScheduler) runOnce(ctx context.Context) {
	now := s.clock.Now()
	count, err := s.repository.CountExpiredRounds(ctx, now)
	if err != nil {
		s.logger.ErrorContext(ctx, "count overdue game rounds", slog.Any("error", err))
		if s.metrics != nil {
			s.metrics.GamePhaseTransitions.WithLabelValues("unknown", "unknown", "error").Inc()
		}
		return
	}
	if s.metrics != nil {
		s.metrics.GameOverdueRounds.Set(float64(count))
	}
	if count == 0 {
		return
	}

	transitions, err := s.repository.AdvanceExpiredRounds(ctx, now, phaseSchedulerBatch)
	if err != nil {
		s.logger.ErrorContext(ctx, "advance expired game rounds", slog.Any("error", err))
		if s.metrics != nil {
			s.metrics.GamePhaseTransitions.WithLabelValues("unknown", "unknown", "error").Inc()
		}
		return
	}
	for _, transition := range transitions {
		if s.metrics != nil {
			s.metrics.GamePhaseTransitions.WithLabelValues(
				string(transition.From), string(transition.To), "success",
			).Inc()
			lag := transition.TransitionedAt.Sub(transition.PreviousDeadlineAt).Seconds()
			if lag < 0 {
				lag = 0
			}
			s.metrics.GamePhaseDeadlineLag.Observe(lag)
		}
		if s.onTransition != nil {
			s.onTransition(ctx, transition)
		}
	}
}
