package simulation

import (
	"math"

	"2d-dungeon/internal/physics"
)

func (ai *aiState) canAttack(target *playerState) bool {
	if ai.settleTicks > 0 || ai.targetID != target.ID || ai.targetTicks < AIReactionTicks || ai.attackCooldown > 0 || ai.Combat.Phase != DashIdle || ai.Combat.Held || ai.Stats.MP < DashCost ||
		CanContactEat(ai.Stats.HP, target.Stats.HP) {
		return false
	}
	hp := ai.Stats.HP
	if CanContactEat(target.Stats.HP, hp) {
		return false
	}
	if hp < target.Stats.HP {
		if target.Stats.HP > ai.Stats.HP && target.Combat.Phase != DashRecovering {
			return false
		}
	}
	if target.Combat.Phase == DashWinding || target.Combat.Phase == DashRushing {
		return false
	}
	distance := math.Hypot(target.X-ai.X, target.Y-ai.Y)
	reach := ai.Stats.MovementSpeed() * SpeedScale * DashSpeedMultiplier * DashBurstDuration(ai.Stats)
	return distance > 0 && distance <= reach*0.7 && ai.clearDash(ai.approach(target.X, target.Y))
}

func (ai *aiState) clearDash(d Direction) bool {
	x, y := ai.X, ai.Y
	distance := ai.Stats.MovementSpeed() * SpeedScale * DashSpeedMultiplier * DashBurstDuration(ai.Stats)
	for distance > 1e-9 {
		step := min(4, distance)
		motion := moveByMotion(x, y, d, step)
		if motion.Blocked() || motion.Rejected {
			return false
		}
		x, y, distance = motion.Bounds.X, motion.Bounds.Y, distance-step
	}
	return true
}

func (w *World) incomingDash(ai *aiState) *playerState {
	var selected *playerState
	best := math.Inf(1)
	for _, p := range w.players {
		if !ai.sees(p.X, p.Y) || p.Dead || (p.Combat.Phase != DashWinding && p.Combat.Phase != DashRushing) {
			continue
		}
		if p.Combat.Phase == DashWinding && p.Combat.Remaining > DashWindup-0.10 {
			continue
		}
		distance := (physics.Vector{X: ai.X, Y: ai.Y}).DistanceSquared(physics.Vector{X: p.X, Y: p.Y})
		if distance > AIDetectionRadius*AIDetectionRadius {
			continue
		}
		burst := DashBurst
		if p.Combat.Phase == DashRushing {
			burst = p.Combat.Remaining
		}
		reach := p.MovementSpeed() * SpeedScale * DashSpeedMultiplier * burst
		end := position{p.X + p.Combat.Direction.X*reach, p.Y + p.Combat.Direction.Y*reach}
		_, hit := eatingContact(position{p.X, p.Y}, end, PlayerSize, position{ai.X, ai.Y}, position{ai.X, ai.Y}, AISize)
		if hit && (selected == nil || distance < best || distance == best && p.ID < selected.ID) {
			selected, best = p, distance
		}
	}
	return selected
}

func (ai *aiState) evade(p *playerState) Direction {
	d := p.Combat.Direction
	best := -math.MaxFloat64
	var chosen Direction
	for _, candidate := range []Direction{{X: -d.Y, Y: d.X}, {X: d.Y, Y: -d.X}} {
		x, y := Move(ai.X, ai.Y, candidate, DashWindup, ai.Stats.MovementSpeed())
		clearance := math.Hypot(x-ai.X, y-ai.Y)
		cross := math.Abs((x-p.X)*d.Y - (y-p.Y)*d.X)
		score := clearance + cross
		if score > best {
			chosen, best = candidate, score
		}
	}
	chosen.Dash = ai.attackCooldown == 0 && ai.Combat.Phase == DashIdle && ai.Stats.MP >= DashCost &&
		CanContactEat(p.Stats.HP, ai.Stats.HP) &&
		p.Combat.Phase == DashWinding && p.Combat.Remaining >= DashWindup-StepSeconds-1e-9 && ai.clearDash(chosen)
	return chosen
}
