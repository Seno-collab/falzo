package roomapp

import (
	roomports "be/internal/application/ports/room"
	domainroom "be/internal/domain/room"
	"context"
	"testing"
	"time"
)

type schedulerTestClock struct{ now time.Time }

func (c schedulerTestClock) Now() time.Time { return c.now }

type schedulerTestRepository struct {
	roomports.Repository
	overdue      int
	transitions  []domainroom.PhaseTransition
	advanceCalls int
}

func (r *schedulerTestRepository) CountExpiredRounds(context.Context, time.Time) (int, error) {
	return r.overdue, nil
}

func (r *schedulerTestRepository) AdvanceExpiredRounds(context.Context, time.Time, int) ([]domainroom.PhaseTransition, error) {
	r.advanceCalls++
	return r.transitions, nil
}

func TestPhaseSchedulerPublishesCommittedTransitions(t *testing.T) {
	now := time.Date(2026, time.August, 14, 12, 0, 0, 0, time.UTC)
	repository := &schedulerTestRepository{
		overdue: 1,
		transitions: []domainroom.PhaseTransition{{
			RoomID: "room-1", RoundNumber: 2, From: domainroom.RoundPhaseVoting,
			To: domainroom.RoundPhaseRevealingResult, TransitionedAt: now,
		}},
	}
	received := make([]domainroom.PhaseTransition, 0, 1)
	scheduler := NewPhaseScheduler(repository, schedulerTestClock{now: now}, nil, nil,
		func(_ context.Context, transition domainroom.PhaseTransition) {
			received = append(received, transition)
		})

	scheduler.runOnce(context.Background())

	if repository.advanceCalls != 1 {
		t.Fatalf("AdvanceExpiredRounds calls = %d, want 1", repository.advanceCalls)
	}
	if len(received) != 1 || received[0].To != domainroom.RoundPhaseRevealingResult {
		t.Fatalf("received transitions = %#v", received)
	}
}

func TestPhaseSchedulerSkipsAdvanceWhenNothingIsOverdue(t *testing.T) {
	repository := &schedulerTestRepository{}
	scheduler := NewPhaseScheduler(repository, schedulerTestClock{now: time.Now()}, nil, nil, nil)
	scheduler.runOnce(context.Background())
	if repository.advanceCalls != 0 {
		t.Fatalf("AdvanceExpiredRounds calls = %d, want 0", repository.advanceCalls)
	}
}
