package simulation

import (
	"math"

	"raylib-2d-sandbox/internal/parameters"
)

const (
	bossHuntTicks    = parameters.BossHuntTicks
	bossWindupTicks  = parameters.BossWindupTicks
	bossRushTicks    = parameters.BossRushTicks
	bossRecoverTicks = parameters.BossRecoverTicks
)

// Keep one target for the attack cycle instead of switching between nearby
// players and food every frame. Retarget immediately if that player leaves.
func (b *Boss) chooseTarget(players []*playerState) *playerState {
	for _, p := range players {
		if p.ID == b.targetID {
			return p
		}
	}
	target := players[0]
	best := math.Inf(1)
	for _, p := range players {
		distance := math.Hypot(p.X-b.X, p.Y-b.Y)
		if distance < best {
			best = distance
			target = p
		}
	}
	b.targetID = target.ID
	return target
}

func (b *Boss) tactic(state string, ticks int) {
	b.Behavior = state
	b.tacticTicks = ticks
	b.Accelerating = state == "rushing"
}

func (w *World) moveBoss(players []*playerState, previous map[string]position) {
	b := &w.boss
	if b.staggerTicks > 0 {
		b.staggerTicks--
		b.Accelerating = false
		return
	}
	b.Accelerating = b.Behavior == "rushing"
	b.Enraged = b.Enraged || b.HP <= b.MaxHP*.5
	target := b.chooseTarget(players)
	if b.tacticTicks <= 0 {
		switch b.Behavior {
		case "winding":
			b.tactic("rushing", bossRushTicks)
		case "rushing":
			w.finishBossRush(players, previous)
		case "recovering":
			b.targetID = ""
			target = b.chooseTarget(players)
			b.Chain = 0
			b.tactic("hunting", bossHuntTicks)
			// Occasionally take a short feeding opportunity when wounded. It never
			// interrupts a telegraphed attack, and cannot become an endless food loop.
			if !b.Enraged && b.HP < b.MaxHP*.75 && b.attackCount%4 == 0 {
				best := 250.0
				for _, f := range w.fruits {
					distance := math.Hypot(f.X-b.X, f.Y-b.Y)
					if f.Source == "boss" || w.tick < f.AvailableTick || !InsideArena(f.X, f.Y, f.Size) {
						continue
					}
					if distance < best {
						best = distance
						b.AttackTarget = Point{f.X, f.Y}
						b.tactic("feeding", 15)
					}
				}
			}
		default:
			b.tactic("hunting", bossHuntTicks)
		}
	}
	b.tacticTicks--
	tx, ty := target.X+(PlayerSize-BossSize)/2, target.Y+(PlayerSize-BossSize)/2
	if b.Behavior == "hunting" && (b.tacticTicks == 0 || math.Hypot(tx-b.X, ty-b.Y) < 420) {
		b.Chain = 1
		w.prepareBossRush(target, previous)

	}
	switch b.Behavior {
	case "winding":
		b.speed = PlayerSpeed * .75
		b.heading = math.Atan2(b.AttackTarget.Y-b.Y, b.AttackTarget.X-b.X)
		return
	case "recovering":
		// A missed rush leaves a predictable punish window.
		b.speed = 0
		return
	case "rushing", "feeding":
		tx, ty = b.AttackTarget.X, b.AttackTarget.Y
	}
	dx, dy := tx-b.X, ty-b.Y
	distance := math.Hypot(dx, dy)
	desiredSpeed := PlayerSpeed * 1.05
	if b.Behavior == "rushing" {
		desiredSpeed = PlayerSpeed * 2.0
		if b.Enraged {
			desiredSpeed = PlayerSpeed * 2.35
		}
	}
	if b.Behavior == "feeding" {
		desiredSpeed = PlayerSpeed * .95
	}
	b.speed += max(-.9, min(.9, desiredSpeed-b.speed))
	// Point at the destination and stop at it. The old fixed turning radius and
	// constant forward motion made the head orbit targets it could not turn into.
	if distance > 0 {
		b.heading = math.Atan2(dy, dx)
		travel := min(distance, b.speed*SpeedScale*StepSeconds)
		b.X = max(ArenaLeft, min(ArenaRight-BossSize, b.X+dx/distance*travel))
		b.Y = max(ArenaTop, min(ArenaBottom-BossSize, b.Y+dy/distance*travel))
	}
	if distance <= max(2, b.speed*SpeedScale*StepSeconds) && b.Behavior == "rushing" {
		w.finishBossRush(players, previous)
	}
	if distance <= 2 && b.Behavior == "feeding" {
		b.tactic("hunting", bossHuntTicks)
	}
}

// Rush through the intercept instead of braking at the player's old position.
// Ray clipping preserves the attack line even when its end meets an arena wall.
func (w *World) prepareBossRush(target *playerState, previous map[string]position) {
	b := &w.boss
	before, ok := previous[target.ID]
	vx, vy := 0.0, 0.0
	if ok {
		vx = (target.X - before.x) / StepSeconds
		vy = (target.Y - before.y) / StepSeconds
	}
	speed := math.Hypot(vx, vy)
	if speed > PlayerSpeed*SpeedScale {
		vx *= PlayerSpeed * SpeedScale / speed
		vy *= PlayerSpeed * SpeedScale / speed
	}
	tx := target.X + (PlayerSize-BossSize)/2 + vx*.28
	ty := target.Y + (PlayerSize-BossSize)/2 + vy*.28
	dx, dy := tx-b.X, ty-b.Y
	distance := math.Hypot(dx, dy)
	if distance < 1 {
		dx, dy, distance = math.Cos(b.heading), math.Sin(b.heading), 1
	}
	dx, dy = dx/distance, dy/distance
	travel := distance + 200
	if dx > 0 {
		travel = min(travel, (ArenaRight-BossSize-b.X)/dx)
	} else if dx < 0 {
		travel = min(travel, (ArenaLeft-b.X)/dx)
	}
	if dy > 0 {
		travel = min(travel, (ArenaBottom-BossSize-b.Y)/dy)
	} else if dy < 0 {
		travel = min(travel, (ArenaTop-b.Y)/dy)
	}
	b.AttackTarget = Point{b.X + dx*max(0, travel), b.Y + dy*max(0, travel)}
	b.tactic("winding", bossWindupTicks)
}

func (w *World) finishBossRush(players []*playerState, previous map[string]position) {
	b := &w.boss
	b.attackCount++
	limit := 2
	if b.Enraged {
		limit = 3
	}
	if b.Chain < limit {
		b.Chain++
		// With friends, the follow-up threatens another player rather than always
		// dogpiling the same one. The warning reveals the new committed path.
		oldTarget := b.targetID
		var alternatives []*playerState
		for _, p := range players {
			if p.ID != oldTarget {
				alternatives = append(alternatives, p)
			}
		}
		b.targetID = ""
		if len(alternatives) == 0 {
			alternatives = players
		}
		target := b.chooseTarget(alternatives)
		w.prepareBossRush(target, previous)
	} else {
		b.tactic("recovering", bossRecoverTicks)
	}
}
