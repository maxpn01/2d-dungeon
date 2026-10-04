package simulation

import "sort"

func (w *World) Finished() bool { return w.winnerID != "" }

func (w *World) RestartRound() {
	players := make([]*playerState, 0, len(w.players))
	for _, p := range w.players {
		players = append(players, p)
	}
	sort.Slice(players, func(i, j int) bool { return players[i].spawnSlot < players[j].spawnSlot })
	next := NewWorld()
	for _, p := range players {
		next.Join(p.ID)
		next.SetCharacter(p.ID, p.Character)
		next.players[p.ID].lastSequence = p.lastSequence
		next.players[p.ID].InputSeq = p.lastSequence
	}
	*w = *next
}
