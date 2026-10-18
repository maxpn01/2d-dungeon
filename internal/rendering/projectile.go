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
			if shot.Kind == "shockwave" {
				start.X -= shot.Direction.X * simulation.ProjectileSpeed(shot) * simulation.StepSeconds
				start.Y -= shot.Direction.Y * simulation.ProjectileSpeed(shot) * simulation.StepSeconds
			} else {
				start.X, start.Y = worldmap.Default().Move(shot.X, shot.Y, -shot.Direction.X*simulation.ProjectileSpeed(shot)*simulation.StepSeconds, -shot.Direction.Y*simulation.ProjectileSpeed(shot)*simulation.StepSeconds, shot.Size, shot.Size)
			}
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
		if shot.Kind == "shockwave" {
			if exists {
				shot.Remaining = start.Remaining + (shot.Remaining-start.Remaining)*alpha
			}
			drawStompShockwave(center, shot)
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
	tint, trailTint := rl.Violet, rl.Purple
	if !shot.MobOwned {
		tint = rl.NewColor(255, 128, 77, 255)
		trailTint = rl.NewColor(204, 75, 40, 255)
	}
	if shot.Kind == "spectral" {
		tint, trailTint = rl.NewColor(192, 211, 255, 255), rl.NewColor(132, 105, 219, 255)
	}
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
		rl.DrawCircleGradient(point, width*1.9, rl.Fade(tint, .45*fade), rl.Fade(tint, 0))
		rl.DrawCircleV(point, width, rl.Fade(trailTint, .7*fade))
		rl.DrawCircleV(point, width*.5, rl.Fade(rl.RayWhite, .65*fade))
	}
	rl.DrawCircleGradient(center, radius*2.9*pulse, rl.Fade(tint, .5), rl.Fade(tint, 0))
	rl.DrawCircleGradient(center, radius*1.8*pulse, rl.Fade(tint, .85), rl.Fade(trailTint, 0))
	rl.EndBlendMode()
	rl.DrawCircleV(center, radius, tint)
	rl.DrawCircleGradient(center, radius, rl.RayWhite, tint)
	rl.DrawCircleV(center, radius*.48, rl.RayWhite)
}

func drawStompShockwave(center rl.Vector2, shot simulation.Projectile) {
	direction := rl.NewVector2(float32(shot.Direction.X), float32(shot.Direction.Y))
	age := max(0, shot.Lifetime-shot.Remaining)
	fade := float32(min(1, shot.Remaining/.15))
	base := float32(math.Atan2(shot.Direction.Y, shot.Direction.X) * 180 / math.Pi)
	radius := float32(shot.Size * .65)
	root := rl.NewVector2(center.X-direction.X*radius*.7, center.Y-direction.Y*radius*.7)
	for i := 3; i >= 0; i-- {
		point := rl.NewVector2(root.X-direction.X*float32(i)*9, root.Y-direction.Y*float32(i)*9)
		rl.DrawRing(point, radius-4, radius+3, base-55, base+55, 16, rl.Fade(rl.Orange, fade*(1-float32(i)*.22)))
	}
	rl.DrawRing(root, radius-1, radius+1, base-55, base+55, 16, rl.Fade(rl.RayWhite, fade))
	for i := 0; i < 9; i++ {
		angle := float64(base-55+float32(i)*110/8) * math.Pi / 180
		dx, dy := float32(math.Cos(angle)), float32(math.Sin(angle))
		inner := rl.NewVector2(root.X+dx*(radius+3), root.Y+dy*(radius+3))
		outer := rl.NewVector2(root.X+dx*(radius+7+float32(i%3)*3), root.Y+dy*(radius+7+float32(i%3)*3))
		rl.DrawLineEx(inner, outer, 2, rl.Fade(rl.Gold, fade*.8))
	}
	for i := 0; i < 28; i++ {
		seed := float64(i) * 2.399963
		particleAge := math.Mod(age+float64(i)*.011, .28)
		if particleAge > age {
			continue
		}
		trail := float32(particleAge * parameters.HeavyStompWaveSpeed)
		side := float32(math.Sin(seed)) * (float32(shot.Size)*.4 + float32(particleAge)*24)
		point := rl.NewVector2(center.X-direction.X*trail-direction.Y*side, center.Y-direction.Y*trail+direction.X*side)
		alpha := fade * float32(1-particleAge/.28)
		rl.DrawCircleV(point, 5+float32(i%4), rl.Fade(rl.NewColor(160, 131, 94, 255), alpha*.38))
		lift := float32(math.Sin(particleAge/.28*math.Pi)) * float32(8+i%9)
		point.Y -= lift
		rockSize := float32(2 + i%4)
		rl.DrawRectanglePro(rl.NewRectangle(point.X, point.Y, rockSize, rockSize), rl.NewVector2(rockSize/2, rockSize/2), float32(seed*45+particleAge*320), rl.Fade(rl.NewColor(181, 145, 96, 255), alpha))
	}
}

func drawStompDebris(center, direction rl.Vector2, age float64, fade float32) {
	for i := 0; i < 22; i++ {
		seed := float64(i) * 2.399963
		distance := float32(age) * float32(110+i%5*30)
		side := float32(math.Sin(seed)) * (5 + float32(age)*35)
		point := rl.NewVector2(center.X+direction.X*distance-direction.Y*side, center.Y+direction.Y*distance+direction.X*side)
		rockSize := float32(2 + i%4)
		rl.DrawCircleV(point, rockSize+2, rl.Fade(rl.NewColor(160, 131, 94, 255), fade*.2))
		point.Y -= float32(math.Sin(min(1, age/.45)*math.Pi)) * float32(12+i%4*6)
		rl.DrawRectanglePro(rl.NewRectangle(point.X, point.Y, rockSize, rockSize), rl.NewVector2(rockSize/2, rockSize/2), float32(seed*45+age*240), rl.Fade(rl.NewColor(181, 145, 96, 255), fade))
	}
}
