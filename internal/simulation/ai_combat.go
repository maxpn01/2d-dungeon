package simulation

import "math"

func (ai *aiState) canAttack(target *playerState) bool {
	if ai.Combat.Phase != DashIdle || ai.Combat.Held || ai.Stats.HP <= DashCost ||
		CanContactEat(ai.Stats.HP, target.Stats.HP) {
		return false
	}
	hp := ai.Stats.HP - DashCost
	if CanContactEat(target.Stats.HP, hp) {
		return false
	}
	if hp < target.Stats.HP {
		if target.Stats.HP <= DashCost || (target.Stats.HP > ai.Stats.HP && target.Combat.Phase != DashRecovering) {
			return false
		}
	}
	if target.Combat.Phase == DashWinding || target.Combat.Phase == DashRushing {
		return false
	}
	distance := math.Hypot(target.X-ai.X, target.Y-ai.Y)
	reach := ai.Stats.Speed * SpeedScale * DashSpeedMultiplier * DashBurst
	return distance > 0 && distance <= reach*0.7 && ai.clearDash(ai.approach(target.X, target.Y))
}

func (ai *aiState) clearDash(d Direction) bool {
	x, y := ai.X, ai.Y
	distance := ai.Stats.Speed * SpeedScale * DashSpeedMultiplier * DashBurst
	for distance > 1e-9 {
		step := min(4, distance)
		nx, ny := MoveBy(x, y, d, step)
		if math.Abs(nx-x-d.X*step) > 1e-6 || math.Abs(ny-y-d.Y*step) > 1e-6 {
			return false
		}
		x, y, distance = nx, ny, distance-step
	}
	return true
}

func (w *World) incomingDash(ai *aiState) *playerState {
	var selected *playerState
	best := math.Inf(1)
	for _, p := range w.players {
		if p.Dead || (p.Combat.Phase != DashWinding && p.Combat.Phase != DashRushing) {
			continue
		}
		dx, dy := ai.X-p.X, ai.Y-p.Y
		distance := dx*dx + dy*dy
		if distance > AIDetectionRadius*AIDetectionRadius {
			continue
		}
		burst := DashBurst
		if p.Combat.Phase == DashRushing {
			burst = p.Combat.Remaining
		}
		reach := p.Stats.Speed * SpeedScale * DashSpeedMultiplier * burst
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
		x, y := Move(ai.X, ai.Y, candidate, DashWindup, ai.Stats.Speed)
		clearance := math.Hypot(x-ai.X, y-ai.Y)
		cross := math.Abs((x-p.X)*d.Y - (y-p.Y)*d.X)
		score := clearance + cross
		if score > best {
			chosen, best = candidate, score
		}
	}
	chosen.Dash = ai.Combat.Phase == DashIdle && ai.Stats.HP > DashCost &&
		CanContactEat(p.Stats.HP, ai.Stats.HP-DashCost) &&
		p.Combat.Phase == DashWinding && p.Combat.Remaining >= DashWindup-StepSeconds-1e-9 && ai.clearDash(chosen)
	return chosen
}
