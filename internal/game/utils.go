package game

import (
	"math/rand"
	"raylib-2d-sandbox/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func checkCollisions(pos1, size1, pos2, size2 rl.Vector2) bool {
	overlapX := pos1.X < pos2.X+size2.X && pos1.X+size1.X > pos2.X
	overlapY := pos1.Y < pos2.Y+size2.Y && pos1.Y+size1.Y > pos2.Y

	return overlapX && overlapY
}

func findSquaredEuclideanDistance(pos1 rl.Vector2, pos2 rl.Vector2) float32 {
	distanceX := pos1.X - pos2.X
	distanceY := pos1.Y - pos2.Y

	return distanceX*distanceX + distanceY*distanceY
}

func clamp(min float32, targetPos *float32, targetSize *float32, max float32) {
	if *targetPos < min {
		*targetPos = min
	}

	if *targetPos+*targetSize > max {
		*targetPos = max - *targetSize
	}
}

func calcExpForNextLvl(lvl int) float32 {
	return float32(simulation.XPForNextLevel(lvl))
}

func randomBetween[T int | float32](min, max T) T {
	return min + T(rand.Intn(int(max-min)+1))
}

func rectCenter(pos, size rl.Vector2) rl.Vector2 {
	return rl.NewVector2(pos.X+size.X/2, pos.Y+size.Y/2)
}
