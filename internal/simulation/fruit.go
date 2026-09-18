package simulation

import (
	"fmt"
	"math/rand/v2"
	"raylib-2d-sandbox/internal/terrain"
)

const (
	FruitSize       = 15.0
	fruitSpawnTicks = TickRate
)

type Fruit struct {
	Patch         string
	Source        string
	HP            float64
	AvailableTick uint64
	ExpiresTick   uint64
	ID            string
	X, Y, Size    float64
}

func (w *World) spawnFruit() {
	for _, patch := range terrain.Patches {
		count := 0
		for _, f := range w.fruits {
			if f.Patch == patch.ID {
				count++
			}
		}
		if count >= patch.Capacity {
			w.patchNext[patch.ID] = w.tick + uint64(patch.Seconds*TickRate)
			continue
		}
		if w.tick < w.patchNext[patch.ID] {
			continue
		}
		if w.spawnPatchFruit(patch) {
			w.patchNext[patch.ID] = w.tick + uint64(patch.Seconds*TickRate)
		}
	}
}
func (w *World) spawnPatchFruit(p terrain.FeedingPatch) bool {
	for range 64 {
		x, y := terrain.CellPosition(p.X+rand.IntN(2*p.Radius+1)-p.Radius, p.Y+rand.IntN(2*p.Radius+1)-p.Radius, FruitSize)
		if !terrain.Walkable(x, y, FruitSize, FruitSize) {
			continue
		}
		overlaps := false
		overlap := func(ax, ay, size float64) bool {
			return x < ax+size && x+FruitSize > ax && y < ay+size && y+FruitSize > ay
		}
		for _, f := range w.fruits {
			if overlap(f.X, f.Y, f.Size) {
				overlaps = true
				break
			}
		}
		for _, a := range w.ais {
			if a.respawnTick == 0 && overlap(a.X, a.Y, AISize) {
				overlaps = true
				break
			}
		}
		for _, a := range w.players {
			if !a.Dead && overlap(a.X, a.Y, PlayerSize) {
				overlaps = true
				break
			}
		}
		if overlaps {
			continue
		}
		w.nextFruit++
		w.fruits = append(w.fruits, Fruit{ID: fmt.Sprintf("f%d", w.nextFruit), Patch: p.ID, X: x, Y: y, Size: FruitSize})
		return true
	}
	return false
}
