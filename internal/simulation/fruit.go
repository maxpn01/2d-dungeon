package simulation

import (
	"fmt"
	"math/rand/v2"

	"raylib-2d-sandbox/internal/parameters"
)

const (
	FruitSize       = parameters.FruitSize
	InitialFruits   = parameters.InitialFruits
	MaxFruits       = parameters.MaxFruits
	fruitSpawnTicks = parameters.FruitSpawnTicks
)

type Fruit struct {
	Ejected              bool
	VelocityX, VelocityY float64
	OwnerPickupTick      uint64
	Patch                string
	Source               string
	HP                   float64
	AvailableTick        uint64
	ExpiresTick          uint64
	ID                   string
	X, Y, Size           float64
}

// Natural food comes in three sizes. HP is stored on each fruit so the
// reward stays tied to the pellet that was spawned, including in snapshots.
var naturalFood = [...]struct {
	size, hp float64
}{
	{10, 0.5},
	{FruitSize, FruitHP},
	{23, 2.5},
}

func naturalFoodType() (size, hp float64) {
	roll := rand.IntN(100)
	switch {
	case roll < 35:
		return naturalFood[0].size, naturalFood[0].hp
	case roll < 85:
		return naturalFood[1].size, naturalFood[1].hp
	default:
		return naturalFood[2].size, naturalFood[2].hp
	}
}

func (f Fruit) rewardHP() float64 {
	if f.HP > 0 {
		return f.HP
	}
	return FruitHP
}

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
		overlaps := false
		for _, fruit := range w.fruits {
			if x < fruit.X+fruit.Size && x+size > fruit.X &&
				y < fruit.Y+fruit.Size && y+size > fruit.Y {
				overlaps = true
				break
			}
		}
		for _, player := range w.players {
			if player.Dead {
				continue
			}
			for _, cell := range player.Split {
				if cell.Active && x < cell.X+PlayerSize && x+size > cell.X && y < cell.Y+PlayerSize && y+size > cell.Y {
					overlaps = true
					break
				}
			}
			if x < player.X+PlayerSize && x+size > player.X && y < player.Y+PlayerSize && y+size > player.Y {
				overlaps = true
				break
			}
		}
		for _, ai := range w.ais {
			if ai.respawnTick != 0 {
				continue
			}
			for _, cell := range ai.Split {
				if cell.Active && x < cell.X+AISize && x+size > cell.X && y < cell.Y+AISize && y+size > cell.Y {
					overlaps = true
					break
				}
			}
			if x < ai.X+AISize && x+size > ai.X && y < ai.Y+AISize && y+size > ai.Y {
				overlaps = true
				break
			}
		}
		if overlaps {
			continue
		}
		w.nextFruit++
		w.fruits = append(w.fruits, Fruit{ID: fmt.Sprintf("f%d", w.nextFruit), X: x, Y: y, Size: size, HP: hp})
		return
	}
}
