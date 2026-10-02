package worldmap

import (
	"image/color"
	"math"

	"raylib-2d-sandbox/internal/rendering/primitives"
	"raylib-2d-sandbox/internal/simulation"
	mapdata "raylib-2d-sandbox/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func drawStatues(statues []simulation.StatueState, active string, px, py float64, dead bool, exploration *Exploration) {
	for _, s := range statues {
		if !exploration.EntityVisible(s.X, s.Y, 0) {
			continue
		}
		x, y := int32(s.X), int32(s.Y)
		tint := primitives.RuneColor
		winged := s.ID == mapdata.WingStatue
		if winged {
			tint = color.RGBA{192, 220, 244, 255}
		}
		if s.ClaimedBy != "" {
			tint = color.RGBA{73, 83, 88, 255}
		}
		if winged {
			drawStatueWings(x, y, tint)
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
				if winged {
					label = "[E] LISTEN TO THE GUARDIAN"
				}
				width := primitives.MeasureText(label, 16)
				rl.DrawRectangle(x-width/2-6, y-71, width+12, 24, rl.Fade(rl.Black, .8))
				primitives.DrawText(label, x-width/2, y-68, 16, tint)
			}
		} else if !dead && !winged && s.Near(px, py) {
			label := "INSCRIPTION CLAIMED"
			primitives.DrawText(label, x-primitives.MeasureText(label, 14)/2, y-68, 14, rl.Gray)
		}
	}
}

func drawStatueWings(x, y int32, tint color.RGBA) {
	center := rl.NewVector2(float32(x), float32(y-16))
	pulse := float32(.65 + .15*math.Sin(rl.GetTime()*1.8))
	rl.DrawCircleGradient(center, 88, rl.Fade(tint, .14*pulse), rl.Blank)
	rl.DrawCircleLinesV(center, 53, rl.Fade(tint, .22*pulse))
	for _, side := range []int32{-1, 1} {
		for i := int32(0); i < 5; i++ {
			root := rl.NewVector2(float32(x+side*12), float32(y-18+i*3))
			tip := rl.NewVector2(float32(x+side*(44-i*4)), float32(y-53+i*10))
			end := rl.NewVector2(float32(x+side*20), float32(y-5+i*2))
			if side < 0 {
				rl.DrawTriangle(root, tip, end, paleStone)
			} else {
				rl.DrawTriangle(root, end, tip, paleStone)
			}
			rl.DrawLineEx(root, tip, 1, rl.Fade(tint, .6*pulse))
		}
	}
	rl.DrawPolyLines(rl.NewVector2(float32(x), float32(y-52)), 4, 7, 45, rl.Fade(tint, pulse))
}
