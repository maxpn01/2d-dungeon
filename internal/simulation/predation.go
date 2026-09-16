package simulation

import (
	"math"
	"sort"
)

type encounter struct {
	player     *playerState
	ai         *aiState
	time       float64
	playerWins bool
}

func (w *World) resolvePredation(previous map[string]position, previousAIs []position) {
	contacts := make([]encounter, 0)
	for id, player := range w.players {
		if player.Dead {
			continue
		}
		p := previous[id]
		for i := range w.ais {
			ai := &w.ais[i]
			if ai.respawnTick != 0 || player.Stats.HP == ai.Stats.HP {
				continue
			}
			a := previousAIs[i]
			contact, hit := fruitContact(p.x, p.y, player.X-(ai.X-a.x), player.Y-(ai.Y-a.y), PlayerSize, Fruit{X: a.x, Y: a.y, Size: AISize})
			if hit {
				contacts = append(contacts, encounter{player: player, ai: ai, time: contact, playerWins: player.Stats.HP > ai.Stats.HP})
			}
		}
	}
	sort.Slice(contacts, func(i, j int) bool {
		a, b := contacts[i], contacts[j]
		if a.time != b.time {
			return a.time < b.time
		}
		if a.player.ID != b.player.ID {
			return a.player.ID < b.player.ID
		}
		return a.ai.ID < b.ai.ID
	})
	for _, hit := range contacts {
		player, ai := hit.player, hit.ai
		if player.Dead || ai.respawnTick != 0 {
			continue
		}
		if hit.playerWins {
			player.Stats.GainXP(float64(ai.Stats.Level))
			ai.respawnTick = w.tick + 3*TickRate
		} else {
			ai.Stats.GainXP(float64(player.Stats.Level))
			player.Dead = true
			player.KilledBy = ai.ID
			player.direction = Direction{}
			player.commands = nil
			player.InputSeq = player.lastSequence
		}
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
		ai.X, ai.Y = Move(ai.X, ai.Y, Direction{X: dx / distance, Y: dy / distance}, min(StepSeconds, distance/ai.Stats.Speed), ai.Stats.Speed)
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
