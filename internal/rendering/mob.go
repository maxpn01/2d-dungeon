package rendering

import (
	"image/color"

	"2d-dungeon/internal/rendering/primitives"
	"2d-dungeon/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func mobTint(t simulation.MobType) color.RGBA {
	if t == simulation.Sentinel {
		return primitives.CharacterColor(simulation.Triangle)
	}
	if t == simulation.GateHeavy {
		return rl.NewColor(93, 138, 111, 255)
	}
	return rl.Green
}

func mobName(t simulation.MobType) string {
	switch t {
	case simulation.DungeonShooter:
		return "SHOOTER"
	case simulation.DungeonPursuer:
		return "PURSUER"
	case simulation.GateHeavy:
		return "GATE WARDEN"
	case simulation.DungeonHeavy:
		return "HEAVY GUARD"
	case simulation.Sentinel:
		return "SENTINEL"
	}
	return "SPAWN GUARD"
}

func drawMobMark(t simulation.MobType, r rl.Rectangle) {
	center := rl.NewVector2(r.X+r.Width/2, r.Y+r.Height/2)
	size := min(r.Width, r.Height) * .26
	tint := rl.NewColor(214, 178, 255, 255)
	outline := rl.NewColor(65, 30, 96, 255)
	stroke := max(float32(2), size*.3)
	border := max(float32(1), size*.15)
	drawSymbol := func(width float32, ink color.RGBA) {
		switch t {
		case simulation.DungeonShooter:
			rl.DrawRing(center, max(float32(0), size-width/2), size+width/2, 0, 360, 32, ink)
			rl.DrawLineEx(rl.NewVector2(center.X-size*1.4, center.Y), rl.NewVector2(center.X+size*1.4, center.Y), width, ink)
			rl.DrawLineEx(rl.NewVector2(center.X, center.Y-size*1.4), rl.NewVector2(center.X, center.Y+size*1.4), width, ink)
		case simulation.DungeonPursuer:
			for _, offset := range []float32{-size * .6, size * .6} {
				left := rl.NewVector2(center.X-size, center.Y+offset+size*.5)
				tip := rl.NewVector2(center.X, center.Y+offset-size*.5)
				right := rl.NewVector2(center.X+size, left.Y)
				rl.DrawLineEx(left, tip, width, ink)
				rl.DrawLineEx(tip, right, width, ink)
			}
		}
	}
	switch t {
	case simulation.DungeonShooter, simulation.DungeonPursuer:
		drawSymbol(stroke+border*2, outline)
		drawSymbol(stroke, tint)
	case simulation.DungeonHeavy, simulation.GateHeavy:
		rl.DrawRectangleLinesEx(r, max(float32(2), r.Width*.08), rl.DarkGreen)
		block := rl.NewRectangle(center.X-size*.8, center.Y-size*.8, size*1.6, size*1.6)
		rl.DrawRectangleRec(rl.NewRectangle(block.X-border, block.Y-border, block.Width+border*2, block.Height+border*2), outline)
		rl.DrawRectangleRec(block, tint)
		if t == simulation.GateHeavy {
			gold := rl.NewColor(223, 186, 106, 255)
			rl.DrawRectangleLinesEx(r, max(float32(2), r.Width*.07), gold)
			rl.DrawPoly(center, 4, size, 0, gold)
			rl.DrawPoly(center, 4, size*.55, 0, outline)
			for _, offset := range []float32{-size * .8, 0, size * .8} {
				rl.DrawPoly(rl.NewVector2(center.X+offset, r.Y+size*.35), 3, size*.26, 0, gold)
			}
		}
	case simulation.SpawnGuard:
		radius := max(float32(2), size*.35)
		rl.DrawCircleV(center, radius+border, outline)
		rl.DrawCircleV(center, radius, tint)
	}
}
