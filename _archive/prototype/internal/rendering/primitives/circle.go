package primitives

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func drawCircle(r rl.Rectangle, tint color.RGBA) {
	rl.DrawCircleV(rl.NewVector2(r.X+r.Width/2, r.Y+r.Height/2), r.Width/2, tint)
}
