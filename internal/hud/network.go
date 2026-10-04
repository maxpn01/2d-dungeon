package hud

import (
	"fmt"
	"time"

	"raylib-2d-sandbox/internal/rendering/primitives"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func DrawPing(ping time.Duration, measured bool, windowWidth int32) {
	label := "Ping: --"
	if measured {
		label = fmt.Sprintf("Ping: %d ms", (ping+time.Millisecond/2)/time.Millisecond)
	}
	x := min(int32(320), windowWidth-220)
	width := primitives.MeasureText(label, 18) + 24
	if x <= 0 || x+width > windowWidth-220 {
		return
	}
	rl.DrawRectangle(x, 0, width, 38, rl.Fade(rl.Black, .8))
	primitives.DrawText(label, x+12, 10, 18, rl.RayWhite)
}
