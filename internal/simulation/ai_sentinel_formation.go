package simulation

import (
	"math"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/physics"
	"raylib-2d-sandbox/internal/worldmap"
)

type sentinelTactic struct {
	leader string
	until  uint64
}

func (w *World) sentinelFormation(ai *aiState, target *playerState) (bool, physics.Vector) {
	guards := make([]*aiState, 0, 4)
	slot, busy := 0, 0
	for i := range w.ais {
		guard := &w.ais[i]
		if guard.Type != Sentinel || guard.den.RoomID != ai.den.RoomID || guard.respawnTick != 0 || guard.Stats.HP <= 0 || guard.Stone || guard.Awakening > 0 {
			continue
		}
		if guard == ai {
			slot = len(guards)
		}
		guards = append(guards, guard)
		if guard != ai && (guard.Combat.SwordRemaining > 0 || guard.Combat.Phase == DashWinding || guard.Combat.Phase == DashRushing) {
			busy++
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
		}
	}
	if leader == nil || leader.Combat.SwordRemaining == 0 && leader.Combat.Phase == DashIdle && (w.tick >= tactic.until || leader.attackCooldown > 0 || !w.aiCanSeePlayer(leader, target, AIDetectionRadius)) {
		best := math.Inf(1)
		for _, guard := range guards {
			if guard.attackCooldown > 0 || guard.Combat.Phase != DashIdle || guard.Combat.SwordRemaining > 0 || !w.aiCanSeePlayer(guard, target, AIDetectionRadius) {
				continue
			}
			intercept := sentinelIntercept(target, .2)
			score := math.Hypot(intercept.X-guard.X, intercept.Y-guard.Y)
			if score < best {
				leader, best = guard, score
			}
		}
		if leader != nil {
			tactic.leader, tactic.until = leader.ID, w.tick+TickRate
		}
	}
	angle := float64(slot) * 2 * math.Pi / float64(len(guards))
	if math.Hypot(target.velocityX, target.velocityY) > 40 {
		angle += math.Atan2(target.velocityY, target.velocityX)
	}
	intercept := sentinelIntercept(target, .25)
	point := physics.Vector{X: intercept.X + math.Cos(angle)*85, Y: intercept.Y + math.Sin(angle)*85}
	if room, ok := worldmap.Default().Room(ai.den.RoomID); ok {
		b := physics.Square(point.X, point.Y, AISize).ClampedTo(room.Bounds().Expanded(-AISize))
		point.X, point.Y = b.X, b.Y
	}
	geometry := worldmap.Default().WithSecrets(w.openSecrets)
	if !geometry.CanOccupy(physics.Square(point.X, point.Y, AISize)) {
		point = physics.Vector{X: target.X, Y: target.Y}
	}
	close := math.Hypot(target.X-ai.X, target.Y-ai.Y) <= SentinelSwordReach+PlayerSize/2+16
	return busy < parameters.SentinelConcurrentAttackers && (ai.ID == tactic.leader || close), point
}

func (w *World) steerSentinel(ai *aiState, input Direction) Direction {
	if ai.Combat.SwordRemaining > 0 || ai.Combat.Phase == DashWinding || ai.Combat.Phase == DashRushing || input.Dash || ai.Behavior == "pursuing" {
		ai.sentinelMoveX, ai.sentinelMoveY = 0, 0
		return input
	}
	x, y := input.X, input.Y
	for i := range w.ais {
		other := &w.ais[i]
		ally := other.Type == Sentinel && other.den.RoomID == ai.den.RoomID
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
