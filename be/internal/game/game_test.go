package game

import (
	domainroom "be/internal/domain/room"
	"testing"
)

func TestAssignClassicRolesIncludesDistinctMrWhite(t *testing.T) {
	picks := []int{1, 1}
	index := 0
	assignments, err := AssignClassicRoles([]int64{10, 20, 30, 40}, true, func(int) (int, error) {
		value := picks[index]
		index++
		return value, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	roles := map[int64]domainroom.CardRole{}
	for _, assignment := range assignments {
		roles[assignment.PlayerID] = assignment.Role
	}
	if roles[20] != domainroom.CardRoleUndercover {
		t.Fatalf("player 20 role = %s, want undercover", roles[20])
	}
	if roles[30] != domainroom.CardRoleMrWhite {
		t.Fatalf("player 30 role = %s, want mr_white", roles[30])
	}
}

func TestAssignClassicRolesWithoutMrWhite(t *testing.T) {
	assignments, err := AssignClassicRoles([]int64{1, 2, 3, 4}, false, func(int) (int, error) {
		return 0, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	undercoverCount := 0
	mrWhiteCount := 0
	for _, assignment := range assignments {
		if assignment.Role == domainroom.CardRoleUndercover {
			undercoverCount++
		}
		if assignment.Role == domainroom.CardRoleMrWhite {
			mrWhiteCount++
		}
	}
	if undercoverCount != 1 || mrWhiteCount != 0 {
		t.Fatalf("roles: undercover=%d, mrWhite=%d", undercoverCount, mrWhiteCount)
	}
}

func TestAssignClassicRolesRejectsFewerThanFourPlayers(t *testing.T) {
	if _, err := AssignClassicRoles([]int64{1, 2, 3}, true, nil); err == nil {
		t.Fatal("expected fewer than four players to be rejected")
	}
}

func TestEvaluateWinner(t *testing.T) {
	tests := []struct {
		name  string
		alive AliveRoles
		guess bool
		want  *domainroom.WinningSide
	}{
		{name: "civilians", alive: AliveRoles{Civilians: 2}, want: winner(domainroom.WinningSideCivilians)},
		{name: "undercover parity", alive: AliveRoles{Civilians: 1, Undercover: 1}, want: winner(domainroom.WinningSideUndercover)},
		{name: "mr white guess", alive: AliveRoles{Civilians: 3, Undercover: 1}, guess: true, want: winner(domainroom.WinningSideMrWhite)},
		{name: "continue", alive: AliveRoles{Civilians: 3, Undercover: 1}, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EvaluateWinner(tt.alive, tt.guess)
			if (got == nil) != (tt.want == nil) || got != nil && *got != *tt.want {
				t.Fatalf("winner = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStateTransitions(t *testing.T) {
	valid := [][2]domainroom.RoundPhase{
		{domainroom.RoundPhaseWaiting, domainroom.RoundPhaseRevealingRole},
		{domainroom.RoundPhaseRevealingRole, domainroom.RoundPhaseDescribing},
		{domainroom.RoundPhaseDescribing, domainroom.RoundPhaseVoting},
		{domainroom.RoundPhaseVoting, domainroom.RoundPhaseRevealingResult},
		{domainroom.RoundPhaseRevealingResult, domainroom.RoundPhaseMrWhiteGuessing},
		{domainroom.RoundPhaseMrWhiteGuessing, domainroom.RoundPhaseGameFinished},
		{domainroom.RoundPhaseGameFinished, domainroom.RoundPhaseRevealingRole},
	}
	for _, transition := range valid {
		if !domainroom.CanTransition(transition[0], transition[1]) {
			t.Fatalf("transition %s -> %s must be valid", transition[0], transition[1])
		}
	}
	if domainroom.CanTransition(domainroom.RoundPhaseVoting, domainroom.RoundPhaseDescribing) {
		t.Fatal("voting must reveal its result before another description cycle")
	}
}

func winner(value domainroom.WinningSide) *domainroom.WinningSide { return &value }
