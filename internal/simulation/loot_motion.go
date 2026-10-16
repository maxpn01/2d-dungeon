package simulation

import (
	"math"

	"raylib-2d-sandbox/internal/worldmap"
)

func (w *World) advanceFood() {
	for i := range w.fruits {
		f := &w.fruits[i]
		if !f.Ejected {
			continue
		}
		f.X, f.Y = worldmap.Default().Move(f.X, f.Y, f.VelocityX*StepSeconds, f.VelocityY*StepSeconds, f.Size, f.Size)
		decay := math.Exp(-7 * StepSeconds)
		f.VelocityX *= decay
		f.VelocityY *= decay
	}
}
