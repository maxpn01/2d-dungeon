package simulation

import (
	"math"

	"2d-dungeon/internal/parameters"
)

const (
	bossHuntTicks    = parameters.BossHuntTicks
	bossWindupTicks  = parameters.BossWindupTicks
	bossRushTicks    = parameters.BossRushTicks
	bossRecoverTicks = parameters.BossRecoverTicks
)

type bossTarget struct {
	ID   string
	X, Y float64
}

func (b *Boss) chooseTarget(targets []bossTarget) bossTarget {
	for _, target := range targets {
		if target.ID == b.targetID {
			return target
		}
	}
	target := targets[0]
	best := math.Inf(1)
	for _, candidate := range targets {
		distance := math.Hypot(candidate.X-b.X, candidate.Y-b.Y)
		if distance < best {
			best = distance
			target = candidate
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

func (w *World) moveBoss(targets []bossTarget, previous map[string]position) {
	b := &w.boss
	if b.staggerTicks > 0 {
		b.staggerTicks--
		b.Accelerating = false
		return
	}
	b.Accelerating = b.Behavior == "rushing"
	b.Enraged = b.Enraged || b.HP <= b.MaxHP*.5
	target := b.chooseTarget(targets)
	if b.tacticTicks <= 0 {
		switch b.Behavior {
		case "winding":
			b.tactic("rushing", bossRushTicks)
		case "rushing":
			w.finishBossRush(targets, previous)
		case "recovering":
			b.targetID = ""
			target = b.chooseTarget(targets)
			b.Chain = 0
			b.tactic("hunting", bossHuntTicks)
			if !b.Enraged && b.HP < b.MaxHP*.75 && b.attackCount%4 == 0 {
				best := 250.0
				for _, f := range w.fruits {
					distance := math.Hypot(f.X-b.X, f.Y-b.Y)
					if f.Source == "boss" || w.tick < f.AvailableTick || f.Y < MapMin {
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
		b.speed = BaseSpeed * .75
		b.heading = math.Atan2(b.AttackTarget.Y-b.Y, b.AttackTarget.X-b.X)
		return
	case "recovering":
		b.speed = 0
		return
	case "rushing", "feeding":
		tx, ty = b.AttackTarget.X, b.AttackTarget.Y
	}
	dx, dy := tx-b.X, ty-b.Y
	distance := math.Hypot(dx, dy)
	desiredSpeed := BaseSpeed * 1.05
	if b.Behavior == "rushing" {
		desiredSpeed = BaseSpeed * 2.0
		if b.Enraged {
			desiredSpeed = BaseSpeed * 2.35
		}
	}
	if b.Behavior == "feeding" {
		desiredSpeed = BaseSpeed * .95
	}
	b.speed += max(-.9, min(.9, desiredSpeed-b.speed))
	if distance > 0 {
		b.heading = math.Atan2(dy, dx)
		travel := min(distance, b.speed*SpeedScale*StepSeconds)
		b.X = max(MapMin, min(MapMax-BossSize, b.X+dx/distance*travel))
		b.Y = max(MapMin, min(MapMax-BossSize, b.Y+dy/distance*travel))
	}
	if distance <= max(2, b.speed*SpeedScale*StepSeconds) && b.Behavior == "rushing" {
		w.finishBossRush(targets, previous)
	}
	if distance <= 2 && b.Behavior == "feeding" {
		b.tactic("hunting", bossHuntTicks)
	}
}

func (w *World) prepareBossRush(target bossTarget, previous map[string]position) {
	b := &w.boss
	before, ok := previous[target.ID]
	vx, vy := 0.0, 0.0
	if ok {
		vx = (target.X - before.x) / StepSeconds
		vy = (target.Y - before.y) / StepSeconds
	}
	speed := math.Hypot(vx, vy)
	if speed > BaseSpeed*SpeedScale {
		vx *= BaseSpeed * SpeedScale / speed
		vy *= BaseSpeed * SpeedScale / speed
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
		travel = min(travel, (MapMax-BossSize-b.X)/dx)
	} else if dx < 0 {
		travel = min(travel, (MapMin-b.X)/dx)
	}
	if dy > 0 {
		travel = min(travel, (MapMax-BossSize-b.Y)/dy)
	} else if dy < 0 {
		travel = min(travel, (MapMin-b.Y)/dy)
	}
	b.AttackTarget = Point{b.X + dx*max(0, travel), b.Y + dy*max(0, travel)}
	b.tactic("winding", bossWindupTicks)
}

func (w *World) finishBossRush(targets []bossTarget, previous map[string]position) {
	b := &w.boss
	b.attackCount++
	limit := 2
	if b.Enraged {
		limit = 3
	}
	if b.Chain < limit {
		b.Chain++
		oldTarget := b.targetID
		var alternatives []bossTarget
		for _, target := range targets {
			if target.ID != oldTarget {
				alternatives = append(alternatives, target)
			}
		}
		b.targetID = ""
		if len(alternatives) == 0 {
			alternatives = targets
		}
		target := b.chooseTarget(alternatives)
		w.prepareBossRush(target, previous)
	} else {
		b.tactic("recovering", bossRecoverTicks)
	}
}
