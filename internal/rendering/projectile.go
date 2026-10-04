package rendering

import (
	"math"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/simulation"
	"raylib-2d-sandbox/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func DrawProjectiles(previous, shots []simulation.Projectile, alpha float64, exploration Visibility) {
	alpha = max(0, min(1, alpha))
	starts := make(map[string]simulation.Projectile, len(previous))
	for _, shot := range previous {
		starts[shot.ID] = shot
	}
	for _, shot := range shots {
		start, exists := starts[shot.ID]
		if !exists {
			start = shot
			start.X, start.Y = worldmap.Default().Move(shot.X, shot.Y, -shot.Direction.X*simulation.ProjectileSpeed(shot)*simulation.StepSeconds, -shot.Direction.Y*simulation.ProjectileSpeed(shot)*simulation.StepSeconds, shot.Size, shot.Size)
		}
		shot.X = start.X + (shot.X-start.X)*alpha
		shot.Y = start.Y + (shot.Y-start.Y)*alpha
		if !exploration.EntityVisible(shot.X, shot.Y, shot.Size) {
			continue
		}
		center := rl.NewVector2(float32(shot.X+shot.Size/2), float32(shot.Y+shot.Size/2))
		if shot.Kind == "arrow" {
			drawVaultArrow(center, shot)
			continue
		}
		drawMagicBall(center, shot)
	}
}

func drawVaultArrow(center rl.Vector2, shot simulation.Projectile) {
	dx, dy := float32(shot.Direction.X), float32(shot.Direction.Y)
	point := func(distance, side float32) rl.Vector2 {
		return rl.NewVector2(center.X+dx*distance-dy*side, center.Y+dy*distance+dx*side)
	}
	rl.DrawLineEx(point(-24, 0), point(4, 0), 7, rl.Fade(rl.Orange, .15))
	rl.DrawLineEx(point(-20, 0), point(4, 0), 3, rl.NewColor(159, 123, 78, 255))
	rl.DrawLineEx(point(-18, -1), point(4, -1), 1, rl.NewColor(219, 198, 149, 255))
	rl.DrawLineEx(point(7, 0), point(-2, -5), 2, rl.RayWhite)
	rl.DrawLineEx(point(7, 0), point(-2, 5), 2, rl.RayWhite)
	rl.DrawLineEx(point(-18, 0), point(-24, -5), 2, rl.NewColor(185, 49, 52, 255))
	rl.DrawLineEx(point(-18, 0), point(-24, 5), 2, rl.NewColor(185, 49, 52, 255))
}

func drawMagicBall(center rl.Vector2, shot simulation.Projectile) {
	radius := float32(shot.Size * .65)
	age := max(0, parameters.ShotLifetime-shot.Remaining)
	tailLength := float32(min(64, age*parameters.ShotSpeed))
	direction := rl.NewVector2(float32(shot.Direction.X), float32(shot.Direction.Y))
	pulse := float32(1 + .08*math.Sin(rl.GetTime()*9))
	rl.BeginBlendMode(rl.BlendAdditive)
	for i := 12; i >= 1; i-- {
		fraction := float32(i) / 12
		fade := (1 - fraction) * (1 - fraction)
		point := rl.NewVector2(center.X-direction.X*tailLength*fraction, center.Y-direction.Y*tailLength*fraction)
		width := radius * (1 - .8*fraction)
		rl.DrawCircleGradient(point, width*1.9, rl.Fade(rl.Violet, .45*fade), rl.Fade(rl.Violet, 0))
		rl.DrawCircleV(point, width, rl.Fade(rl.Purple, .7*fade))
		rl.DrawCircleV(point, width*.5, rl.Fade(rl.RayWhite, .65*fade))
	}
	rl.DrawCircleGradient(center, radius*2.9*pulse, rl.Fade(rl.Violet, .5), rl.Fade(rl.Violet, 0))
	rl.DrawCircleGradient(center, radius*1.8*pulse, rl.Fade(rl.Violet, .85), rl.Fade(rl.Purple, 0))
	rl.EndBlendMode()
	rl.DrawCircleV(center, radius, rl.Violet)
	rl.DrawCircleGradient(center, radius, rl.RayWhite, rl.Violet)
	rl.DrawCircleV(center, radius*.48, rl.RayWhite)
}
