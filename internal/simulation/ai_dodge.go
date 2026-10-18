package simulation

import (
	"math"
	"math/rand/v2"

	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/physics"
	"2d-dungeon/internal/worldmap"
)

func (w *World) avoidPlayerShots(ai *aiState, input Direction) Direction {
	ai.shotDodgeCooldown = max(0, ai.shotDodgeCooldown-1)
	if ai.returning || ai.Combat.Phase != DashIdle || ai.Combat.SwordRemaining > 0 || ai.Combat.StompRemaining > 0 {
		ai.evadeTicks, ai.shotDodgeReactionTicks = 0, 0
		return input
	}
	if ai.shotDodgeReactionTicks > 0 {
		ai.shotDodgeReactionTicks--
		return input
	}
	geometry := worldmap.Default().WithSecrets(w.openSecrets).WithVaultGate(w.orb.GateOpen)
	if ai.evadeTicks > 0 {
		ai.evadeTicks--
		if !w.shotDodgeFits(ai, geometry, ai.evadeDirection, StepSeconds) {
			ai.evadeTicks = 0
			return input
		}
		input.X, input.Y = ai.evadeDirection.X, ai.evadeDirection.Y
		input.Dash, input.Boost = false, false
		ai.Behavior = "dodging"
		return input
	}
	if ai.shotDodgeCooldown > 0 || parameters.AIShotDodgeChance <= 0 {
		return input
	}
	shot, found := w.incomingPlayerShot(ai, input, geometry)
	if !found {
		return input
	}
	ai.shotDodgeCooldown = parameters.AIShotDodgeCooldownTicks
	chance := parameters.AIShotDodgeChance
	if ai.Type.IsHeavy() {
		chance *= parameters.HeavyDodgeChanceMultiplier
	}
	if rand.Float64() >= chance {
		return input
	}
	duration := float64(parameters.AIShotDodgeDurationTicks) * StepSeconds
	best := -math.MaxFloat64
	var direction Direction
	for _, d := range []Direction{{X: -shot.Direction.Y, Y: shot.Direction.X}, {X: shot.Direction.Y, Y: -shot.Direction.X}} {
		if !w.shotDodgeFits(ai, geometry, d, duration) {
			continue
		}
		x := ai.X + AISize/2 + d.X*ai.Stats.MovementSpeed()*SpeedScale*duration
		y := ai.Y + AISize/2 + d.Y*ai.Stats.MovementSpeed()*SpeedScale*duration
		separation := math.Abs((x-shot.X-shot.Size/2)*shot.Direction.Y - (y-shot.Y-shot.Size/2)*shot.Direction.X)
		score := separation + 5*(d.X*input.X+d.Y*input.Y)
		if score > best {
			best, direction = score, d
		}
	}
	if best == -math.MaxFloat64 {
		return input
	}
	ai.evadeDirection = direction
	ai.evadeTicks = parameters.AIShotDodgeDurationTicks
	ai.shotDodgeReactionTicks = max(0, parameters.AIShotDodgeReactionTicks-1)
	return input
}

func (w *World) shotDodgeFits(ai *aiState, geometry *worldmap.Map, d Direction, duration float64) bool {
	distance := ai.Stats.MovementSpeed() * SpeedScale * duration
	x, y := ai.X+d.X*distance, ai.Y+d.Y*distance
	body := MobBody(ai.Type, ai.X, ai.Y)
	motion := geometry.MoveBody(body, physics.Vector{X: d.X * distance, Y: d.Y * distance})
	mx, my := motion.Bounds.X+ai.X-body.X, motion.Bounds.Y+ai.Y-body.Y
	if math.Hypot(mx-x, my-y) > .1 || geometry.PrivateGateCrossed(ai.X, ai.Y, x, y, AISize) {
		return false
	}
	if ai.Type == DungeonShooter && !geometry.OverlapsCover(physics.Square(ai.X, ai.Y, AISize)) {
		cover := physics.World{CanOccupy: func(body physics.Rect) bool {
			return !geometry.OverlapsCover(body)
		}}
		if _, blocked := cover.MovementLimit(physics.Square(ai.X, ai.Y, AISize), physics.Vector{X: x - ai.X, Y: y - ai.Y}); blocked {
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
		if shot.MobOwned || shot.Kind != "" || w.players[shot.Owner] == nil || math.Hypot(shot.X+shot.Size/2-ai.X-AISize/2, shot.Y+shot.Size/2-ai.Y-AISize/2) > AIDetectionRadius {
			continue
		}
		if !geometry.Visible(ai.X+AISize/2, ai.Y+AISize/2, shot.X+shot.Size/2, shot.Y+shot.Size/2) {
			continue
		}
		duration := min(parameters.AIShotDodgeLookahead, shot.Remaining)
		if duration <= 0 {
			continue
		}
		travel := ProjectileSpeed(shot) * duration
		start := position{shot.X, shot.Y}
		end := position{shot.X + shot.Direction.X*travel, shot.Y + shot.Direction.Y*travel}
		speed := ai.Stats.MovementSpeed() * SpeedScale
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
