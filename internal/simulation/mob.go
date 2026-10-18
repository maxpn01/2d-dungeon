package simulation

import (
	"math"
	"math/rand/v2"

	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/physics"
	"2d-dungeon/internal/worldmap"
)

type MobType = worldmap.MobType

const (
	SpawnGuard     = worldmap.SpawnGuard
	DungeonShooter = worldmap.DungeonShooter
	DungeonPursuer = worldmap.DungeonPursuer
	DungeonHeavy   = worldmap.DungeonHeavy
	GateHeavy      = worldmap.GateHeavy
	Sentinel       = worldmap.Sentinel
	ForestSnake    = worldmap.ForestSnake
	CaveHydra      = worldmap.CaveHydra
)

func MobBody(t MobType, x, y float64) physics.Rect {
	size := float64(parameters.AISize)
	if t == GateHeavy {
		size *= parameters.GateHeavySizeMultiplier
	}
	offset := (size - parameters.AISize) / 2
	return physics.Square(x-offset, y-offset, size)
}

type MobMeleeAttack struct {
	Reach, Windup, Duration, Active, DamageMultiplier, AdvanceMultiplier float64
}

func MobMelee(t MobType) MobMeleeAttack {
	switch t {
	case DungeonPursuer:
		return MobMeleeAttack{
			Reach: parameters.PursuerMeleeReach, Windup: parameters.PursuerMeleeWindup, Duration: parameters.PursuerMeleeDuration,
			Active: parameters.PursuerMeleeActive, DamageMultiplier: 1, AdvanceMultiplier: parameters.PursuerMeleeAdvanceMultiplier,
		}
	case DungeonHeavy, GateHeavy:
		return MobMeleeAttack{
			Reach: parameters.HeavyMeleeReach, Windup: parameters.HeavyMeleeWindup, Duration: parameters.HeavyMeleeDuration,
			Active: parameters.HeavyMeleeActive, DamageMultiplier: parameters.HeavyMeleeDamageMultiplier, AdvanceMultiplier: parameters.HeavyMeleeAdvanceMultiplier,
		}
	case ForestSnake:
		return MobMeleeAttack{
			Reach: parameters.SnakeBiteReach, Windup: parameters.SnakeBiteWindup, Duration: parameters.SnakeBiteDuration,
			Active: parameters.SnakeBiteActive, DamageMultiplier: parameters.SnakeBiteDamage, AdvanceMultiplier: parameters.SnakeBiteAdvance,
		}
	case Sentinel:
		return MobMeleeAttack{
			Reach: SentinelSwordReach, Windup: SentinelSwordWindup, Duration: SentinelSwordDuration,
			Active: parameters.SentinelMeleeActive, DamageMultiplier: 1, AdvanceMultiplier: parameters.SentinelMeleeAdvanceMultiplier,
		}
	}
	return MobMeleeAttack{}
}

func newMobStats(level int, t MobType) Stats {
	stats := NewMobStatsAtLevel(level)
	stats.MobType = t
	stats.rebuildPerks()
	stats.HP, stats.MP = stats.MaxHP(), stats.MaxMP()
	return stats
}

