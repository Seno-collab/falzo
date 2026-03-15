package lib

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
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
		fmt.Printf("shutdown phase starting, phase. %s", phase.name)
		if err := phase.fn(ctx); err != nil {
			fmt.Printf("shutdown phase failed, phase %s, error %v", phase.name, err)
			errs = append(errs, fmt.Errorf("phase %s: %w", phase.name, err))
		} else {
			fmt.Printf("shutdown phase complete, phase %s", phase.name)
		}
		cancel()
	}
	return errors.Join(errs...)
}
