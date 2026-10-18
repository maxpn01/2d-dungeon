package game

import (
	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/worldmap"
)

func (w *World) interpolateActors() {
	if w.Dead() || w.previousSnapshot.Player.ID == "" {
		return
	}
	alpha := w.accumulator / parameters.StepSeconds
	interpolate := func(ax, ay, bx, by float64) (float64, float64) {
		if worldmap.Default().FloorAt(ax, ay) != worldmap.Default().FloorAt(bx, by) {
			return bx, by
		}
		return worldmap.Default().Move(ax, ay, (bx-ax)*alpha, (by-ay)*alpha, parameters.PlayerSize, parameters.PlayerSize)
	}
	before, after := w.previousSnapshot.Player, w.snapshot.Player
	if before.Flight.Flying || after.Flight.Flying {
		w.player.X, w.player.Y = worldmap.Default().WithSecrets(after.OpenSecrets).MoveFlying(before.X, before.Y, (after.X-before.X)*alpha, (after.Y-before.Y)*alpha, parameters.PlayerSize)
	} else {
		w.player.X, w.player.Y = interpolate(before.X, before.Y, after.X, after.Y)
	}
	for i := range w.ais {
		ai := &w.ais[i]
		current := w.snapshot.AIs[i]
		for _, previous := range w.previousSnapshot.AIs {
			if ai.ID != previous.ID {
				continue
			}
			ai.X, ai.Y = interpolate(previous.X, previous.Y, current.X, current.Y)
			break
		}
	}
}
