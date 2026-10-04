package simulation

import (
	"math"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/physics"
	"raylib-2d-sandbox/internal/worldmap"
)

const (
	SentinelSwordReach    = AISize/2 + 48.0
	SentinelSwordWindup   = 0.3
	SentinelSwordDuration = 0.55
)

func (w *World) sentinelInput(ai *aiState) Direction {
	ai.attackCooldown = max(0, ai.attackCooldown-1)
	if ai.Combat.SwordRemaining > 0 {
		ai.Behavior = "striking"
		if ai.Combat.SwordRemaining > SentinelSwordDuration-SentinelSwordWindup {
			ai.Behavior = "winding sword"
		}
		return Direction{AimX: ai.Combat.SwordDirection.X, AimY: ai.Combat.SwordDirection.Y}
	}
	if ai.returning {
		if d, returning := w.investigateAI(ai); returning {
			return d
		}
	}
	target := w.rangedTarget(ai)
	if target == nil {
		if d, investigating := w.investigateAI(ai); investigating {
			return d
		}
		return ai.patrolInput()
	}
	dx, dy := target.X-ai.X, target.Y-ai.Y
	distance := math.Hypot(dx, dy)
	d := Direction{AimX: dx / max(1, distance), AimY: dy / max(1, distance)}
	attacking, flank := w.sentinelFormation(ai, target)
	if !attacking || ai.attackCooldown > 0 {
		ai.Behavior = "flanking"
		movement := ai.navigateMap(worldmap.Default().WithSecrets(w.openSecrets), flank.X, flank.Y)
		if distance := math.Hypot(flank.X-ai.X, flank.Y-ai.Y); distance > 4 {
			speed := min(1, distance/70)
			d.X, d.Y = movement.X*speed, movement.Y*speed
		}
		return d
	}
	if distance <= SentinelSwordReach+PlayerSize/2-8 {
		ai.Behavior = "attacking"
		if ai.Combat.Phase == DashIdle && ai.attackCooldown == 0 {
			ai.Combat.SwordRemaining = SentinelSwordDuration
			ai.Combat.SwordDirection = Direction{X: d.AimX, Y: d.AimY}
			ai.attackCooldown = int(math.Ceil((SentinelSwordDuration + 1/ai.Stats.AttackRate) * TickRate))
			ai.Behavior = "winding sword"
		}
		return d
	}
	ai.Behavior = "pursuing"
	ai.alertX, ai.alertY = target.X, target.Y
	ai.aggressorID, ai.alertUntil = target.ID, w.tick+3*TickRate
	geometry := worldmap.Default().WithSecrets(w.openSecrets)
	movement := ai.navigateMap(geometry, target.X, target.Y)
	d.X, d.Y = movement.X, movement.Y
	reach := ai.Stats.MovementSpeed() * SpeedScale * DashSpeedMultiplier * DashBurst
	if ai.Combat.Phase == DashIdle && !ai.Combat.Held && ai.Stats.MP >= DashCost && distance <= reach+SentinelSwordReach && distance > reach && ai.attackCooldown == 0 {
		ex, ey := geometry.Move(ai.X, ai.Y, d.AimX*reach, d.AimY*reach, AISize, AISize)
		if math.Hypot(ex-ai.X-d.AimX*reach, ey-ai.Y-d.AimY*reach) < .1 && geometry.RoomContains(ai.den.RoomID, physics.Square(ex, ey, AISize)) {
			d.Dash = true
			ai.attackCooldown = TickRate
		}
	}
	return d
}

func (w *World) advanceSentinelSwords(actors []eater) {
	for i := range w.ais {
		ai := &w.ais[i]
		if ai.Character != Triangle || ai.respawnTick != 0 || ai.Stats.HP <= 0 || ai.Combat.SwordRemaining <= 0 {
			continue
		}
		before := ai.Combat.SwordRemaining
		ai.Combat.SwordRemaining = max(0, before-StepSeconds)
		impact := SentinelSwordDuration - SentinelSwordWindup
		if before <= impact || ai.Combat.SwordRemaining > impact {
			continue
		}
		for _, target := range actors {
			if target.player == nil || target.dead() || target.protectedAt(1, w.safeZones) {
				continue
			}
			dx := target.end.x + target.size/2 - ai.X - AISize/2
			dy := target.end.y + target.size/2 - ai.Y - AISize/2
			distance := math.Hypot(dx, dy)
			facing := ai.Combat.SwordDirection
			if distance > SentinelSwordReach+target.size/2 || (dx*facing.X+dy*facing.Y)/max(1, distance) < math.Cos(math.Pi/3) || !w.aiCanSeePlayer(ai, target.player, SentinelSwordReach+target.size) {
				continue
			}
			hp := target.stats.HP
			target.stats.HP = max(0, hp-ai.Stats.Damage)
			if target.stats.HP == 0 {
				w.dropKillLoot(target, hp, target.ownerStats().MP)
				ai.Stats.GainXP(float64(target.stats.Level) * parameters.KillXPPerLevel)
				target.die(ai.ID, w.tick)
			}
		}
	}
}
