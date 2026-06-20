package shutdown

import (
	"context"
	"errors"
	"falzo-be/pkg/logger"
	"fmt"
	"sync"
	"time"
)

var shutdownLog = logger.For("shutdown")

type Manager struct {
	mu     sync.Mutex
	phases []phase
}

type phase struct {
	name    string
	timeout time.Duration
	fn      func(ctx context.Context) error
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Register(name string, timeout time.Duration, fn func(ctx context.Context) error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.phases = append(m.phases, phase{
		name:    name,
		timeout: timeout,
		fn:      fn,
	})
}

func (m *Manager) Shutdown() error {
	m.mu.Lock()
	phases := append([]phase(nil), m.phases...)
	m.mu.Unlock()

	var errs []error
	for _, phase := range phases {
		ctx, cancel := context.WithTimeout(context.Background(), phase.timeout)
		shutdownLog.Info(ctx, "shutdown phase starting", logger.Str("phase", phase.name), logger.Dur("timeout", phase.timeout))

		if err := phase.fn(ctx); err != nil {
			shutdownLog.Error(ctx, err, "shutdown phase failed", logger.Str("phase", phase.name))
			errs = append(errs, fmt.Errorf("phase %s: %w", phase.name, err))
		} else {
			shutdownLog.Info(ctx, "shutdown phase complete", logger.Str("phase", phase.name))
		}
		cancel()
	}

	return errors.Join(errs...)
}
