package game

import (
	"math/rand"
	"raylib-2d-sandbox/internal/simulation"
)

func calcExpForNextLvl(lvl int) float32 {
	return float32(simulation.XPForNextLevel(lvl))
}

func randomBetween[T int | float32](min, max T) T {
	return min + T(rand.Intn(int(max-min)+1))
}
