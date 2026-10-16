package worldmap

import (
	"image/color"
	"math"

	"raylib-2d-sandbox/internal/rendering/primitives"
	"raylib-2d-sandbox/internal/simulation"
	mapdata "raylib-2d-sandbox/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func drawStatueGate(bounds mapdata.GateBounds, gate simulation.EntranceGateState, prompt bool) {
	x, y := bounds.Left, bounds.Y
	w := bounds.Right - x
	if w <= 0 {
		return
	}
	shadow := color.RGBA{17, 22, 30, 255}
	inset := color.RGBA{32, 39, 49, 255}
	iron := color.RGBA{67, 78, 87, 255}
	eye := color.RGBA{166, 79, 87, 255}
	leaf := int(float64(w/2) * (1 - gate.Open))
	for side := 0; side < 2; side++ {
		dx := x
		if side == 1 {
			dx = x + w - leaf
		}
		if leaf <= 0 {
			continue
		}
		dungeonRect(dx, y-30, leaf, 60, shadow)
		if leaf > 6 {
			dungeonRect(dx+1, y-29, leaf-2, 50, stone)
			dungeonRect(dx+3, y-27, leaf-6, 2, paleStone)
			dungeonRect(dx+3, y-24, 2, 42, iron)
			dungeonRect(dx+1, y+21, leaf-2, 8, inset)
		}
		for offset := 40; offset < leaf; offset += 40 {
			dungeonRect(dx+offset, y-29, 1, 50, shadow)
		}
		if leaf > 28 {
			handle := dx + leaf - 16
			if side == 1 {
				handle = dx + 8
			}
			dungeonRect(handle, y-8, 8, 16, inset)
			dungeonRect(handle+2, y-6, 4, 3, paleStone)
			dungeonRect(handle+2, y+3, 4, 3, paleStone)
		}
	}
	for _, sx := range []int{x - 40, x + w} {
		dungeonRect(sx, y-60, 40, 120, shadow)
		for row := -60; row < 60; row += 40 {
			dungeonRect(sx+1, y+row+1, 38, 30, stone)
			dungeonRect(sx+3, y+row+3, 34, 2, paleStone)
			dungeonRect(sx+3, y+row+6, 2, 22, iron)
		}
		dungeonRect(sx+6, y-32, 28, 58, inset)
		dungeonRect(sx+8, y-30, 24, 48, stone)
		dungeonRect(sx+11, y-15, 3, 29, iron)
		dungeonRect(sx+26, y-15, 3, 29, shadow)
		dungeonRect(sx+4, y-27, 7, 23, iron)
		dungeonRect(sx+29, y-27, 7, 23, iron)
		dungeonRect(sx+10, y-44, 20, 21, shadow)
		dungeonRect(sx+11, y-43, 18, 17, stone)
		dungeonRect(sx+13, y-41, 14, 2, paleStone)
		dungeonRect(sx+13, y-34, 4, 2, eye)
		dungeonRect(sx+23, y-34, 4, 2, eye)
		dungeonRect(sx+16, y-28, 8, 1, shadow)
		dungeonRect(sx+7, y+18, 26, 8, shadow)
		dungeonRect(sx+9, y+18, 22, 2, paleStone)
		center := rl.NewVector2(float32(sx+20), float32(y+44))
		pulse := float32(.6 + .1*math.Sin(rl.GetTime()*1.6))
		rl.DrawCircleGradient(center, 15, rl.Fade(primitives.RuneColor, .16*pulse), rl.Blank)
		rl.DrawRectangleLines(int32(sx+15), int32(y+39), 10, 10, rl.Fade(primitives.RuneColor, pulse))
	}
	if prompt {
		label := "[E] OPEN GATE"
		if gate.Remaining > 0 {
			label = "GATE OPEN"
		}
		promptY := y + 72

		drawInteractionPrompt(label, rl.NewVector2(float32(x+w/2), float32(promptY)))
	}
}

func drawInteractionPrompt(label string, at rl.Vector2) {
	width := primitives.MeasureText(label, 16)
	primitives.DrawText(label, int32(at.X)-width/2, int32(at.Y), 16, primitives.RuneColor)
}
