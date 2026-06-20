package auth

import (
	"context"
	"falzo-be/pkg/logger"
	"time"
)

const sessionCleanupConfigRefreshInterval = time.Minute

var sessionCleanupLog = logger.For("auth.session_cleanup")

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

			sessionCleanupLog.Error(ctx, err, "auth session cleanup config listener failed")
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
		sessionCleanupLog.Error(ctx, err, "auth session cleanup config load failed")
		return
	}
	if !cfg.Enabled {
		sessionCleanupLog.Debug(ctx, "auth session cleanup disabled")
		return
	}
	if cfg.Interval <= 0 {
		sessionCleanupLog.Error(ctx, nil, "auth session cleanup interval is invalid", logger.Dur("interval", cfg.Interval))
		return
	}
	if cfg.Retention < 0 {
		sessionCleanupLog.Error(ctx, nil, "auth session cleanup retention is invalid", logger.Dur("retention", cfg.Retention))
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
		sessionCleanupLog.Error(ctx, err, "auth session cleanup failed")
		return
	}

	sessionCleanupLog.Info(ctx, "auth session cleanup completed", logger.Int64("deleted_sessions", deleted))
}
