package rendering

import (
	"image/color"

	"raylib-2d-sandbox/internal/rendering/primitives"
	"raylib-2d-sandbox/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func mobTint(t simulation.MobType) color.RGBA {
	if t == simulation.Sentinel {
		return primitives.CharacterColor(simulation.Triangle)
	}
	return rl.Green
}

func mobName(t simulation.MobType) string {
	switch t {
	case simulation.DungeonShooter:
		return "SHOOTER"
	case simulation.DungeonPursuer:
		return "PURSUER"
	case simulation.DungeonHeavy:
		return "HEAVY GUARD"
	case simulation.Sentinel:
		return "SENTINEL"
	}
	return "SPAWN GUARD"
}

func drawMobMark(t simulation.MobType, r rl.Rectangle) {
	center := rl.NewVector2(r.X+r.Width/2, r.Y+r.Height/2)
	size := r.Width * .22
	tint := rl.Fade(rl.RayWhite, .8)
	switch t {
	case simulation.DungeonShooter:
		rl.DrawCircleLinesV(center, size, tint)
		rl.DrawLineEx(rl.NewVector2(center.X-size*1.5, center.Y), rl.NewVector2(center.X+size*1.5, center.Y), 2, tint)
	case simulation.DungeonPursuer:
		for _, offset := range []float32{-size * .6, size * .6} {
			left := rl.NewVector2(center.X-size, center.Y+offset+size*.5)
			tip := rl.NewVector2(center.X, center.Y+offset-size*.5)
			right := rl.NewVector2(center.X+size, left.Y)
			rl.DrawLineEx(left, tip, 2, tint)
			rl.DrawLineEx(tip, right, 2, tint)
		}
	case simulation.DungeonHeavy:
		rl.DrawRectangleLinesEx(r, max(float32(2), r.Width*.08), rl.DarkGreen)
		rl.DrawPoly(center, 4, size, 0, tint)
	case simulation.SpawnGuard:
		rl.DrawCircleV(center, max(float32(2), size*.35), tint)
	}
}
