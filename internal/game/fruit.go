package game

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"image/color"
	"math"
)

type Fruit struct {
	mass      bool
	pos, size rl.Vector2
	color     color.RGBA
}

type FruitSpawner struct{ fruits []Fruit }

func (fs *FruitSpawner) draw() {
	for _, fruit := range fs.fruits {
		drawFruit(fruit.pos.X, fruit.pos.Y, fruit.size.X, fruit.color, fruit.mass)
	}
}

func drawFruit(x, y, size float32, tint color.RGBA, mass bool) {
	if mass {
		pulse := float32(0.18 + 0.08*math.Sin(rl.GetTime()*6+float64(x+y)))
		rl.DrawRectangleLinesEx(rl.NewRectangle(x-3, y-3, size+6, size+6), 1, rl.Fade(tint, pulse+0.25))
	}
	rl.DrawRectangleRec(rl.NewRectangle(x, y, size, size), tint)
	if mass {
		rl.DrawRectangleRec(rl.NewRectangle(x+3, y+3, 3, 3), rl.Fade(rl.RayWhite, 0.8))
	}
}