func (w *World) shooterInput(ai *aiState, target *playerState) (input Direction) {
	defer func() {
		ai.shooter.lastInput = input
		ai.shooter.sightUntil = w.tick + parameters.ShooterLostSightTicks
	}()
	ai.patrol.ticks, ai.patrol.pause = 0, 0
	geometry := w.geometry()
	dx, dy := target.X-ai.X, target.Y-ai.Y
	distance := math.Hypot(dx, dy)
	d := Direction{AimX: dx / max(1, distance), AimY: dy / max(1, distance), Shoot: true}
	ai.shotTargetID, ai.Behavior = target.ID, BehaviorStrafing
	if ai.shooter.strafeDirection == 0 || w.tick >= ai.shooter.strafeUntil && w.tick >= ai.shooter.strafeTurnUntil {
		ai.shooter.strafeDirection = 1
		if rand.IntN(2) == 0 {
			ai.shooter.strafeDirection = -1
		}
		ai.shooter.strafeTurnUntil = w.tick + parameters.ShooterStrafeTurnCooldownTicks
		ai.shooter.strafeUntil = w.tick + uint64(parameters.ShooterStrafeMinimumTicks+rand.IntN(parameters.ShooterStrafeVariationTicks+1))
	}
	preferredDistance := (parameters.ShooterMinimumDistance + parameters.ShooterMaximumDistance) / 2
	if ai.shooter.rangeMode < 0 && distance >= preferredDistance || ai.shooter.rangeMode > 0 && distance <= preferredDistance {
		ai.shooter.rangeMode = 0
	}
	if distance < parameters.ShooterMinimumDistance {
		ai.shooter.rangeMode = -1
	} else if distance > parameters.ShooterMaximumDistance {
		ai.shooter.rangeMode = 1
	}
	radial := float64(ai.shooter.rangeMode) * .8
	if radial > 0 {
		ai.Behavior = BehaviorClosingDistance
	} else if radial < 0 {
		ai.Behavior = BehaviorKeepingDistance
	}
	for _, side := range []float64{ai.shooter.strafeDirection, -ai.shooter.strafeDirection} {
		if side != ai.shooter.strafeDirection && w.tick < ai.shooter.strafeTurnUntil {
			continue
		}
		move := Direction{
			X: d.AimX*radial - d.AimY*side*parameters.ShooterStrafeSpeed,
			Y: d.AimY*radial + d.AimX*side*parameters.ShooterStrafeSpeed,
		}
		if w.shotDodgeFits(ai, geometry, move, .2) {
			d.X, d.Y = move.X, move.Y
			if side != ai.shooter.strafeDirection {
				ai.shooter.strafeTurnUntil = w.tick + parameters.ShooterStrafeTurnCooldownTicks
			}
			ai.shooter.strafeDirection = side
			return d
		}
	}
	for _, side := range []float64{ai.shooter.strafeDirection, -ai.shooter.strafeDirection} {
		if side != ai.shooter.strafeDirection && w.tick < ai.shooter.strafeTurnUntil {
			continue
		}
		move := Direction{X: -d.AimY * side * parameters.ShooterStrafeSpeed, Y: d.AimX * side * parameters.ShooterStrafeSpeed}
		if w.shotDodgeFits(ai, geometry, move, .15) {
			d.X, d.Y = move.X, move.Y
			if side != ai.shooter.strafeDirection {
				ai.shooter.strafeTurnUntil = w.tick + parameters.ShooterStrafeTurnCooldownTicks
			}
			ai.shooter.strafeDirection = side
			return d
		}
	}
	if radial > 0 {
		movement := ai.navigateMap(geometry, target.X, target.Y)
		d.X, d.Y = movement.X, movement.Y
	}
	return d
}

func (w *World) steerShooter(ai *aiState, input Direction) Direction {
	if ai.Behavior == BehaviorDodging {
		ai.shooter.moveX, ai.shooter.moveY = input.X, input.Y
		return input
	}
	blend := 1 - math.Exp(-parameters.StepSeconds/.12)
	dx, dy := (input.X-ai.shooter.moveX)*blend, (input.Y-ai.shooter.moveY)*blend
	if change := math.Hypot(dx, dy); change > parameters.StepSeconds*4 {
		scale := parameters.StepSeconds * 4 / change
		dx, dy = dx*scale, dy*scale
	}
	ai.shooter.moveX += dx
	ai.shooter.moveY += dy
	input.X, input.Y = ai.shooter.moveX, ai.shooter.moveY
	if math.Hypot(input.X, input.Y) < .01 || !w.shotDodgeFits(ai, w.geometry(), input, parameters.StepSeconds) {
		input.X, input.Y = 0, 0
		ai.shooter.moveX, ai.shooter.moveY = 0, 0
	}
	return input
}

