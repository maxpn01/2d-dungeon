package simulation

import (
	"math"
	"raylib-2d-sandbox/internal/physics"
	"raylib-2d-sandbox/internal/worldmap"
)

type sentinelTactic struct {
	leader string
	next   int
	until  uint64
	struck bool
}

func (w *World) sentinelFormation(ai *aiState, target *playerState) (bool, physics.Vector) {
	guards := make([]*aiState, 0, 4)
	slot := 0
	busy := false
	for i := range w.ais {
		guard := &w.ais[i]
		if guard.Character != Triangle || guard.den.RoomID != ai.den.RoomID || guard.respawnTick != 0 || guard.Stats.HP <= 0 || guard.Stone || guard.Awakening > 0 {
			continue
		}
		if guard == ai {
			slot = len(guards)
		}
		guards = append(guards, guard)
		if guard != ai && (guard.Combat.SwordRemaining > 0 || guard.Combat.Phase == DashWinding || guard.Combat.Phase == DashRushing) {
			busy = true
		}
	}
	if len(guards) <= 1 {
		return true, physics.Vector{X: target.X, Y: target.Y}
	}
	if w.sentinelTactics == nil {
		w.sentinelTactics = make(map[string]*sentinelTactic)
	}
	tactic := w.sentinelTactics[ai.den.RoomID]
	if tactic == nil {
		tactic = &sentinelTactic{}
		w.sentinelTactics[ai.den.RoomID] = tactic
	}
	var leader *aiState
	for _, guard := range guards {
		if guard.ID == tactic.leader {
			leader = guard
			if guard.Combat.SwordRemaining > 0 {
				tactic.struck = true
			}
		}
	}
	if leader == nil || w.tick >= tactic.until && !busy || tactic.struck && leader.Combat.SwordRemaining == 0 {
		for offset := range len(guards) {
			index := (tactic.next + offset) % len(guards)
			guard := guards[index]
			if guard.attackCooldown > 0 || guard.Combat.Phase != DashIdle || guard.Combat.SwordRemaining > 0 {
				continue
			}
			tactic.leader, tactic.next = guard.ID, (index+1)%len(guards)
			tactic.until, tactic.struck = w.tick+5*TickRate, false
			break
		}
	}
	angle := float64(slot)*2*math.Pi/float64(len(guards)) + .6*math.Sin(float64(w.tick)/TickRate*(.45+float64(slot)*.12)+float64(slot))
	radius := 115.0 + float64(slot%2)*35 + 20*math.Sin(float64(w.tick)/TickRate*.7+float64(slot)*2)
	point := physics.Vector{X: target.X + math.Cos(angle)*radius, Y: target.Y + math.Sin(angle)*radius}
	if room, ok := worldmap.Default().Room(ai.den.RoomID); ok {
		b := physics.Square(point.X, point.Y, AISize).ClampedTo(room.Bounds().Expanded(-AISize))
		point.X, point.Y = b.X, b.Y
	}
	geometry := worldmap.Default().WithSecrets(w.openSecrets)
	if !geometry.CanOccupy(physics.Square(point.X, point.Y, AISize)) {
		point = physics.Vector{X: ai.X, Y: ai.Y}
	}
	return ai.ID == tactic.leader && !busy, point
}

func (w *World) steerSentinel(ai *aiState, input Direction) Direction {
	if ai.Combat.SwordRemaining > 0 || ai.Combat.Phase == DashWinding || ai.Combat.Phase == DashRushing {
		ai.sentinelMoveX, ai.sentinelMoveY = 0, 0
		return input
	}
	x, y := input.X, input.Y
	for i := range w.ais {
		other := &w.ais[i]
		ally := other.Character == Triangle && other.den.RoomID == ai.den.RoomID
		if ai.Frenzied {
			ally = other.Frenzied
		}
		if other == ai || !ally || other.respawnTick != 0 || other.Stats.HP <= 0 {
			continue
		}
		dx, dy := ai.X-other.X, ai.Y-other.Y
		distance := math.Hypot(dx, dy)
		if distance >= 80 {
			continue
		}
		if distance < .01 {
			dx, dy = -1, -.5
			if ai.ID > other.ID {
				dx, dy = 1, .5
			}
			distance = math.Hypot(dx, dy)
		}
		weight := 1.6 * math.Pow(1-distance/80, 2)
		x, y = x+dx/distance*weight, y+dy/distance*weight
	}
	if length := math.Hypot(x, y); length > 1 {
		x, y = x/length, y/length
	}
	blend := 1 - math.Exp(-StepSeconds/.2)
	dx, dy := (x-ai.sentinelMoveX)*blend, (y-ai.sentinelMoveY)*blend
	if change := math.Hypot(dx, dy); change > StepSeconds*2.5 {
		scale := StepSeconds * 2.5 / change
		dx, dy = dx*scale, dy*scale
	}
	ai.sentinelMoveX += dx
	ai.sentinelMoveY += dy
	if math.Hypot(ai.sentinelMoveX, ai.sentinelMoveY) < .015 && math.Hypot(x, y) < .015 {
		ai.sentinelMoveX, ai.sentinelMoveY = 0, 0
	}
	input.X, input.Y = ai.sentinelMoveX, ai.sentinelMoveY
	return input
}
