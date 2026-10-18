package simulation

import (
	"math"

	"2d-dungeon/internal/parameters"
)

func (w *World) steerSentinel(ai *aiState, input Direction) Direction {
	if ai.Combat.SwordRemaining > 0 || ai.Combat.Phase == DashWinding || ai.Combat.Phase == DashRushing || input.Dash {
		ai.sentinel.moveX, ai.sentinel.moveY = 0, 0
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
	if ai.Behavior == BehaviorPursuing {
		ai.sentinel.moveX, ai.sentinel.moveY = x, y
		input.X, input.Y = x, y
		return input
	}
	blend := 1 - math.Exp(-parameters.StepSeconds/.2)
	dx, dy := (x-ai.sentinel.moveX)*blend, (y-ai.sentinel.moveY)*blend
	if change := math.Hypot(dx, dy); change > parameters.StepSeconds*2.5 {
		scale := parameters.StepSeconds * 2.5 / change
		dx, dy = dx*scale, dy*scale
	}
	ai.sentinel.moveX += dx
	ai.sentinel.moveY += dy
	if math.Hypot(ai.sentinel.moveX, ai.sentinel.moveY) < .015 && math.Hypot(x, y) < .015 {
		ai.sentinel.moveX, ai.sentinel.moveY = 0, 0
	}
	input.X, input.Y = ai.sentinel.moveX, ai.sentinel.moveY
	return input
}
