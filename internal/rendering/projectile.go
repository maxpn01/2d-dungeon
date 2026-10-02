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
			start.X, start.Y = worldmap.Default().Move(shot.X, shot.Y, -shot.Direction.X*parameters.ShotSpeed*simulation.StepSeconds, -shot.Direction.Y*parameters.ShotSpeed*simulation.StepSeconds, shot.Size, shot.Size)
		}
		shot.X = start.X + (shot.X-start.X)*alpha
		shot.Y = start.Y + (shot.Y-start.Y)*alpha
		if !exploration.EntityVisible(shot.X, shot.Y, shot.Size) {
			continue
		}
		center := rl.NewVector2(float32(shot.X+shot.Size/2), float32(shot.Y+shot.Size/2))
		drawMagicBall(center, shot)
	}
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
