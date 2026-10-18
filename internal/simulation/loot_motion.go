package simulation

import (
	"math"

	"dungeon-raylib-go/internal/parameters"
	"dungeon-raylib-go/internal/worldmap"
)

func (w *World) advanceFood() {
	for i := range w.fruits {
		f := &w.fruits[i]
		if !f.Ejected {
			continue
		}
		f.X, f.Y = worldmap.Default().Move(f.X, f.Y, f.VelocityX*parameters.StepSeconds, f.VelocityY*parameters.StepSeconds, f.Size, f.Size)
		decay := math.Exp(-7 * parameters.StepSeconds)
		f.VelocityX *= decay
		f.VelocityY *= decay
	}
}
