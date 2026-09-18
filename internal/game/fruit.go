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

// Motions are visual only; pickup positions and mass rewards stay authoritative.
func drawFruit(x, y, size float32, tint color.RGBA, mass bool) {
	phase := rl.GetTime()*2.8 + float64(x*.07+y*.11)
	pulse := float32(.85 + .15*math.Sin(phase))
	center := rl.NewVector2(x+size/2, y+size/2+float32(math.Sin(phase*.65))*1.5)
	radius := size * .34
	if !mass {
		tint = energyColor
	}
	rl.DrawCircleGradient(center, radius*3.5, rl.Fade(tint, .25*pulse), rl.Fade(tint, 0))
	rl.DrawCircleV(center, radius*1.25, rl.Fade(tint, .28*pulse))
	rl.DrawCircleV(center, radius, rl.Fade(tint, .9))
	rl.DrawCircleV(center, radius*.45, rl.NewColor(230, 255, 247, 255))
	if mass {
		rl.DrawCircleLinesV(center, radius*1.8, rl.Fade(tint, .45*pulse))
	}
}

var energyColor = color.RGBA{112, 229, 214, 255}
