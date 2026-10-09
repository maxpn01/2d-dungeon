package simulation

import (
	"math"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/physics"
	"raylib-2d-sandbox/internal/worldmap"
)

const (
	SentinelSwordReach    = AISize/2 + 48.0
	SentinelSwordWindup   = 0.2
	SentinelSwordDuration = 0.4
)

func sentinelIntercept(target *playerState, seconds float64) physics.Vector {
	velocity := physics.Vector{X: target.velocityX, Y: target.velocityY}
	if speed := math.Hypot(velocity.X, velocity.Y); speed > target.MovementSpeed()*SpeedScale {
		velocity = velocity.Scale(target.MovementSpeed() * SpeedScale / speed)
	}
	return physics.Vector{X: target.X + velocity.X*seconds*parameters.SentinelAttackPrediction, Y: target.Y + velocity.Y*seconds*parameters.SentinelAttackPrediction}
}

func (w *World) sentinelInput(ai *aiState) Direction {
	ai.attackCooldown = max(0, ai.attackCooldown-1)
	target := w.rangedTarget(ai)
	if ai.Combat.SwordRemaining > 0 {
		ai.Behavior = "striking"
		d := Direction{AimX: ai.Combat.SwordDirection.X, AimY: ai.Combat.SwordDirection.Y}
		if ai.Combat.SwordRemaining > SentinelSwordDuration-SentinelSwordWindup {
			ai.Behavior = "winding sword"
			if target != nil {
				impactIn := ai.Combat.SwordRemaining - (SentinelSwordDuration - SentinelSwordWindup)
				intercept := sentinelIntercept(target, impactIn)
				dx, dy := intercept.X-ai.X, intercept.Y-ai.Y
				if ai.Combat.SwordRemaining > SentinelSwordDuration-parameters.SentinelSwordTrackingDuration {
					facing := ai.Combat.SwordDirection
					turn := math.Atan2(facing.X*dy-facing.Y*dx, facing.X*dx+facing.Y*dy)
					angle := math.Atan2(facing.Y, facing.X) + max(-parameters.SentinelSwordTurnRate*StepSeconds, min(parameters.SentinelSwordTurnRate*StepSeconds, turn))
					ai.Combat.SwordDirection = Direction{X: math.Cos(angle), Y: math.Sin(angle)}
					d.AimX, d.AimY = ai.Combat.SwordDirection.X, ai.Combat.SwordDirection.Y
				}
				if dx*d.AimX+dy*d.AimY > SentinelSwordReach*.5 {
					d.X, d.Y = d.AimX*parameters.SentinelSwordAdvanceMultiplier, d.AimY*parameters.SentinelSwordAdvanceMultiplier
				}
			}
		}
		return d
	}
	if ai.Combat.Phase == DashWinding || ai.Combat.Phase == DashRushing {
		ai.Behavior = "charging"
		return Direction{AimX: ai.Combat.Direction.X, AimY: ai.Combat.Direction.Y}
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
	attacking, flank := w.sentinelFormation(ai, target)
	geometry := worldmap.Default().WithSecrets(w.openSecrets).WithVaultGate(w.orb.GateOpen)
	if attacking && ai.Combat.Phase == DashIdle && ai.attackCooldown == 0 && distance <= SentinelSwordReach+PlayerSize/2+16 {
		intercept := sentinelIntercept(target, SentinelSwordWindup)
		sx, sy := intercept.X-ai.X, intercept.Y-ai.Y
		length := math.Hypot(sx, sy)
		d.AimX, d.AimY = sx/max(1, length), sy/max(1, length)
		ai.Combat.SwordRemaining = SentinelSwordDuration
		ai.Combat.SwordDirection = Direction{X: d.AimX, Y: d.AimY}
		ai.attackCooldown = int(math.Ceil((SentinelSwordDuration + 1/ai.Stats.AttackRate) * TickRate))
		ai.Behavior = "winding sword"
		d.X, d.Y = d.AimX*parameters.SentinelSwordAdvanceMultiplier, d.AimY*parameters.SentinelSwordAdvanceMultiplier
		return d
	}
	goal := flank
	ai.Behavior = "flanking"
	if attacking {
		ai.Behavior = "pursuing"
		goal = sentinelIntercept(target, .2)
	}
	if !geometry.CanOccupy(physics.Square(goal.X, goal.Y, AISize)) {
		goal = physics.Vector{X: target.X, Y: target.Y}
	}
	movement := ai.navigateMap(geometry, goal.X, goal.Y)
	d.X, d.Y = movement.X, movement.Y
	if !attacking {
		return d
	}
	speed := ai.Stats.MovementSpeed() * SpeedScale * DashSpeedMultiplier
	seconds := min(SentinelDashBurst, max(0, distance-SentinelSwordReach*.6)/speed)
	landing := sentinelIntercept(target, DashWindup+seconds)
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

func (w *World) advanceMobMelee(actors []eater) {
	for i := range w.ais {
		ai := &w.ais[i]
		attack := MobMelee(ai.Type)
		dashAttack := ai.Type == Sentinel && ai.Combat.Attack && ai.Combat.Phase != DashBumped
		if attack.Duration == 0 || ai.respawnTick != 0 || ai.Stats.HP <= 0 || !dashAttack && ai.Combat.SwordRemaining <= 0 {
			continue
		}
		start := position{ai.X, ai.Y}
		if dashAttack {
			for _, actor := range actors {
				if actor.ai == ai && actor.cell == nil {
					start = actor.start
					break
				}
			}
		} else {
			before := ai.Combat.SwordRemaining
			ai.Combat.SwordRemaining = max(0, before-StepSeconds)
			impact := attack.Duration - attack.Windup
			if before <= impact || ai.Combat.SwordRemaining > impact {
				continue
			}
		}
		for _, target := range actors {
			if target.player == nil || target.dead() || target.protectedAt(1, w.safeZones) {
				continue
			}
			if dashAttack {
				if ai.sentinelDashHits[target.id] || !w.sentinelDashContact(ai, start, target) {
					continue
				}
				if ai.sentinelDashHits == nil {
					ai.sentinelDashHits = make(map[string]bool)
				}
				ai.sentinelDashHits[target.id] = true
			} else {
				dx := target.end.x + target.size/2 - ai.X - AISize/2
				dy := target.end.y + target.size/2 - ai.Y - AISize/2
				distance := math.Hypot(dx, dy)
				facing := ai.Combat.SwordDirection
				if distance > attack.Reach+target.size/2 || (dx*facing.X+dy*facing.Y)/max(1, distance) < math.Cos(math.Pi/3) || !w.aiCanSeePlayer(ai, target.player, attack.Reach+target.size) {
					continue
				}
			}
			hp := target.stats.HP
			target.stats.HP = max(0, hp-ai.Stats.Damage*attack.DamageMultiplier)
			if target.stats.HP == 0 {
				w.dropKillLoot(target, hp, target.ownerStats().MP)
				ai.Stats.GainXP(float64(target.stats.Level) * parameters.KillXPPerLevel)
				target.die(ai.ID, w.tick)
			}
		}
	}
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