func (w *World) meleeGuardInput(ai *aiState, target *playerState) Direction {
	attack := MobMelee(ai.Type)
	stomp := MobStomp(ai.Type)
	if ai.Combat.StompRemaining > 0 {
		ai.Behavior = BehaviorStomping
		if ai.Combat.StompRemaining > parameters.HeavyStompDuration-parameters.HeavyStompWindup {
			ai.Behavior = BehaviorWindingStomp
			if ai.Combat.StompRemaining > parameters.HeavyStompDuration-parameters.HeavyStompWindup+parameters.HeavyStompAimLock {
				if target := w.playerWithID(ai.alert.aggressorID); target != nil && w.aiCanSeePlayer(ai, target, stomp.Reach+parameters.PlayerSize) {
					impactIn := ai.Combat.StompRemaining - (parameters.HeavyStompDuration - parameters.HeavyStompWindup)
					travel := math.Hypot(target.X-ai.X, target.Y-ai.Y) / parameters.HeavyStompWaveSpeed
					intercept := meleeIntercept(target, impactIn+travel)
					dx, dy := intercept.X-ai.X, intercept.Y-ai.Y
					if distance := math.Hypot(dx, dy); distance > 0 {
						ai.Combat.StompDirection = Direction{X: dx / distance, Y: dy / distance}
						ai.Combat.Facing = ai.Combat.StompDirection
					}
				}
			}
		}
		d := Direction{AimX: ai.Combat.StompDirection.X, AimY: ai.Combat.StompDirection.Y}
		if target := w.playerWithID(ai.alert.aggressorID); target != nil && !target.Dead {
			movement := ai.navigateMap(w.geometry(), target.X, target.Y)
			d.X, d.Y = movement.X, movement.Y
		}
		return limitMeleeAdvance(ai, w.playerWithID(ai.alert.aggressorID), d, 1)
	}
	if ai.Combat.SwordRemaining > 0 {
		ai.Behavior = BehaviorStriking
		elapsed := attack.Duration - ai.Combat.SwordRemaining
		if elapsed < attack.Windup {
			ai.Behavior = BehaviorWindingStrike
			if target := w.playerWithID(ai.alert.aggressorID); target != nil && elapsed < attack.Windup-parameters.MeleeAimLock && w.aiCanSeePlayer(ai, target, attack.Reach+parameters.PlayerSize+80) {
				trackMeleeAim(ai, target, attack.Windup-elapsed, parameters.MeleeTurnRate)
			}
		}
		d := Direction{AimX: ai.Combat.SwordDirection.X, AimY: ai.Combat.SwordDirection.Y}
		if elapsed < attack.Windup+attack.Active {
			d.X, d.Y = d.AimX, d.AimY
		} else if target := w.playerWithID(ai.alert.aggressorID); target != nil && !target.Dead {
			movement := ai.navigateMap(w.geometry(), target.X, target.Y)
			d.X, d.Y = movement.X, movement.Y
		}
		multiplier := 1.0
		if elapsed < attack.Windup+attack.Active {
			multiplier = attack.AdvanceMultiplier
		}
		return limitMeleeAdvance(ai, w.playerWithID(ai.alert.aggressorID), d, multiplier)
	}

	if ai.Combat.Phase != DashIdle {
		ai.Behavior = BehaviorCharging
		if ai.Combat.Phase == DashWinding {
			ai.Behavior = BehaviorWindingCharge
		}
		return Direction{AimX: ai.Combat.Direction.X, AimY: ai.Combat.Direction.Y}
	}
	if target == nil {
		return Direction{}
	}
	dx, dy := target.X-ai.X, target.Y-ai.Y
	distance := math.Hypot(dx, dy)
	d := Direction{AimX: dx / max(1, distance), AimY: dy / max(1, distance)}
	ai.alert.aggressorID, ai.alert.x, ai.alert.y = target.ID, target.X, target.Y
	ai.alert.until = w.tick + parameters.AIAlertSeconds*parameters.TickRate
	if ai.attackCooldown > 0 {
		ai.Behavior = BehaviorRecoveringStrike
		movement := ai.navigateMap(w.geometry(), target.X, target.Y)
		d.X, d.Y = movement.X, movement.Y
		return limitMeleeAdvance(ai, target, d, 1)
	}
	canStomp := ai.Type.IsHeavy() && ai.mobChargeCooldown == 0 && ai.Stats.MP >= parameters.HeavyStompCost
	inStompRange := distance > attack.Reach+parameters.PlayerSize/2-8 && distance <= stomp.Reach+parameters.PlayerSize/2
	if canStomp && inStompRange && w.aiCanSeePlayer(ai, target, stomp.Reach+parameters.PlayerSize) {
		ai.Stats.MP -= parameters.HeavyStompCost
		ai.Combat.StompRemaining = parameters.HeavyStompDuration
		intercept := meleeIntercept(target, parameters.HeavyStompWindup+distance/parameters.HeavyStompWaveSpeed)
		sx, sy := intercept.X-ai.X, intercept.Y-ai.Y
		length := math.Hypot(sx, sy)
		ai.Combat.StompDirection = Direction{X: sx / max(1, length), Y: sy / max(1, length)}
		ai.Combat.Facing = ai.Combat.StompDirection
		ai.mobChargeCooldown = parameters.HeavyStompCooldownTicks
		ai.attackCooldown = int(math.Ceil((parameters.HeavyStompDuration + 1/ai.Stats.AttackRate) * parameters.TickRate))
		ai.Behavior = BehaviorWindingStomp
		movement := ai.navigateMap(w.geometry(), target.X, target.Y)
		d.X, d.Y = movement.X, movement.Y
		return limitMeleeAdvance(ai, target, d, 1)
	}
	if distance <= attack.Reach+parameters.PlayerSize/2-8 && w.aiCanSeePlayer(ai, target, attack.Reach+parameters.PlayerSize) {
		intercept := meleeIntercept(target, attack.Windup)
		sx, sy := intercept.X-ai.X, intercept.Y-ai.Y
		length := math.Hypot(sx, sy)
		d.AimX, d.AimY = sx/max(1, length), sy/max(1, length)
		ai.Combat.SwordRemaining = attack.Duration
		ai.Combat.SwordDirection = Direction{X: d.AimX, Y: d.AimY}
		ai.meleeHits = nil
		d.X, d.Y = d.AimX, d.AimY
		ai.attackCooldown = int(math.Ceil((attack.Duration + 1/ai.Stats.AttackRate) * parameters.TickRate))
		ai.Behavior = BehaviorWindingStrike
		return limitMeleeAdvance(ai, target, d, attack.AdvanceMultiplier)
	}
	movement := ai.navigateMap(w.geometry(), target.X, target.Y)
	d.X, d.Y = movement.X, movement.Y
	ai.Behavior = BehaviorPursuing
	return d
}
