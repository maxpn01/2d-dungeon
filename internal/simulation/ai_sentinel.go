package simulation

import (
	"math"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/physics"
	"raylib-2d-sandbox/internal/worldmap"
)

const (
	SentinelSwordReach    = AISize/2 + 65.0
	SentinelSwordWindup   = 0.14
	SentinelSwordDuration = 0.34
)

func (w *World) sentinelInput(ai *aiState) Direction {
	ai.attackCooldown = max(0, ai.attackCooldown-1)
	target := w.sentinelTarget(ai)
	if ai.Combat.SwordRemaining > 0 {
		ai.Behavior = "striking"
		elapsed := SentinelSwordDuration - ai.Combat.SwordRemaining
		if elapsed < SentinelSwordWindup {
			ai.Behavior = "winding sword"
			if target != nil && elapsed < parameters.SentinelSwordTrackingDuration {
				trackMeleeAim(ai, target, SentinelSwordWindup-elapsed, parameters.SentinelSwordTurnRate)
			}
		}
		d := Direction{AimX: ai.Combat.SwordDirection.X, AimY: ai.Combat.SwordDirection.Y}
		if elapsed < SentinelSwordWindup+parameters.SentinelMeleeActive {
			d.X, d.Y = d.AimX*parameters.SentinelSwordAdvanceMultiplier, d.AimY*parameters.SentinelSwordAdvanceMultiplier
		} else if target != nil {
			movement := ai.navigateMap(worldmap.Default().WithSecrets(w.openSecrets).WithVaultGate(w.orb.GateOpen), target.X, target.Y)
			d.X, d.Y = movement.X, movement.Y
		}
		multiplier := 1.0
		if elapsed < SentinelSwordWindup+parameters.SentinelMeleeActive {
			multiplier = MobMelee(ai.Type).AdvanceMultiplier
		}
		return limitMeleeAdvance(ai, target, d, multiplier)
	}

	if ai.Combat.Phase == DashWinding || ai.Combat.Phase == DashRushing {
		ai.Behavior = "charging"
		return Direction{AimX: ai.Combat.Direction.X, AimY: ai.Combat.Direction.Y}
	}
	if target != nil {
		ai.returning = false
	}
	if ai.returning {
		if d, returning := w.investigateAI(ai); returning {
			return d
		}
	}
	if target == nil {
		if d, investigating := w.investigateAI(ai); investigating {
			return d
		}
		return ai.patrolInput(worldmap.Default().WithSecrets(w.openSecrets).WithVaultGate(w.orb.GateOpen))
	}
	ai.alertX, ai.alertY = target.X, target.Y
	ai.aggressorID, ai.alertUntil = target.ID, w.tick+3*TickRate
	dx, dy := target.X-ai.X, target.Y-ai.Y
	distance := math.Hypot(dx, dy)
	d := Direction{AimX: dx / max(1, distance), AimY: dy / max(1, distance)}
	geometry := worldmap.Default().WithSecrets(w.openSecrets).WithVaultGate(w.orb.GateOpen)
	if ai.Combat.Phase == DashIdle && ai.attackCooldown == 0 && distance <= SentinelSwordReach+PlayerSize/2+16 {
		intercept := meleeIntercept(target, SentinelSwordWindup)
		sx, sy := intercept.X-ai.X, intercept.Y-ai.Y
		length := math.Hypot(sx, sy)
		d.AimX, d.AimY = sx/max(1, length), sy/max(1, length)
		ai.meleeHits = nil
		ai.Combat.SwordRemaining = SentinelSwordDuration
		ai.Combat.SwordDirection = Direction{X: d.AimX, Y: d.AimY}
		ai.attackCooldown = int(math.Ceil((SentinelSwordDuration + 1/ai.Stats.AttackRate) * TickRate))
		ai.Behavior = "winding sword"
		d.X, d.Y = d.AimX*parameters.SentinelSwordAdvanceMultiplier, d.AimY*parameters.SentinelSwordAdvanceMultiplier
		return limitMeleeAdvance(ai, target, d, MobMelee(ai.Type).AdvanceMultiplier)
	}
	ai.Behavior = "pursuing"
	goal := meleeIntercept(target, .2)
	if !geometry.CanOccupy(physics.Square(goal.X, goal.Y, AISize)) || !geometry.RoomContains(ai.den.RoomID, physics.Square(goal.X, goal.Y, AISize)) {
		goal = physics.Vector{X: target.X, Y: target.Y}
	}
	movement := ai.navigateMap(geometry, goal.X, goal.Y)
	d.X, d.Y = movement.X, movement.Y
	speed := ai.Stats.MovementSpeed() * SpeedScale * DashSpeedMultiplier
	seconds := min(SentinelDashBurst, max(0, distance-SentinelSwordReach*.6)/speed)
	landing := meleeIntercept(target, DashWindup+seconds)
	lx, ly := landing.X-ai.X, landing.Y-ai.Y
	length := math.Hypot(lx, ly)
	travel := min(speed*SentinelDashBurst, max(0, length-SentinelSwordReach*.6))
	if ai.Combat.Phase == DashIdle && !ai.Combat.Held && ai.Stats.MP >= DashCost && ai.mobChargeCooldown == 0 && distance > SentinelSwordReach+PlayerSize/2 && travel >= speed*.05 && length <= speed*SentinelDashBurst+SentinelSwordReach {
		sx, sy := lx/max(1, length), ly/max(1, length)
		ex, ey := ai.X+sx*travel, ai.Y+sy*travel
		limit, _ := geometry.MovementLimit(ai.X, ai.Y, sx*travel, sy*travel, AISize, AISize)
		if limit >= 1-1e-6 && geometry.RoomContains(ai.den.RoomID, physics.Square(ex, ey, AISize)) {
			d.AimX, d.AimY, d.Dash = sx, sy, true
			ai.sentinelDashHits = nil
			ai.Combat.dashDuration = travel / speed
			ai.mobChargeCooldown = int(math.Ceil((DashWindup + ai.Combat.dashDuration + DashRecovery + .35) * TickRate))
			ai.attackCooldown = ai.mobChargeCooldown
			ai.Behavior = "charging"
		}
	}
	return d
}

