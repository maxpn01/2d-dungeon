package simulation

import (
	"math"

	"2d-dungeon/internal/worldmap"
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
