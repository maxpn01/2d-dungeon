package game

import (
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Fruit struct {
	mass      bool
	mana      bool
	pos, size rl.Vector2
	color     color.RGBA
}

type FruitSpawner struct{ fruits []Fruit }

func (fs *FruitSpawner) draw() {
	for _, fruit := range fs.fruits {
		drawFruit(fruit.pos.X, fruit.pos.Y, fruit.size.X, fruit.color, fruit.mass, fruit.mana)
	}
}

func drawFruit(x, y, size float32, tint color.RGBA, mass bool, mana bool) {
	phase := rl.GetTime()*2.4 + float64(x*.07+y*.11)
	pulse := float32(.9 + .1*math.Sin(phase))
	center := rl.NewVector2(x+size/2, y+size/2+float32(math.Sin(phase*.65))*1.5)
	radius := size * .34
	if !mass {
		tint = foodColor(size)
	}
	if mana {
		tint = rl.SkyBlue
	}
	rl.DrawCircleGradient(center, radius*2.8, rl.Fade(tint, .2*pulse), rl.Fade(tint, 0))
	rl.DrawCircleV(center, radius*1.18, rl.Fade(tint, .18*pulse))
	rl.DrawCircleV(center, radius, rl.Fade(tint, .94))
	rl.DrawCircleV(center, radius*.36, rl.Fade(rl.White, .7))
	if mass {
		rl.DrawCircleLinesV(center, radius*1.5, rl.Fade(tint, .35*pulse))
	}
}

var energyColor = color.RGBA{112, 229, 214, 255}

func foodColor(size float32) color.RGBA {
	switch {
	case size < 12:
		return energyColor
	case size < 19:
		return color.RGBA{185, 145, 245, 255}
	default:
		return color.RGBA{255, 199, 109, 255}
	}
}