func (w *World) sentinelTarget(ai *aiState) *playerState {
	geometry := worldmap.Default()
	var target *playerState
	best := math.Inf(1)
	for _, p := range w.players {
		if p.Dead || !geometry.RoomContains(ai.den.RoomID, physics.Square(p.X, p.Y, PlayerSize)) {
			continue
		}
		distance := math.Hypot(p.X-ai.X, p.Y-ai.Y)
		if target == nil || distance < best || distance == best && p.ID < target.ID {
			target, best = p, distance
		}
	}
	return target
}

func (w *World) sentinelDashContact(ai *aiState, start position, target eater) bool {
	geometry := worldmap.Default().WithSecrets(w.openSecrets).WithVaultGate(w.orb.GateOpen)
	facing := ai.Combat.Direction
	const bladeWidth = 16.0
	root := float64(AISize)/2 + 7
	steps := int(math.Ceil((SentinelSwordReach - root) / (bladeWidth * .75)))
	for i := 0; i <= steps; i++ {
		distance := root + (SentinelSwordReach-root)*float64(i)/float64(steps)
		offsetX := AISize/2 + facing.X*distance - bladeWidth/2
		offsetY := AISize/2 + facing.Y*distance - bladeWidth/2
		from := position{start.x + offsetX, start.y + offsetY}
		to := position{ai.X + offsetX, ai.Y + offsetY}
		fraction, hit := eatingContact(from, to, bladeWidth, target.start, target.end, target.size)
		if !hit {
			continue
		}
		x, y := start.x+(ai.X-start.x)*fraction+AISize/2, start.y+(ai.Y-start.y)*fraction+AISize/2
		tx := target.start.x + (target.end.x-target.start.x)*fraction + target.size/2
		ty := target.start.y + (target.end.y-target.start.y)*fraction + target.size/2
		if geometry.Visible(x, y, tx, ty) {
			return true
		}
	}
	return false
}
