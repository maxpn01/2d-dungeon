package simulation

import "math"

// Encounter boundaries do not reset the creature's life, injury or progression.
func (w *World) resumeBossLife() {
	b := &w.boss
	b.Phase = "waiting"
	b.Countdown = 0
	b.Chain = 0
	b.targetID = ""
	b.tactic("feeding", 0)
	w.bossHitUntil = make(map[string]uint64)
	w.openArena()
}

func (w *World) updateBossLife() {
	b := &w.boss
	old := position{b.X, b.Y}
	best := math.Inf(1)
	found := false
	if b.HP < BossMaxHP || b.Level < MaxLevel {
		for _, f := range w.fruits {
			if f.Source == "boss" || !InsideArena(f.X, f.Y, f.Size) || w.tick < f.AvailableTick || f.ExpiresTick != 0 && w.tick >= f.ExpiresTick {
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
		points := [4]Point{{ArenaLeft + 100, ArenaTop + 100}, {ArenaRight - 140, ArenaTop + 100}, {ArenaRight - 140, ArenaBottom - 140}, {ArenaLeft + 100, ArenaBottom - 140}}
		b.AttackTarget = points[b.patrol%len(points)]
		if math.Hypot(b.AttackTarget.X-b.X, b.AttackTarget.Y-b.Y) < 20 {
			b.patrol = (b.patrol + 1) % len(points)
			b.AttackTarget = points[b.patrol]
		}
	}
	b.Accelerating = false
	dx, dy := b.AttackTarget.X-b.X, b.AttackTarget.Y-b.Y
	distance := math.Hypot(dx, dy)
	b.speed += max(-.3, min(.3, PlayerSpeed*.65-b.speed))
	if distance > 0 {
		b.heading = math.Atan2(dy, dx)
		travel := min(distance, b.speed*SpeedScale*StepSeconds)
		b.X = max(ArenaLeft, min(ArenaRight-BossSize, b.X+dx/distance*travel))
		b.Y = max(ArenaTop, min(ArenaBottom-BossSize, b.Y+dy/distance*travel))
	}
	b.followTail()
	w.feedBoss(old)
}
