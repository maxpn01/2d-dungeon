package simulation

import (
	"math"
	"math/rand/v2"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/physics"
	"raylib-2d-sandbox/internal/worldmap"
)

type MobType = worldmap.MobType

const (
	SpawnGuard     = worldmap.SpawnGuard
	DungeonShooter = worldmap.DungeonShooter
	DungeonPursuer = worldmap.DungeonPursuer
	DungeonHeavy   = worldmap.DungeonHeavy
	GateHeavy      = worldmap.GateHeavy
	Sentinel       = worldmap.Sentinel
)

func MobBody(t MobType, x, y float64) physics.Rect {
	size := float64(AISize)
	if t == GateHeavy {
		size *= parameters.GateHeavySizeMultiplier
	}
	offset := (size - AISize) / 2
	return physics.Square(x-offset, y-offset, size)
}

type MobMeleeAttack struct {
	Reach, Windup, Duration, DamageMultiplier float64
}

func MobMelee(t MobType) MobMeleeAttack {
	switch t {
	case DungeonPursuer:
		return MobMeleeAttack{parameters.PursuerMeleeReach, parameters.PursuerMeleeWindup, parameters.PursuerMeleeDuration, 1}
	case DungeonHeavy, GateHeavy:
		return MobMeleeAttack{parameters.HeavyMeleeReach, parameters.HeavyMeleeWindup, parameters.HeavyMeleeDuration, parameters.HeavyMeleeDamageMultiplier}
	case Sentinel:
		return MobMeleeAttack{SentinelSwordReach, SentinelSwordWindup, SentinelSwordDuration, 1}
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
		ai.shooterLastInput = input
		ai.shooterSightUntil = w.tick + parameters.ShooterLostSightTicks
	}()
	ai.patrolTicks, ai.patrolPause = 0, 0
	geometry := worldmap.Default().WithSecrets(w.openSecrets).WithVaultGate(w.orb.GateOpen)
	dx, dy := target.X-ai.X, target.Y-ai.Y
	distance := math.Hypot(dx, dy)
	d := Direction{AimX: dx / max(1, distance), AimY: dy / max(1, distance), Shoot: true}
	ai.shotTargetID, ai.Behavior = target.ID, "strafing"
	if ai.strafeDirection == 0 || w.tick >= ai.strafeUntil && w.tick >= ai.strafeTurnUntil {
		ai.strafeDirection = 1
		if rand.IntN(2) == 0 {
			ai.strafeDirection = -1
		}
		ai.strafeTurnUntil = w.tick + parameters.ShooterStrafeTurnCooldownTicks
		ai.strafeUntil = w.tick + uint64(parameters.ShooterStrafeMinimumTicks+rand.IntN(parameters.ShooterStrafeVariationTicks+1))
	}
	preferredDistance := (parameters.ShooterMinimumDistance + parameters.ShooterMaximumDistance) / 2
	if ai.shooterRangeMode < 0 && distance >= preferredDistance || ai.shooterRangeMode > 0 && distance <= preferredDistance {
		ai.shooterRangeMode = 0
	}
	if distance < parameters.ShooterMinimumDistance {
		ai.shooterRangeMode = -1
	} else if distance > parameters.ShooterMaximumDistance {
		ai.shooterRangeMode = 1
	}
	radial := float64(ai.shooterRangeMode) * .8
	if radial > 0 {
		ai.Behavior = "closing distance"
	} else if radial < 0 {
		ai.Behavior = "keeping distance"
	}
	for _, side := range []float64{ai.strafeDirection, -ai.strafeDirection} {
		if side != ai.strafeDirection && w.tick < ai.strafeTurnUntil {
			continue
		}
		move := Direction{
			X: d.AimX*radial - d.AimY*side*parameters.ShooterStrafeSpeed,
			Y: d.AimY*radial + d.AimX*side*parameters.ShooterStrafeSpeed,
		}
		if w.shotDodgeFits(ai, geometry, move, .2) {
			d.X, d.Y = move.X, move.Y
			if side != ai.strafeDirection {
				ai.strafeTurnUntil = w.tick + parameters.ShooterStrafeTurnCooldownTicks
			}
			ai.strafeDirection = side
			return d
		}
	}
	for _, side := range []float64{ai.strafeDirection, -ai.strafeDirection} {
		if side != ai.strafeDirection && w.tick < ai.strafeTurnUntil {
			continue
		}
		move := Direction{X: -d.AimY * side * parameters.ShooterStrafeSpeed, Y: d.AimX * side * parameters.ShooterStrafeSpeed}
		if w.shotDodgeFits(ai, geometry, move, .15) {
			d.X, d.Y = move.X, move.Y
			if side != ai.strafeDirection {
				ai.strafeTurnUntil = w.tick + parameters.ShooterStrafeTurnCooldownTicks
			}
			ai.strafeDirection = side
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
	if ai.Behavior == "dodging" {
		ai.shooterMoveX, ai.shooterMoveY = input.X, input.Y
		return input
	}
	blend := 1 - math.Exp(-StepSeconds/.12)
	dx, dy := (input.X-ai.shooterMoveX)*blend, (input.Y-ai.shooterMoveY)*blend
	if change := math.Hypot(dx, dy); change > StepSeconds*4 {
		scale := StepSeconds * 4 / change
		dx, dy = dx*scale, dy*scale
	}
	ai.shooterMoveX += dx
	ai.shooterMoveY += dy
	input.X, input.Y = ai.shooterMoveX, ai.shooterMoveY
	if math.Hypot(input.X, input.Y) < .01 || !w.shotDodgeFits(ai, worldmap.Default().WithSecrets(w.openSecrets).WithVaultGate(w.orb.GateOpen), input, StepSeconds) {
		input.X, input.Y = 0, 0
		ai.shooterMoveX, ai.shooterMoveY = 0, 0
	}
	return input
}

func (w *World) meleeGuardInput(ai *aiState, target *playerState) Direction {
	attack := MobMelee(ai.Type)
	if ai.Combat.StompRemaining > 0 {
		ai.Behavior = "stomping"
		if ai.Combat.StompRemaining > parameters.HeavyStompDuration-parameters.HeavyStompWindup {
			ai.Behavior = "winding stomp"
			if ai.Combat.StompRemaining > parameters.HeavyStompDuration-parameters.HeavyStompWindup+parameters.HeavyStompAimLock {
				if target := w.players[ai.aggressorID]; target != nil && w.aiCanSeePlayer(ai, target, parameters.HeavyStompReach+PlayerSize) {
					dx, dy := target.X-ai.X, target.Y-ai.Y
					if distance := math.Hypot(dx, dy); distance > 0 {
						ai.Combat.StompDirection = Direction{X: dx / distance, Y: dy / distance}
						ai.Combat.Facing = ai.Combat.StompDirection
					}
				}
			}
		}
		return Direction{AimX: ai.Combat.StompDirection.X, AimY: ai.Combat.StompDirection.Y}
	}
	if ai.Combat.SwordRemaining > 0 {
		ai.Behavior = "striking"
		if ai.Combat.SwordRemaining > attack.Duration-attack.Windup {
			ai.Behavior = "winding strike"
		}
		return Direction{AimX: ai.Combat.SwordDirection.X, AimY: ai.Combat.SwordDirection.Y}
	}
	if ai.Combat.Phase != DashIdle {
		ai.Behavior = "charging"
		if ai.Combat.Phase == DashWinding {
			ai.Behavior = "winding charge"
		}
		return Direction{AimX: ai.Combat.Direction.X, AimY: ai.Combat.Direction.Y}
	}
	if target == nil {
		return Direction{}
	}
	dx, dy := target.X-ai.X, target.Y-ai.Y
	distance := math.Hypot(dx, dy)
	d := Direction{AimX: dx / max(1, distance), AimY: dy / max(1, distance)}
	ai.aggressorID, ai.alertX, ai.alertY = target.ID, target.X, target.Y
	ai.alertUntil = w.tick + parameters.AIAlertSeconds*TickRate
	if ai.attackCooldown > 0 {
		ai.Behavior = "recovering strike"
		if distance > attack.Reach+PlayerSize/2 {
			movement := ai.navigateMap(worldmap.Default().WithSecrets(w.openSecrets).WithVaultGate(w.orb.GateOpen), target.X, target.Y)
			d.X, d.Y = movement.X*.5, movement.Y*.5
		}
		return d
	}
	if ai.Type.IsHeavy() && ai.mobChargeCooldown == 0 && ai.Stats.MP >= parameters.HeavyStompCost && distance <= parameters.HeavyStompReach+PlayerSize/2 && w.aiCanSeePlayer(ai, target, parameters.HeavyStompReach+PlayerSize) {
		ai.Stats.MP -= parameters.HeavyStompCost
		ai.Combat.StompRemaining = parameters.HeavyStompDuration
		ai.Combat.StompDirection = Direction{X: d.AimX, Y: d.AimY}
		ai.Combat.Facing = ai.Combat.StompDirection
		ai.mobChargeCooldown = parameters.HeavyStompCooldownTicks
		ai.attackCooldown = int(math.Ceil((parameters.HeavyStompDuration + 1/ai.Stats.AttackRate) * TickRate))
		ai.Behavior = "winding stomp"
		return d
	}
	if distance <= attack.Reach+PlayerSize/2-8 && w.aiCanSeePlayer(ai, target, attack.Reach+PlayerSize) {
		ai.Combat.SwordRemaining = attack.Duration
		ai.Combat.SwordDirection = Direction{X: d.AimX, Y: d.AimY}
		ai.attackCooldown = int(math.Ceil((attack.Duration + 1/ai.Stats.AttackRate) * TickRate))
		ai.Behavior = "winding strike"
		return d
	}
	movement := ai.navigateMap(worldmap.Default().WithSecrets(w.openSecrets).WithVaultGate(w.orb.GateOpen), target.X, target.Y)
	d.X, d.Y = movement.X, movement.Y
	ai.Behavior = "pursuing"
	return d
}
