package simulation

import (
	"fmt"
	"math"
	"math/rand/v2"

	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/physics"
	"2d-dungeon/internal/worldmap"
)

type Fruit struct {
	Ejected              bool
	VelocityX, VelocityY float64
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
	{parameters.FruitSize, parameters.FruitHP},
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
	return parameters.FruitHP
}

func naturalFoodReward(hp float64) (float64, float64) {
	if rand.IntN(100) < parameters.ManaFoodChance {
		return 0, parameters.FruitMP * hp / parameters.FruitHP
	}
	return hp, 0
}

func (w *World) replenishEntranceFood() {
	for _, patch := range worldmap.Default().Patches() {
		if w.tick < w.patchNext[patch.ID] {
			continue
		}
		w.refillPatch(patch)
	}
}

func (w *World) refillPatch(patch worldmap.FeedingPatch) {
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
	w.patchNext[patch.ID] = w.tick + uint64(math.Ceil(patch.Respawn.Seconds()*parameters.TickRate))
}

func (w *World) spawnPatchFruit(patch worldmap.FeedingPatch) bool {
	size, hp := naturalFoodType()
	if w.tick == 0 {
		kind := naturalFood[w.nextFruit%uint64(len(naturalFood))]
		size, hp = kind.size, kind.hp
	}
	centerX := float64(patch.At.X*parameters.Tile + parameters.Tile/2)
	centerY := float64(patch.At.Y*parameters.Tile + parameters.Tile/2)
	for range 40 {
		x := centerX + (rand.Float64()*2-1)*float64(patch.Radius*parameters.Tile) - size/2
		y := centerY + (rand.Float64()*2-1)*float64(patch.Radius*parameters.Tile) - size/2
		if !worldmap.Default().RoomContains(patch.RoomID, physics.Square(x, y, size)) || !worldmap.Default().Walkable(x, y, size, size) {
			continue
		}
		body := physics.Square(x, y, size).Expanded(5)
		occupied := false
		for _, fruit := range w.fruits {
			if body.Overlaps(physics.Square(fruit.X, fruit.Y, fruit.Size)) {
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
