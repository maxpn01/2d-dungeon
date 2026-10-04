package rendering

import (
	"math"

	"raylib-2d-sandbox/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func drawSentinelSword(center rl.Vector2, size float32, c simulation.CombatState) {
	facing := c.Facing
	if c.SwordRemaining > 0 {
		facing = c.SwordDirection
	}
	angle := math.Atan2(facing.Y, facing.X)
	tint := rl.NewColor(153, 119, 255, 255)
	if c.SwordRemaining > 0 {
		elapsed := simulation.SentinelSwordDuration - c.SwordRemaining
		if elapsed < simulation.SentinelSwordWindup {
			progress := float32(elapsed / simulation.SentinelSwordWindup)
			base := float32(math.Atan2(facing.Y, facing.X) * 180 / math.Pi)
			reach := float32(simulation.SentinelSwordReach)
			warning := rl.Orange
			rl.DrawCircleSector(center, reach, base-60, base+60, 24, rl.Fade(warning, .10+.18*progress))
			rl.DrawCircleSectorLines(center, reach, base-60, base+60, 24, rl.Fade(warning, .5+.5*progress))
			rl.DrawRing(center, reach+5, reach+8, base-60, base-60+120*progress, 24, rl.Gold)
			rl.DrawRing(center, reach*progress, reach*progress+2, base-60, base+60, 24, rl.Fade(rl.Gold, .8))
			angle -= math.Pi / 3 * elapsed / simulation.SentinelSwordWindup
		} else {
			progress := (elapsed - simulation.SentinelSwordWindup) / (simulation.SentinelSwordDuration - simulation.SentinelSwordWindup)
			base := float32(math.Atan2(facing.Y, facing.X) * 180 / math.Pi)
			rl.DrawRing(center, size/2+25, float32(simulation.SentinelSwordReach), base-60, base-60+120*float32(progress), 24, rl.Fade(tint, .35))
			angle += -math.Pi/3 + progress*2*math.Pi/3
		}
	} else {
		angle += math.Pi / 6
	}
	dx, dy := float32(math.Cos(angle)), float32(math.Sin(angle))
	point := func(distance, side float32) rl.Vector2 {
		return rl.NewVector2(center.X+dx*distance-dy*side, center.Y+dy*distance+dx*side)
	}
	root, tip := point(size/2+7, 0), point(float32(simulation.SentinelSwordReach), 0)
	rl.DrawLineEx(root, tip, 12, rl.Fade(tint, .18))
	rl.DrawLineEx(root, tip, 5, tint)
	rl.DrawLineEx(root, tip, 2, rl.RayWhite)
	rl.DrawLineEx(point(size/2+5, -7), point(size/2+5, 7), 3, rl.Gold)
	rl.DrawLineEx(point(size/2-4, 0), root, 4, rl.DarkPurple)
}
