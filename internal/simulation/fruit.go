package simulation

import (
	"fmt"
	"math/rand/v2"
)

const (
	FruitSize       = 15.0
	InitialFruits   = 8
	MaxFruits       = 20
	fruitSpawnTicks = TickRate
)

type Fruit struct {
	ID         string
	X, Y, Size float64
}

func (w *World) spawnFruit() {
	if len(w.fruits) >= MaxFruits {
		return
	}
	for range 32 {
		x := MapMin + rand.Float64()*(MapMax-MapMin-FruitSize)
		y := MapMin + rand.Float64()*(MapMax-MapMin-FruitSize)
		overlaps := false
		for _, fruit := range w.fruits {
			if x < fruit.X+fruit.Size && x+FruitSize > fruit.X &&
				y < fruit.Y+fruit.Size && y+FruitSize > fruit.Y {
				overlaps = true
				break
			}
		}
		for _, player := range w.players {
			if x < player.X+PlayerSize && x+FruitSize > player.X && y < player.Y+PlayerSize && y+FruitSize > player.Y {
				overlaps = true
				break
			}
		}
		if overlaps {
			continue
		}
		w.nextFruit++
		w.fruits = append(w.fruits, Fruit{ID: fmt.Sprintf("f%d", w.nextFruit), X: x, Y: y, Size: FruitSize})
		return
	}
}
