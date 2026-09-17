package simulation

import (
	"math"
	"sort"
)

type encounter struct {
	first, second eater
	time          float64
	firstWins     bool
}

func (w *World) resolvePredation(eaters []eater) {
	contacts := make([]encounter, 0)
	for i, first := range eaters {
		for _, second := range eaters[i+1:] {
			if first.player == nil && second.player == nil || first.stats.HP == second.stats.HP {
				continue
			}
			contact, hit := eatingContact(first.start, first.end, first.size, second.start, second.end, second.size)
			if hit {
				contacts = append(contacts, encounter{first: first, second: second, time: contact, firstWins: first.stats.HP > second.stats.HP})
			}
		}
	}
	sort.Slice(contacts, func(i, j int) bool {
		a, b := contacts[i], contacts[j]
		if a.time != b.time {
			return a.time < b.time
		}
		if a.first.id != b.first.id {
			return a.first.id < b.first.id
		}
		return a.second.id < b.second.id
	})
	for _, hit := range contacts {
		if hit.first.dead() || hit.second.dead() {
			continue
		}
		predator, prey := hit.first, hit.second
		if !hit.firstWins {
			predator, prey = prey, predator
		}
		predator.stats.GainXP(float64(prey.stats.Level))
		prey.die(predator.id, w.tick)
	}
}

func (w *World) spawnPosition(slot uint64) (float64, float64) {
	bestX, bestY, bestDistance := 300.0, 300.0, -1.0
	for offset := uint64(0); offset < 400; offset++ {
		cell := (slot + offset) % 400
		x, y := 300+float64(cell%20)*60, 300+float64(cell/20)*60
		distance := math.Inf(1)
		for _, ai := range w.ais {
			if ai.respawnTick == 0 {
				distance = min(distance, math.Hypot(x-ai.X, y-ai.Y))
			}
		}
		for _, player := range w.players {
			if !player.Dead {
				distance = min(distance, math.Hypot(x-player.X, y-player.Y))
			}
		}
		if distance > bestDistance {
			bestX, bestY, bestDistance = x, y, distance
		}
		if distance > AIDetectionRadius+PlayerSize {
			return x, y
		}
	}
	return bestX, bestY
}

func (w *World) nearestPlayer(ai *aiState, threat bool) (*playerState, float64) {
	radius := AIDetectionRadius
	if threat && ai.Behavior == "fleeing" {
		radius *= 1.2
	}
	best := radius * radius
	var selected *playerState
	for _, player := range w.players {
		if player.Dead || threat && player.Stats.HP <= ai.Stats.HP || !threat && player.Stats.HP >= ai.Stats.HP {
			continue
		}
		dx, dy := player.X-ai.X, player.Y-ai.Y
		distance := dx*dx + dy*dy
		if distance <= best && (selected == nil || distance < best || player.ID < selected.ID) {
			selected, best = player, distance
		}
	}
	return selected, best
}

func (ai *aiState) approach(x, y float64) {
	dx, dy := x-ai.X, y-ai.Y
	distance := math.Hypot(dx, dy)
	if distance > 0 {
		ai.X, ai.Y = Move(ai.X, ai.Y, Direction{X: dx / distance, Y: dy / distance}, min(StepSeconds, distance/(ai.Stats.Speed*SpeedScale)), ai.Stats.Speed)
	}
	ai.wanderTicks = 0
}

func (ai *aiState) flee(player *playerState) {
	dx, dy := ai.X-player.X, ai.Y-player.Y
	length := math.Hypot(dx, dy)
	d := Direction{X: 1}
	if length > 0 {
		d = Direction{X: dx / length, Y: dy / length}
	}
	if ai.escapeTicks > 0 {
		d = ai.escapeDirection
		ai.escapeTicks--
	}
	blocked := (ai.X <= MapMin && d.X < 0) || (ai.X >= MaxPosition && d.X > 0) || (ai.Y <= MapMin && d.Y < 0) || (ai.Y >= MaxPosition && d.Y > 0)
	if blocked {
		best := -1.0
		for _, candidate := range [8]Direction{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}, {X: 1, Y: 1}, {X: 1, Y: -1}, {X: -1, Y: 1}, {X: -1, Y: -1}} {
			x, y := Move(ai.X, ai.Y, candidate, StepSeconds, ai.Stats.Speed)
			if x == ai.X && y == ai.Y {
				continue
			}
			score := (x-player.X)*(x-player.X) + (y-player.Y)*(y-player.Y)
			if score > best {
				best = score
				d = candidate
			}
		}
		ai.escapeDirection, ai.escapeTicks = d, 9
	}
	ai.X, ai.Y = Move(ai.X, ai.Y, d, StepSeconds, ai.Stats.Speed)
	ai.wanderTicks = 0
}
