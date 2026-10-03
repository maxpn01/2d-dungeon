package simulation

import "raylib-2d-sandbox/internal/worldmap"

func (w *World) replenishPrivateRoute(entranceID string) {
	for _, room := range worldmap.Default().Rooms() {
		if !room.HasRole(worldmap.Private) {
			continue
		}
		owned := false
		for _, gate := range room.Gates {
			if gate.ID == entranceID {
				owned = true
				break
			}
		}
		if !owned {
			continue
		}
		oldGuards := make(map[string]bool)
		for i := range w.ais {
			ai := &w.ais[i]
			if ai.den.RoomID == room.ID {
				oldGuards[ai.ID] = true
				*ai = w.newDenAI(ai.den)
			}
		}
		remaining := w.fruits[:0]
		for _, fruit := range w.fruits {
			if !oldGuards[fruit.Source] {
				remaining = append(remaining, fruit)
			}
		}
		w.fruits = remaining
		shots := w.projectiles[:0]
		for _, shot := range w.projectiles {
			if !oldGuards[shot.Owner] {
				shots = append(shots, shot)
			}
		}
		w.projectiles = shots
		for _, patch := range worldmap.Default().Patches() {
			if patch.RoomID == room.ID {
				w.refillPatch(patch)
			}
		}
		return
	}
}
