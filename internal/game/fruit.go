package game

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"image/color"
)

type Fruit struct {
	pos, size rl.Vector2
	color     color.RGBA
}

type FruitSpawner struct{ fruits []Fruit }

func (fs *FruitSpawner) draw() {
	for _, fruit := range fs.fruits {
		rl.DrawRectangleV(fruit.pos, fruit.size, fruit.color)
	}
}
