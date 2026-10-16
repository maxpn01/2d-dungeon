package rendering

import (
	"math"

	"raylib-2d-sandbox/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func drawSentinelSword(center rl.Vector2, size float32, c simulation.CombatState) {
	drawMobWeapon(center, size, c, simulation.Sentinel)
}

func drawMobWeapon(center rl.Vector2, size float32, c simulation.CombatState, kind simulation.MobType) {
	if c.StompRemaining > 0 {
		return
	}
	attack := simulation.MobMelee(kind)
	facing := c.Facing
	lunging := kind == simulation.Sentinel && (c.Phase == simulation.DashWinding || c.Phase == simulation.DashRushing || c.Attack)
	if lunging {
		facing = c.Direction
	} else if c.SwordRemaining > 0 {
		facing = c.SwordDirection
	}
	angle := math.Atan2(facing.Y, facing.X)
	tint := rl.NewColor(153, 119, 255, 255)
	accent, grip := rl.Gold, rl.DarkPurple
	warning := rl.Orange
	if kind != simulation.Sentinel {
		tint, accent, grip = rl.Green, rl.Lime, rl.DarkGreen
		warning = rl.Green
	}
	if lunging {
		base := float32(angle * 180 / math.Pi)
		if c.Phase == simulation.DashWinding {
			progress := float32(1 - c.Remaining/simulation.DashWindup)
			rl.DrawCircleSector(center, float32(attack.Reach), base-25, base+25, 16, rl.Fade(warning, .12+.18*progress))
			rl.DrawCircleSectorLines(center, float32(attack.Reach), base-25, base+25, 16, rl.Fade(warning, .5+.5*progress))
		} else {
			rl.DrawRing(center, float32(attack.Reach)-8, float32(attack.Reach)+2, base-25, base+25, 16, rl.Fade(tint, .6))
		}
	} else if c.SwordRemaining > 0 {
		elapsed := attack.Duration - c.SwordRemaining
		if elapsed < attack.Windup {
			progress := float32(elapsed / attack.Windup)
			base := float32(math.Atan2(facing.Y, facing.X) * 180 / math.Pi)
			reach := float32(attack.Reach)
			rl.DrawCircleSector(center, reach, base-60, base+60, 24, rl.Fade(warning, .10+.18*progress))
			rl.DrawCircleSectorLines(center, reach, base-60, base+60, 24, rl.Fade(warning, .5+.5*progress))
			rl.DrawRing(center, reach+5, reach+8, base-60, base-60+120*progress, 24, accent)
			rl.DrawRing(center, reach*progress, reach*progress+2, base-60, base+60, 24, rl.Fade(accent, .8))
			angle -= math.Pi / 3 * elapsed / attack.Windup
		} else {
			progress := min(1, (elapsed-attack.Windup)/attack.Active)
			fade := float32(max(0, 1-max(0, elapsed-attack.Windup-attack.Active)/max(.01, attack.Duration-attack.Windup-attack.Active)))
			base := float32(math.Atan2(facing.Y, facing.X) * 180 / math.Pi)
			rl.DrawRing(center, size/2+25, float32(attack.Reach), base-60, base-60+120*float32(progress), 24, rl.Fade(tint, .35*fade))
			angle += -math.Pi/3 + progress*2*math.Pi/3
		}
	} else {
		angle += math.Pi / 6
	}
	dx, dy := float32(math.Cos(angle)), float32(math.Sin(angle))
	point := func(distance, side float32) rl.Vector2 {
		return rl.NewVector2(center.X+dx*distance-dy*side, center.Y+dy*distance+dx*side)
	}
	root, tip := point(size/2+7, 0), point(float32(attack.Reach), 0)
	rl.DrawLineEx(root, tip, 12, rl.Fade(tint, .18))
	rl.DrawLineEx(root, tip, 5, tint)
	rl.DrawLineEx(root, tip, 2, rl.RayWhite)
	rl.DrawLineEx(point(size/2+5, -7), point(size/2+5, 7), 3, accent)
	rl.DrawLineEx(point(size/2-4, 0), root, 4, grip)
}
