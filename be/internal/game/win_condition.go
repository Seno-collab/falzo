package game

import domainroom "be/internal/domain/room"

type AliveRoles struct {
	Civilians  int
	Undercover int
	MrWhite    int
}

func EvaluateWinner(alive AliveRoles, mrWhiteGuessCorrect bool) *domainroom.WinningSide {
	if mrWhiteGuessCorrect {
		winner := domainroom.WinningSideMrWhite
		return &winner
	}
	if alive.Undercover == 0 && alive.MrWhite == 0 {
		winner := domainroom.WinningSideCivilians
		return &winner
	}
	if alive.Undercover > 0 && alive.Undercover >= alive.Civilians {
		winner := domainroom.WinningSideUndercover
		return &winner
	}
	return nil
}
