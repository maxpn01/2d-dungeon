package primitives

import (
	"image/color"

	"raylib-2d-sandbox/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func CharacterColor(c simulation.Character) color.RGBA {
	switch c {
	case simulation.Triangle:
		return rl.Yellow
	default:
		return rl.Red
	}
}

func DrawCharacterShape(c simulation.Character, r rl.Rectangle, tint color.RGBA) {
	switch c {
	case simulation.Triangle:
		rl.DrawTriangle(rl.NewVector2(r.X+r.Width/2, r.Y), rl.NewVector2(r.X, r.Y+r.Height), rl.NewVector2(r.X+r.Width, r.Y+r.Height), tint)
	default:
		rl.DrawRectangleRec(r, tint)
	}
}
