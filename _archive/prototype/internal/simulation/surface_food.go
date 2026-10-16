package simulation

import (
	"fmt"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/physics"
)

const (
	InitialFruits   = parameters.InitialFruits
	MaxFruits       = parameters.MaxFruits
	fruitSpawnTicks = parameters.FruitSpawnTicks
)

func (w *World) spawnFruit() {
	natural := 0
	for _, fruit := range w.fruits {
		if fruit.Source == "" {
			natural++
		}
	}
	if natural >= MaxFruits {
		return
	}
	size, hp := naturalFoodType()
	for range 32 {
		x := MapMin + rand.Float64()*(MapMax-MapMin-size)
		y := MapMin + rand.Float64()*(MapMax-MapMin-size)
		body := physics.Square(x, y, size)
		overlaps := false
		for _, fruit := range w.fruits {
			if body.Overlaps(physics.Square(fruit.X, fruit.Y, fruit.Size)) {
				overlaps = true
				break
			}
		}
		for _, player := range w.players {
			if player.Dead {
				continue
			}
			for _, cell := range player.Split {
				if cell.Active && body.Overlaps(physics.Square(cell.X, cell.Y, PlayerSize)) {
					overlaps = true
					break
				}
			}
			if body.Overlaps(physics.Square(player.X, player.Y, PlayerSize)) {
				overlaps = true
				break
			}
		}
		for _, ai := range w.ais {
			if ai.respawnTick != 0 {
				continue
			}
			for _, cell := range ai.Split {
				if cell.Active && body.Overlaps(physics.Square(cell.X, cell.Y, AISize)) {
					overlaps = true
					break
				}
			}
			if body.Overlaps(physics.Square(ai.X, ai.Y, AISize)) {
				overlaps = true
				break
			}
		}
		if overlaps {
			continue
		}
		hp, mp := naturalFoodReward(hp)
		w.nextFruit++
		w.fruits = append(w.fruits, Fruit{ID: fmt.Sprintf("f%d", w.nextFruit), X: x, Y: y, Size: size, HP: hp, MP: mp})
		return
	}
}
