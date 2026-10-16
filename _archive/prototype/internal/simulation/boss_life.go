package simulation

import (
	"math"

	"raylib-2d-sandbox/internal/parameters"
)

func (w *World) updateBossLife() {
	b := &w.boss
	old := position{b.X, b.Y}
	best := math.Inf(1)
	found := false
	if b.HP < BossMaxHP || b.Level < MaxLevel {
		for _, f := range w.fruits {
			if f.Source == "boss" || f.X < parameters.BossSpawnMinX || f.X > parameters.BossSpawnMaxX || f.Y < parameters.BossSpawnMinY || f.Y > parameters.BossSpawnMaxY || w.tick < f.AvailableTick || f.ExpiresTick != 0 && w.tick >= f.ExpiresTick {
				continue
			}
			distance := math.Hypot(f.X-b.X, f.Y-b.Y)
			if distance < best {
				best = distance
				found = true
				b.AttackTarget = Point{f.X + (f.Size-BossSize)/2, f.Y + (f.Size-BossSize)/2}
			}
		}
	}
	if found {
		b.Behavior = "feeding"
	} else {
		b.Behavior = "patrolling"
		points := [4]Point{{parameters.BossSpawnMinX, parameters.BossSpawnMinY}, {parameters.BossSpawnMaxX, parameters.BossSpawnMinY}, {parameters.BossSpawnMaxX, parameters.BossSpawnMaxY}, {parameters.BossSpawnMinX, parameters.BossSpawnMaxY}}
		b.AttackTarget = points[b.patrol%len(points)]
		if math.Hypot(b.AttackTarget.X-b.X, b.AttackTarget.Y-b.Y) < 20 {
			b.patrol = (b.patrol + 1) % len(points)
			b.AttackTarget = points[b.patrol]
		}
	}
	b.Accelerating = false
	dx, dy := b.AttackTarget.X-b.X, b.AttackTarget.Y-b.Y
	distance := math.Hypot(dx, dy)
	b.speed += max(-.3, min(.3, BaseSpeed*.65-b.speed))
	if distance > 0 {
		b.heading = math.Atan2(dy, dx)
		travel := min(distance, b.speed*SpeedScale*StepSeconds)
		b.X = max(MapMin, min(MapMax-BossSize, b.X+dx/distance*travel))
		b.Y = max(MapMin, min(MapMax-BossSize, b.Y+dy/distance*travel))
	}
	b.followTail()
	w.feedBoss(old)
}

func (w *World) bossFoodDistance() float64 {
	b := &w.boss
	if b.HP >= BossMaxHP && b.Level >= MaxLevel {
		return math.Inf(1)
	}
	best := math.Inf(1)
	for _, f := range w.fruits {
		if f.Source == "boss" || f.X < parameters.BossSpawnMinX || f.X > parameters.BossSpawnMaxX || f.Y < parameters.BossSpawnMinY || f.Y > parameters.BossSpawnMaxY || w.tick < f.AvailableTick || f.ExpiresTick != 0 && w.tick >= f.ExpiresTick {
			continue
		}
		best = min(best, math.Hypot(f.X-b.X, f.Y-b.Y))
	}
	return best
}

func (w *World) fleeBoss(threat bossTarget) {
	b := &w.boss
	b.tactic("fleeing", 0)
	b.Chain = 0
	b.targetID = ""
	b.speed += max(-.3, min(.3, BaseSpeed*.9-b.speed))
	travel := b.speed * SpeedScale * StepSeconds
	best := -1.0
	bestX, bestY := b.X, b.Y
	for _, direction := range [8]Point{{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}} {
		length := math.Hypot(direction.X, direction.Y)
		x := max(MapMin, min(MapMax-BossSize, b.X+direction.X/length*travel))
		y := max(MapMin, min(MapMax-BossSize, b.Y+direction.Y/length*travel))
		score := math.Hypot(x-threat.X, y-threat.Y)
		if score > best {
			best, bestX, bestY = score, x, y
		}
	}
	b.X, b.Y = bestX, bestY
	b.AttackTarget = Point{bestX, bestY}
}
