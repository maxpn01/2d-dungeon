package simulation

import (
	"fmt"
	"math"
	"sort"
)

type position struct{ x, y float64 }

type eater struct {
	cell       *SplitCell
	id         string
	start, end position
	size       float64
	stats      *Stats
	player     *playerState
	ai         *aiState
}

func (w *World) eaters(previous map[string]position, previousAIs []position) []eater {
	eaters := make([]eater, 0, len(w.players)+len(w.ais))
	for id, player := range w.players {
		if player.Dead {
			continue
		}
		eaters = append(eaters, eater{id: id, start: previous[id], end: position{player.X, player.Y}, size: PlayerSize, stats: &player.Stats, player: player})
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
	sort.Slice(eaters, func(i, j int) bool { return eaters[i].id < eaters[j].id })
	for i := range w.ais {
		ai := &w.ais[i]
		if ai.respawnTick != 0 {
			continue
		}
		eaters = append(eaters, eater{id: ai.ID, start: previousAIs[i], end: position{ai.X, ai.Y}, size: AISize, stats: &ai.Stats, ai: ai})
	}
	return eaters
}

func (e eater) dead() bool {
	if e.player != nil {
		return e.player.Dead || e.stats.HP <= 0
	}
	return e.ai.respawnTick != 0
}

func (e eater) die(killer string, tick uint64) {
	if e.player != nil && e.player.Split.Count() > 0 {
		e.stats.HP = 0
		if e.player.Stats.HP > 0 || e.player.Split.Mass() > 0 {
			return
		}
	}
	if e.player == nil {
		e.ai.respawnTick = tick + AIRespawnSeconds*TickRate
		return
	}
	e.player.Dead = true
	e.player.KilledBy = killer
	e.player.direction = Direction{}
	e.player.commands = nil
	e.player.InputSeq = e.player.lastSequence
}

func (w *World) consumeFruits(eaters []eater) {
	remaining := w.fruits[:0]
	for _, fruit := range w.fruits {
		if fruit.ExpiresTick != 0 && w.tick >= fruit.ExpiresTick {
			continue
		}
		if w.tick < fruit.AvailableTick {
			remaining = append(remaining, fruit)
			continue
		}
		winner := -1
		earliest := math.Inf(1)
		for i, actor := range eaters {
			if fruit.Ejected && actor.player != nil && actor.player.ID == fruit.Source && w.tick < fruit.OwnerPickupTick {
				continue
			}
			if actor.dead() || actor.ai != nil && (actor.ai.Behavior == "sleeping" || actor.ai.Behavior == "emerging") {
				continue
			}
			location := position{fruit.X, fruit.Y}
			contact, hit := eatingContact(actor.start, actor.end, actor.size, location, location, fruit.Size)
			if hit && (contact < earliest || contact == earliest && (winner < 0 || actor.id < eaters[winner].id)) {
				winner, earliest = i, contact
			}
		}
		if winner < 0 {
			remaining = append(remaining, fruit)
		} else {
			hp := FruitHP
			if fruit.Source != "" {
				hp = fruit.HP
			}
			xp := FruitXP * hp
			if fruit.Ejected {
				xp = 0
			}
			eaters[winner].gain(hp, xp)
		}
	}
	w.fruits = remaining
}

func eatingContact(start, end position, size float64, targetStart, targetEnd position, targetSize float64) (float64, bool) {
	enter, exit := 0.0, 1.0
	for _, axis := range [][4]float64{{start.x, end.x - start.x - (targetEnd.x - targetStart.x), targetStart.x - size, targetStart.x + targetSize}, {start.y, end.y - start.y - (targetEnd.y - targetStart.y), targetStart.y - size, targetStart.y + targetSize}} {
		start, delta, lo, hi := axis[0], axis[1], axis[2], axis[3]
		if delta == 0 {
			if start <= lo || start >= hi {
				return 0, false
			}
			continue
		}
		a, b := (lo-start)/delta, (hi-start)/delta
		if a > b {
			a, b = b, a
		}
		enter = max(enter, a)
		exit = min(exit, b)
		if enter >= exit {
			return 0, false
		}
	}
	return enter, enter < 1 && exit > 0
}

func (e eater) gain(hp, xp float64) {
	if e.player == nil || e.player.Split.Count() == 0 {
		e.stats.Eat(hp, xp)
		return
	}
	p := e.player
	p.Stats.GainXP(xp)
	for i := range p.Split {
		if p.Split[i].Active {
			p.Split[i].Stats.Level, p.Split[i].Stats.XP = p.Stats.Level, p.Stats.XP
		}
	}
	room := max(0, MaxHPForLevel(p.Stats.Level)-p.Stats.HP-p.Split.Mass())
	e.stats.HP += min(hp, room)
}
