package ui

import (
	"image/color"

	"raylib-2d-sandbox/internal/rendering/primitives"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func DrawFullMapBackground(background color.RGBA, height int32) {
	rl.ClearBackground(background)
	primitives.DrawCenteredText("WORLD MAP", 24, 28, rl.RayWhite)
	primitives.DrawCenteredText("M / Esc to close", height-44, 20, rl.LightGray)
}
