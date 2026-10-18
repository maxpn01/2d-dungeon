package worldmap

import (
	"fmt"
	"math"

	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/rendering/primitives"
	"2d-dungeon/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func drawOrbPedestal(orb simulation.OrbState, local simulation.Player, exploration *Exploration) {
	if !exploration.EntityVisible(orb.X, orb.Y, 0) {
		return
	}
	center := rl.NewVector2(float32(orb.X), float32(orb.Y))
	rl.DrawEllipse(int32(orb.X), int32(orb.Y+18), 38, 15, rl.Fade(rl.Black, .6))
	rl.DrawPoly(center, 6, 30, 30, stone)
	rl.DrawPolyLines(center, 6, 30, 30, paleStone)
	rl.DrawRectangle(int32(orb.X-14), int32(orb.Y-19), 28, 32, stone)
	rl.DrawEllipse(int32(orb.X), int32(orb.Y-19), 25, 10, paleStone)
	label := "ORB TAKEN"
	tint := rl.Gray
	if orb.Consumed {
		label = "ORB CONSUMED"
	} else if orb.Carrier == "" {
		tint = primitives.RuneColor
		if !orb.Unlocked {
			tint = rl.NewColor(184, 140, 255, 255)
		}
		orbCenter := rl.NewVector2(center.X, center.Y-38+float32(math.Sin(rl.GetTime()*2))*4)
		rl.DrawCircleGradient(orbCenter, 40, rl.Fade(tint, .3), rl.Blank)
		rl.DrawCircleV(orbCenter, 12, tint)
		rl.DrawCircleV(rl.NewVector2(orbCenter.X-3, orbCenter.Y-4), 4, rl.RayWhite)
		for i := 0; i < 3; i++ {
			angle := rl.GetTime()*.7 + float64(i)*math.Pi*2/3
			pos := rl.NewVector2(orbCenter.X+float32(math.Cos(angle))*22, orbCenter.Y+float32(math.Sin(angle))*9)
			rl.DrawCircleV(pos, 2, rl.Gold)
		}
		rl.DrawEllipseLines(int32(orbCenter.X), int32(orbCenter.Y), 24, 10, rl.Fade(tint, .65))
		label = fmt.Sprintf("SEALED / %d OF 4 SEALS LIT", 4-orb.GuardsRemaining)
		if orb.Unlocked {
			label = "[E] TAKE THE ORB"
		}
	}
	if !local.Dead && math.Hypot(local.X+parameters.PlayerSize/2-orb.X, local.Y+parameters.PlayerSize/2-orb.Y) <= 100 {
		primitives.DrawTextAt(label, rl.NewVector2(center.X-float32(primitives.MeasureText(label, 16))/2, center.Y-85), 16, tint)
	}
}
