package simulation

import (
	"math"
	"math/rand/v2"

	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/physics"
	"2d-dungeon/internal/worldmap"
)

func (w *World) avoidPlayerShots(ai *aiState, input Direction) Direction {
	ai.dodge.cooldown = max(0, ai.dodge.cooldown-1)
	if ai.patrol.returning || ai.Combat.Phase != DashIdle || ai.Combat.SwordRemaining > 0 || ai.Combat.StompRemaining > 0 {
		ai.dodge.evadeTicks, ai.dodge.reactionTicks = 0, 0
		return input
	}
	if ai.dodge.reactionTicks > 0 {
		ai.dodge.reactionTicks--
		return input
	}
	geometry := w.geometry()
	if ai.dodge.evadeTicks > 0 {
		ai.dodge.evadeTicks--
		if !w.shotDodgeFits(ai, geometry, ai.dodge.evadeDirection, parameters.StepSeconds) {
			ai.dodge.evadeTicks = 0
			return input
		}
		input.X, input.Y = ai.dodge.evadeDirection.X, ai.dodge.evadeDirection.Y
		input.Dash, input.Boost = false, false
		ai.Behavior = BehaviorDodging
		return input
	}
	if ai.dodge.cooldown > 0 || parameters.AIShotDodgeChance <= 0 {
		return input
	}
	shot, found := w.incomingPlayerShot(ai, input, geometry)
	if !found {
		return input
	}
	ai.dodge.cooldown = parameters.AIShotDodgeCooldownTicks
	chance := parameters.AIShotDodgeChance
	if ai.Type.IsHeavy() {
		chance *= parameters.HeavyDodgeChanceMultiplier
	}
	if rand.Float64() >= chance {
		return input
	}
	duration := float64(parameters.AIShotDodgeDurationTicks) * parameters.StepSeconds
	best := -math.MaxFloat64
	var direction Direction
	for _, d := range []Direction{{X: -shot.Direction.Y, Y: shot.Direction.X}, {X: shot.Direction.Y, Y: -shot.Direction.X}} {
		if !w.shotDodgeFits(ai, geometry, d, duration) {
			continue
		}
		x := ai.X + parameters.AISize/2 + d.X*ai.Stats.MovementSpeed()*parameters.SpeedScale*duration
		y := ai.Y + parameters.AISize/2 + d.Y*ai.Stats.MovementSpeed()*parameters.SpeedScale*duration
		separation := math.Abs((x-shot.X-shot.Size/2)*shot.Direction.Y - (y-shot.Y-shot.Size/2)*shot.Direction.X)
		score := separation + 5*(d.X*input.X+d.Y*input.Y)
		if score > best {
			best, direction = score, d
		}
	}
	if best == -math.MaxFloat64 {
		return input
	}
	ai.dodge.evadeDirection = direction
	ai.dodge.evadeTicks = parameters.AIShotDodgeDurationTicks
	ai.dodge.reactionTicks = max(0, parameters.AIShotDodgeReactionTicks-1)
	return input
}

func (w *World) shotDodgeFits(ai *aiState, geometry *worldmap.Map, d Direction, duration float64) bool {
	distance := ai.Stats.MovementSpeed() * parameters.SpeedScale * duration
	x, y := ai.X+d.X*distance, ai.Y+d.Y*distance
	body := MobBody(ai.Type, ai.X, ai.Y)
	motion := geometry.MoveBody(body, physics.Vector{X: d.X * distance, Y: d.Y * distance})
	mx, my := motion.Bounds.X+ai.X-body.X, motion.Bounds.Y+ai.Y-body.Y
	if math.Hypot(mx-x, my-y) > .1 || geometry.PrivateGateCrossed(ai.X, ai.Y, x, y, parameters.AISize) {
		return false
	}
	if ai.Type == DungeonShooter && !geometry.OverlapsCover(physics.Square(ai.X, ai.Y, parameters.AISize)) {
		cover := physics.World{CanOccupy: func(body physics.Rect) bool {
			return !geometry.OverlapsCover(body)
		}}
		if _, blocked := cover.MovementLimit(physics.Square(ai.X, ai.Y, parameters.AISize), physics.Vector{X: x - ai.X, Y: y - ai.Y}); blocked {
			return false
		}
	}
	candidate := *ai
	candidate.X, candidate.Y = x, y
	return ai.den.ID == "" || w.aiWithinTerritory(&candidate)
}

func (w *World) incomingPlayerShot(ai *aiState, input Direction, geometry *worldmap.Map) (Projectile, bool) {
	var selected Projectile
	soonest := math.Inf(1)
	for _, shot := range w.projectiles {
		if shot.MobOwned || shot.Kind != "" || w.playerWithID(shot.Owner) == nil {
			continue
		}
		if math.Hypot(shot.X+shot.Size/2-ai.X-parameters.AISize/2, shot.Y+shot.Size/2-ai.Y-parameters.AISize/2) > parameters.AIDetectionRadius {
			continue
		}
		if !geometry.Visible(ai.X+parameters.AISize/2, ai.Y+parameters.AISize/2, shot.X+shot.Size/2, shot.Y+shot.Size/2) {
			continue
		}
		duration := min(parameters.AIShotDodgeLookahead, shot.Remaining)
		if duration <= 0 {
			continue
		}
		travel := ProjectileSpeed(shot) * duration
		start := position{shot.X, shot.Y}
		end := position{shot.X + shot.Direction.X*travel, shot.Y + shot.Direction.Y*travel}
		speed := ai.Stats.MovementSpeed() * parameters.SpeedScale
		targetEnd := position{ai.X + input.X*speed*duration, ai.Y + input.Y*speed*duration}
		body := MobBody(ai.Type, ai.X, ai.Y)
		offset := ai.X - body.X
		targetEnd.x -= offset
		targetEnd.y -= offset
		fraction, hit := eatingContact(start, end, shot.Size, position{body.X, body.Y}, targetEnd, body.W)
		if !hit || fraction*duration >= soonest {
			continue
		}
		wallFraction, blocked := geometry.MovementLimit(shot.X, shot.Y, end.x-shot.X, end.y-shot.Y, shot.Size, shot.Size)
		if blocked && wallFraction <= fraction || geometry.PrivateGateCrossed(shot.X, shot.Y, end.x, end.y, shot.Size) {
			continue
		}
		selected, soonest = shot, fraction*duration
	}
	return selected, !math.IsInf(soonest, 1)
}
