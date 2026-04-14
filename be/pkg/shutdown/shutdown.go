package shutdown

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

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
		log.Info().Str("phase", phase.name).Dur("timeout", phase.timeout).Msg("shutdown phase starting")

		if err := phase.fn(ctx); err != nil {
			log.Error().Err(err).Str("phase", phase.name).Msg("shutdown phase failed")
			errs = append(errs, fmt.Errorf("phase %s: %w", phase.name, err))
		} else {
			log.Info().Str("phase", phase.name).Msg("shutdown phase complete")
		}
		cancel()
	}

	return errors.Join(errs...)
}
