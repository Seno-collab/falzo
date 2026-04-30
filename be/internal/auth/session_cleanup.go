package auth

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

const sessionCleanupConfigRefreshInterval = time.Minute

func RunSessionCleanup(ctx context.Context, sessions SessionRepository) {
	if sessions == nil {
		return
	}

	configChanged := make(chan struct{}, 1)
	go watchSessionCleanupConfig(ctx, sessions, configChanged)

	ticker := time.NewTicker(sessionCleanupConfigRefreshInterval)
	defer ticker.Stop()

	var lastRun time.Time
	runSessionCleanupIfDue(ctx, sessions, &lastRun)

	for {
		select {
		case <-ctx.Done():
			return
		case <-configChanged:
			runSessionCleanupIfDue(ctx, sessions, &lastRun)
		case <-ticker.C:
			runSessionCleanupIfDue(ctx, sessions, &lastRun)
		}
	}
}

func watchSessionCleanupConfig(ctx context.Context, sessions SessionRepository, changed chan<- struct{}) {
	for {
		if err := sessions.WaitSessionCleanupConfigChange(ctx); err != nil {
			if ctx.Err() != nil {
				return
			}

			log.Error().Err(err).Msg("auth session cleanup config listener failed")
			select {
			case <-ctx.Done():
				return
			case <-time.After(sessionCleanupConfigRefreshInterval):
				continue
			}
		}

		select {
		case changed <- struct{}{}:
		default:
		}
	}
}

func runSessionCleanupIfDue(ctx context.Context, sessions SessionRepository, lastRun *time.Time) {
	cfg, err := sessions.SessionCleanupConfig(ctx)
	if err != nil {
		log.Error().Err(err).Msg("auth session cleanup config load failed")
		return
	}
	if !cfg.Enabled {
		log.Debug().Msg("auth session cleanup disabled")
		return
	}
	if cfg.Interval <= 0 {
		log.Error().Dur("interval", cfg.Interval).Msg("auth session cleanup interval is invalid")
		return
	}
	if cfg.Retention < 0 {
		log.Error().Dur("retention", cfg.Retention).Msg("auth session cleanup retention is invalid")
		return
	}

	now := time.Now().UTC()
	if !lastRun.IsZero() && now.Sub(*lastRun) < cfg.Interval {
		return
	}

	runSessionCleanup(ctx, sessions, cfg.Retention)
	*lastRun = now
}

func runSessionCleanup(ctx context.Context, sessions SessionRepository, retention time.Duration) {
	deleted, err := sessions.CleanupExpired(ctx, retention)
	if err != nil {
		log.Error().Err(err).Msg("auth session cleanup failed")
		return
	}

	log.Info().Int64("deleted_sessions", deleted).Msg("auth session cleanup completed")
}
