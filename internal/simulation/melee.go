package simulation

import (
	"math"

	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/physics"
)

func meleeIntercept(target *playerState, seconds float64) physics.Vector {
	velocity := physics.Vector{X: target.velocityX, Y: target.velocityY}
	if speed := math.Hypot(velocity.X, velocity.Y); speed > target.MovementSpeed()*parameters.SpeedScale {
		velocity = velocity.Scale(target.MovementSpeed() * parameters.SpeedScale / speed)
	}
	return physics.Vector{X: target.X + velocity.X*seconds*parameters.MeleeAttackPrediction, Y: target.Y + velocity.Y*seconds*parameters.MeleeAttackPrediction}
}

func trackMeleeAim(ai *aiState, target *playerState, impactIn, turnRate float64) {
	intercept := meleeIntercept(target, impactIn)
	dx, dy := intercept.X-ai.X, intercept.Y-ai.Y
	facing := ai.Combat.SwordDirection
	turn := math.Atan2(facing.X*dy-facing.Y*dx, facing.X*dx+facing.Y*dy)
	angle := math.Atan2(facing.Y, facing.X) + max(-turnRate*parameters.StepSeconds, min(turnRate*parameters.StepSeconds, turn))
	ai.Combat.SwordDirection = Direction{X: math.Cos(angle), Y: math.Sin(angle)}
}

func limitMeleeAdvance(ai *aiState, target *playerState, d Direction, speedMultiplier float64) Direction {
	if target == nil || target.Dead {
		return d
	}
	distance := math.Hypot(target.X-ai.X, target.Y-ai.Y)
	travel := ai.Stats.MovementSpeed() * parameters.SpeedScale * parameters.StepSeconds * speedMultiplier
	scale := min(1, max(0, distance-parameters.AISize*.75)/max(1, travel))
	d.X, d.Y = d.X*scale, d.Y*scale
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
				if actor.ai == ai {
					start = actor.start
					break
				}
			}
		} else {
			before := ai.Combat.SwordRemaining
			ai.Combat.SwordRemaining = max(0, before-parameters.StepSeconds)
			impact := attack.Duration - attack.Windup
			if before <= impact-attack.Active || ai.Combat.SwordRemaining > impact {
				continue
			}
		}
		for _, target := range actors {
			if target.dead() || target.ai == ai {
				continue
			}
			if target.player == nil {
				if dashAttack || ai.Type != ForestSnake || ai.meleeHits[target.id] || !w.snakeBiteReaches(ai, target) {
					continue
				}
				if ai.meleeHits == nil {
					ai.meleeHits = make(map[string]bool)
				}
				ai.meleeHits[target.id] = true
				w.provokeHydra(target.ai, ai)
				w.mobStrike(ai, target, attack.DamageMultiplier)
				continue
			}
			if dashAttack {
				if ai.sentinel.dashHits[target.id] || !w.sentinelDashContact(ai, start, target) {
					continue
				}
				if ai.sentinel.dashHits == nil {
					ai.sentinel.dashHits = make(map[string]bool)
				}
				ai.sentinel.dashHits[target.id] = true
			} else {
				if ai.meleeHits[target.id] {
					continue
				}
				dx := target.end.x + target.size/2 - ai.X - parameters.AISize/2
				dy := target.end.y + target.size/2 - ai.Y - parameters.AISize/2
				distance := math.Hypot(dx, dy)
				facing := ai.Combat.SwordDirection
				if distance > attack.Reach+target.size/2 || (distance > 1 && (dx*facing.X+dy*facing.Y)/distance < math.Cos(math.Pi/3)) || !w.aiCanSeePlayer(ai, target.player, attack.Reach+target.size) {
					continue
				}
			}
			if !dashAttack {
				if ai.meleeHits == nil {
					ai.meleeHits = make(map[string]bool)
				}
				ai.meleeHits[target.id] = true
			}
			w.mobStrike(ai, target, attack.DamageMultiplier)
		}
	}
}

func (w *World) mobStrike(ai *aiState, target eater, multiplier float64) {
	if ai.Type == ForestSnake {
		target.status().ApplyPoison(PoisonStrength(ai.Stats, len(ai.Segments)), parameters.SnakePoisonSeconds)
	}
	if target.protected() {
		return
	}
	hp := target.stats.HP
	target.stats.HP = max(0, hp-ai.Stats.Damage*multiplier)
	if target.stats.HP == 0 {
		w.dropKillLoot(target, hp, target.ownerStats().MP)
		ai.Stats.GainXP(float64(target.stats.Level) * parameters.KillXPPerLevel)
		target.die(ai.ID, w.tick)
	}
}
