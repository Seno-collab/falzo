package lib

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type ShutdownManager struct {
	mu      sync.Mutex
	phases  []shutDownPhase
	timeout time.Duration
}

type shutDownPhase struct {
	name    string
	timeout time.Duration
	fn      func(ctx context.Context) error
}

func NewShutdownManager(total time.Duration) *ShutdownManager {
	return &ShutdownManager{timeout: total}
}

func (s *ShutdownManager) Register(name string, timeout time.Duration, fn func(ctx context.Context) error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.phases = append(s.phases, shutDownPhase{name, timeout, fn})
}

func (s *ShutdownManager) Shutdown() error {
	s.mu.Lock()
	phases := s.phases
	s.mu.Unlock()
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
