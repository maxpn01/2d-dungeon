package primitives

import (
	"image/color"

	"dungeon-raylib-go/internal/appearance"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func CharacterColor(c appearance.Character) color.RGBA {
	switch c {
	case appearance.Triangle:
		return rl.Yellow
	default:
		return rl.Red
	}
}

func DrawCharacterShape(c appearance.Character, r rl.Rectangle, tint color.RGBA) {
	switch c {
	case appearance.Triangle:
		rl.DrawTriangle(rl.NewVector2(r.X+r.Width/2, r.Y), rl.NewVector2(r.X, r.Y+r.Height), rl.NewVector2(r.X+r.Width, r.Y+r.Height), tint)
	default:
		rl.DrawRectangleRec(r, tint)
	}
}
