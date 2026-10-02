package worldmap

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/rendering/primitives"
	"raylib-2d-sandbox/internal/simulation"
)

func drawOrbPedestal(orb simulation.OrbState, local protocol.Player, exploration *Exploration) {
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
	if orb.Carrier == "" {
		tint = primitives.RuneColor
		if !orb.Unlocked {
			tint = rl.Orange
		}
		orbCenter := rl.NewVector2(center.X, center.Y-38+float32(math.Sin(rl.GetTime()*2))*4)
		rl.DrawCircleGradient(orbCenter, 40, rl.Fade(tint, .3), rl.Blank)
		rl.DrawCircleV(orbCenter, 12, tint)
		rl.DrawCircleV(rl.NewVector2(orbCenter.X-3, orbCenter.Y-4), 4, rl.RayWhite)
		label = fmt.Sprintf("SEALED / %d SENTINELS REMAIN", orb.GuardsRemaining)
		if orb.Unlocked {
			label = "[E] TAKE THE ORB"
		}
	}
	if !local.Dead && math.Hypot(local.X+local.Size/2-orb.X, local.Y+local.Size/2-orb.Y) <= 100 {
		primitives.DrawTextAt(label, rl.NewVector2(center.X-float32(primitives.MeasureText(label, 16))/2, center.Y-85), 16, tint)
	}
}
