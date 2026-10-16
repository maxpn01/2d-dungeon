package simulation

import (
	"fmt"
)

func (e eater) takeMP() float64 {
	owner := e.ownerStats()
	totalHP := owner.HP
	if e.player != nil {
		totalHP += e.player.Split.Mass()
	} else {
		totalHP += e.ai.Split.Mass()
	}
	amount := owner.MP * min(1, e.stats.HP/max(e.stats.HP, totalHP))
	owner.MP = max(0, owner.MP-amount)
	return amount * ActorMPFraction
}

func (w *World) splitEaters(previous map[string]position) []eater {
	var eaters []eater
	for id, player := range w.players {
		if player.Dead {
			continue
		}
		for i := range player.Split {
			cell := &player.Split[i]
			if !cell.Active {
				continue
			}
			cellID := fmt.Sprintf("%s/split/%d", id, i)
			start, ok := previous[cellID]
			if !ok {
				start = position{cell.X, cell.Y}
			}
			eaters = append(eaters, eater{id: cellID, start: start, end: position{cell.X, cell.Y}, size: PlayerSize, stats: &cell.Stats, player: player, cell: cell})
		}
	}
	for i := range w.ais {
		ai := &w.ais[i]
		if ai.respawnTick != 0 {
			continue
		}
		for j := range ai.Split {
			cell := &ai.Split[j]
			if !cell.Active {
				continue
			}
			id := fmt.Sprintf("%s/split/%d", ai.ID, j)
			start, ok := previous[id]
			if !ok {
				start = position{cell.X, cell.Y}
			}
			eaters = append(eaters, eater{id: id, start: start, end: position{cell.X, cell.Y}, size: AISize, stats: &cell.Stats, ai: ai, cell: cell})
		}
	}
	return eaters
}

func (e eater) splitSurvives() bool {
	if e.player != nil && e.player.Split.Count() > 0 {
		e.stats.HP = 0
		if e.player.Stats.HP > 0 || e.player.Split.Mass() > 0 {
			return true
		}
	}
	if e.player == nil {
		if e.ai.Split.Count() > 0 {
			e.stats.HP = 0
			if e.ai.Stats.HP > 0 || e.ai.Split.Mass() > 0 {
				return true
			}
		}
	}
	return false
}

func (e eater) gainSplit(hp, xp float64) {
	var stats *Stats
	var split *SplitCells
	if e.player != nil {
		stats, split = &e.player.Stats, &e.player.Split
	} else {
		stats, split = &e.ai.Stats, &e.ai.Split
	}
	stats.GainXP(xp)
	for i := range *split {
		if (*split)[i].Active {
			(*split)[i].Stats.Level, (*split)[i].Stats.XP = stats.Level, stats.XP
			(*split)[i].Stats.rebuildPerks()
		}
	}
	room := max(0, stats.MaxHP()-stats.HP-split.Mass())
	e.stats.HP += min(hp, room)
}
