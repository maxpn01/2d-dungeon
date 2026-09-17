package game

import (
	"image/color"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Fruit struct {
	pos   rl.Vector2
	size  rl.Vector2
	color color.RGBA
}

type FruitSpawner struct {
	fruits []Fruit

	fruitSize  rl.Vector2
	fruitColor color.RGBA

	fruitSpawnTimer              float32
	fruitSpawnInterval           int
	fruitSpawnIntervalMaxSeconds int

	maxFruits int
}

func NewFruitSpawner(
	size rl.Vector2,
	color color.RGBA,
	spawnIntervalMaxSeconds int,
	maxFruits int) *FruitSpawner {
	return &FruitSpawner{
		fruits:                       []Fruit{},
		fruitSize:                    size,
		fruitColor:                   color,
		fruitSpawnTimer:              0,
		fruitSpawnInterval:           1 + rand.Intn(spawnIntervalMaxSeconds),
		fruitSpawnIntervalMaxSeconds: spawnIntervalMaxSeconds,
		maxFruits:                    maxFruits,
	}
}

func (fs *FruitSpawner) update(w *World, dt float32) {
	spawnFruit(fs, w.gameMap, dt)
}

func (fs *FruitSpawner) draw() {
	for _, fruit := range fs.fruits {
		rl.DrawRectangleV(fruit.pos, fruit.size, fruit.color)
	}
}

func spawnFruit(fs *FruitSpawner, gameMap *Map, dt float32) {
	fs.fruitSpawnTimer += dt

	if fs.fruitSpawnTimer >= float32(fs.fruitSpawnInterval) && len(fs.fruits) < fs.maxFruits {
		x, y := gameMap.edgePosX, gameMap.edgePosY

		fruitRandX := randomBetween(x.start, x.end-fs.fruitSize.X)
		fruitRandY := randomBetween(y.start, y.end-fs.fruitSize.Y)

		fs.fruits = append(fs.fruits, Fruit{
			pos:   rl.NewVector2(fruitRandX, fruitRandY),
			size:  fs.fruitSize,
			color: fs.fruitColor,
		})

		fs.fruitSpawnTimer = 0
		fs.fruitSpawnInterval = 1 + rand.Intn(fs.fruitSpawnIntervalMaxSeconds)
	}
}

func (fs *FruitSpawner) despawnFruit(fruitIndex int) {
	// when order matters, but much slower as it shifts the elements to the left
	// fs.fruits = append(fs.fruits[:fruitIndex], fs.fruits[fruitIndex+1:]...)
	fs.fruits[fruitIndex] = fs.fruits[len(fs.fruits)-1]
	fs.fruits = fs.fruits[:len(fs.fruits)-1]
}
