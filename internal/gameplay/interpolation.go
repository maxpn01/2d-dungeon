package gameplay

import (
	"raylib-2d-sandbox/internal/simulation"
	"raylib-2d-sandbox/internal/worldmap"
)

func (w *World) interpolateActors() {
	if w.Dead() || len(w.previousSnapshot.Players) == 0 {
		return
	}
	alpha := w.accumulator / simulation.StepSeconds
	interpolate := func(ax, ay, bx, by float64) (float64, float64) {
		if worldmap.Default().FloorAt(ax, ay) != worldmap.Default().FloorAt(bx, by) {
			return bx, by
		}
		return worldmap.Default().Move(ax, ay, (bx-ax)*alpha, (by-ay)*alpha, simulation.PlayerSize, simulation.PlayerSize)
	}
	before, after := w.previousSnapshot.Players[0], w.snapshot.Players[0]
	if before.Flight.Flying || after.Flight.Flying {
		w.player.X, w.player.Y = worldmap.Default().WithSecrets(after.OpenSecrets).MoveFlying(before.X, before.Y, (after.X-before.X)*alpha, (after.Y-before.Y)*alpha, simulation.PlayerSize)
	} else {
		w.player.X, w.player.Y = interpolate(before.X, before.Y, after.X, after.Y)
	}
	for i, cell := range after.Split {
		if !cell.Active {
			continue
		}
		startX, startY := cell.X-after.X, cell.Y-after.Y
		if i < len(before.Split) && before.Split[i].Active {
			startX, startY = before.Split[i].X-before.X, before.Split[i].Y-before.Y
		}
		w.player.Split[i].X = w.player.X + startX + (cell.X-after.X-startX)*alpha
		w.player.Split[i].Y = w.player.Y + startY + (cell.Y-after.Y-startY)*alpha
	}
	for i := range w.ais {
		ai := &w.ais[i]
		current := w.snapshot.AIs[i]
		for _, previous := range w.previousSnapshot.AIs {
			if ai.ID != previous.ID {
				continue
			}
			ai.X, ai.Y = interpolate(previous.X, previous.Y, current.X, current.Y)
			for j, cell := range current.Split {
				if cell.Active && j < len(previous.Split) && previous.Split[j].Active {
					ai.Split[j].X, ai.Split[j].Y = interpolate(previous.Split[j].X, previous.Split[j].Y, cell.X, cell.Y)
				}
			}
			break
		}
	}
}
