package worldmap

import (
	"image/color"

	"raylib-2d-sandbox/internal/rendering/primitives"
	"raylib-2d-sandbox/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func drawStatues(statues []simulation.StatueState, active string, px, py float64, dead bool, exploration *Exploration) {
	for _, s := range statues {
		if !exploration.EntityVisible(s.X, s.Y, 0) {
			continue
		}
		x, y := int32(s.X), int32(s.Y)
		tint := primitives.RuneColor
		if s.ClaimedBy != "" {
			tint = color.RGBA{73, 83, 88, 255}
		}
		rl.DrawEllipse(x, y+19, 31, 12, rl.Fade(rl.Black, .5))
		rl.DrawRectangle(x-27, y+10, 54, 12, stone)
		rl.DrawRectangle(x-22, y+6, 44, 7, paleStone)
		rl.DrawRectangle(x-13, y-25, 26, 31, stone)
		rl.DrawRectangle(x-9, y-41, 18, 17, paleStone)
		rl.DrawRectangle(x-23, y-20, 9, 17, paleStone)
		rl.DrawRectangle(x+14, y-20, 9, 17, paleStone)
		rl.DrawRectangle(x-20, y-23, 40, 29, color.RGBA{29, 39, 47, 255})
		rl.DrawRectangleLines(x-20, y-23, 40, 29, paleStone)
		for i := int32(0); i < 3; i++ {
			rl.DrawLine(x-12, y-16+i*6, x+12-i*4, y-16+i*6, tint)
		}
		if s.ClaimedBy == "" {
			rl.DrawCircleLines(x, y-9, 37, rl.Fade(tint, .3))
			if !dead && active == "" && s.Near(px, py) {
				label := "[E] READ INSCRIPTION"
				width := primitives.MeasureText(label, 16)
				rl.DrawRectangle(x-width/2-6, y-71, width+12, 24, rl.Fade(rl.Black, .8))
				primitives.DrawText(label, x-width/2, y-68, 16, tint)
			}
		} else if !dead && s.Near(px, py) {
			label := "INSCRIPTION CLAIMED"
			primitives.DrawText(label, x-primitives.MeasureText(label, 14)/2, y-68, 14, rl.Gray)
		}
	}
}
