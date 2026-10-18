package ui

import (
	"fmt"
	"image/color"
	"math"

	"dungeon-raylib-go/internal/parameters"
	"dungeon-raylib-go/internal/rendering"
	"dungeon-raylib-go/internal/rendering/primitives"
	"dungeon-raylib-go/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func DrawBossBar(ai simulation.AI, width int32) {
	h := ai.Hydra
	panel := min(int32(560), width-80)
	if panel <= 200 {
		return
	}
	x, y := (width-panel)/2, int32(104)
	rl.DrawRectangle(x, y, panel, 86, rl.Fade(rl.Black, .82))
	title := fmt.Sprintf("CAVE HYDRA  LV %d", ai.Stats.Level)
	primitives.DrawText(title, x+14, y+10, 20, rl.NewColor(214, 176, 255, 255))
	state := map[simulation.HydraPhase]string{
		simulation.HydraSleeping: "ASLEEP", simulation.HydraForaging: "FORAGING",
		simulation.HydraHunting: "HUNTING", simulation.HydraReturning: "RETURNING",
	}[h.Phase]
	if h.Engaged {
		state = "ENRAGED"
	}
	primitives.DrawText(state, x+panel-14-primitives.MeasureText(state, 16), y+13, 16, rl.LightGray)
	track := rl.NewRectangle(float32(x+14), float32(y+38), float32(panel-28), 14)
	rl.DrawRectangleRec(track, color.RGBA{40, 32, 44, 255})
	fill := track
	fill.Width *= float32(max(0, min(1, ai.Stats.HP/ai.Stats.MaxHP())))
	tint := color.RGBA{178, 44, 84, 255}
	if ai.Status.Poison.Active() {
		tint = rendering.PoisonColor
	}
	rl.DrawRectangleRec(fill, tint)
	rl.DrawRectangleLinesEx(track, 1, rl.Fade(rl.RayWhite, .3))
	hp := fmt.Sprintf("%.0f / %.0f", ai.Stats.HP, ai.Stats.MaxHP())
	primitives.DrawText(hp, int32(track.X+track.Width/2)-primitives.MeasureText(hp, 12)/2, y+39, 12, rl.RayWhite)
	slot := (panel - 28 - 16) / int32(len(h.Heads))
	for i, head := range h.Heads {
		hx := x + 14 + int32(i)*(slot+8)
		bar := rl.NewRectangle(float32(hx), float32(y+62), float32(slot), 12)
		rl.DrawRectangleRec(bar, color.RGBA{40, 32, 44, 255})
		label, ink := "", rl.LightGray
		switch {
		case head.Severed && head.Growing > 0:
			grow := bar
			grow.Width *= float32(head.Growing)
			rl.DrawRectangleRec(grow, rl.NewColor(150, 236, 120, 255))
			label, ink = "REGROWING", rl.NewColor(150, 236, 120, 255)
		case head.Severed && head.Regrow > 0:
			label, ink = fmt.Sprintf("SEVERED %.0fs", math.Ceil(head.Regrow)), rl.NewColor(230, 96, 120, 255)
		case head.Severed:
			label, ink = fmt.Sprintf("NEEDS %.0f MP", parameters.HydraRegrowCost), rl.Orange
		default:
			fill := bar
			fill.Width *= float32(max(0, min(1, head.HP/max(1, head.MaxHP))))
			rl.DrawRectangleRec(fill, color.RGBA{104, 160, 150, 255})
		}
		rl.DrawRectangleLinesEx(bar, 1, rl.Fade(rl.RayWhite, .25))
		if label != "" {
			primitives.DrawText(label, hx+int32(slot)/2-primitives.MeasureText(label, 12)/2, y+62, 12, ink)
		}
	}
}
