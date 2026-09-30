package simulation

import (
	"fmt"
	"math/rand/v2"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/terrain"
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
	HP, MP               float64
	AvailableTick        uint64
	ExpiresTick          uint64
	ID                   string
	X, Y, Size           float64
}

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
	if f.MP > 0 {
		return 0
	}
	if f.HP > 0 {
		return f.HP
	}
	return FruitHP
}

func naturalFoodReward(hp float64) (float64, float64) {
	if rand.IntN(100) < parameters.ManaFoodChance {
		return 0, parameters.FruitMP * hp / FruitHP
	}
	return hp, 0
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
		hp, mp := naturalFoodReward(hp)
		w.nextFruit++
		w.fruits = append(w.fruits, Fruit{ID: fmt.Sprintf("f%d", w.nextFruit), X: x, Y: y, Size: size, HP: hp, MP: mp})
		return
	}
}

func (w *World) replenishEntranceFood() {
	for _, patch := range terrain.Patches {
		if w.tick < w.patchNext[patch.ID] {
			continue
		}
		count := 0
		for _, fruit := range w.fruits {
			if fruit.Patch == patch.ID {
				count++
			}
		}
		for ; count < patch.Capacity; count++ {
			if !w.spawnPatchFruit(patch) {
				break
			}
		}
		w.patchNext[patch.ID] = w.tick + uint64(patch.Seconds*TickRate)
	}
}

func (w *World) spawnPatchFruit(patch terrain.FeedingPatch) bool {
	size, hp := naturalFoodType()
	if w.tick == 0 {
		kind := naturalFood[w.nextFruit%uint64(len(naturalFood))]
		size, hp = kind.size, kind.hp
	}
	centerX := float64(patch.X*terrain.Tile + terrain.Tile/2)
	centerY := float64(patch.Y*terrain.Tile + terrain.Tile/2)
	for range 40 {
		x := centerX + (rand.Float64()*2-1)*float64(patch.Radius*terrain.Tile) - size/2
		y := centerY + (rand.Float64()*2-1)*float64(patch.Radius*terrain.Tile) - size/2
		if !terrain.InEntranceRoom(x, y, size) || !terrain.Walkable(x, y, size, size) {
			continue
		}
		occupied := false
		for _, fruit := range w.fruits {
			if x < fruit.X+fruit.Size+5 && x+size+5 > fruit.X && y < fruit.Y+fruit.Size+5 && y+size+5 > fruit.Y {
				occupied = true
				break
			}
		}
		if occupied {
			continue
		}
		hp, mp := naturalFoodReward(hp)
		w.nextFruit++
		w.fruits = append(w.fruits, Fruit{ID: fmt.Sprintf("f%d", w.nextFruit), Patch: patch.ID, X: x, Y: y, Size: size, HP: hp, MP: mp})
		return true
	}
	return false
}
