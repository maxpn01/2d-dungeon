package simulation

import "math"

// A landed attack sets up a gradual recoil and a visible grace window. Pick a retreat
// direction clear of the entire snake, rather than bouncing into another segment
// or through the arena boundary when fighting against a wall.
func (w *World) bossHitEscape(p *playerState) {
	b := &w.boss
	w.protectedUntil[p.ID] = w.tick + uint64(math.Ceil(BossHitGrace*TickRate))
	p.Combat.HitGrace = BossHitGrace
	p.Combat.Attack = false
	p.Combat.Phase = DashRecoiling
	p.Combat.Remaining = BossRecoilDuration
	b.staggerTicks = 9
	b.Accelerating = false
	angle := math.Atan2(-p.Combat.Direction.Y, -p.Combat.Direction.X)
	bestX, bestY, bestScore := p.X, p.Y, math.Inf(-1)
	for i := 0; i < 16; i++ {
		// Prefer the reverse of the dash; try both sides if the tail blocks it.
		offset := float64((i+1)/2) * math.Pi / 8
		if i%2 == 0 {
			offset = -offset
		}
		x := max(ArenaLeft, min(ArenaRight-PlayerSize, p.X+math.Cos(angle+offset)*BossRecoilDistance))
		y := max(ArenaTop, min(ArenaBottom-PlayerSize, p.Y+math.Sin(angle+offset)*BossRecoilDistance))
		clearance := func(sx, sy float64) float64 {
			dx := max(0, max(sx-(x+PlayerSize), x-(sx+BossSize)))
			dy := max(0, max(sy-(y+PlayerSize), y-(sy+BossSize)))
			return math.Hypot(dx, dy)
		}
		score := clearance(b.X, b.Y)
		for _, segment := range b.Tail {
			score = min(score, clearance(segment.X, segment.Y))
		}
		// Once a candidate clears the snake by 50 units, favor the expected bounce.
		score = min(score, 50) - float64(i)*.5
		if score > bestScore {
			bestX, bestY, bestScore = x, y, score
		}
	}
	p.Combat.RecoilX, p.Combat.RecoilY = bestX-p.X, bestY-p.Y
}
