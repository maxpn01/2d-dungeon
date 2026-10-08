package rendering

import (
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func DrawFruit(x, y, size float32, tint color.RGBA, mass bool, mana bool) {
	phase := rl.GetTime()*2.4 + float64(x*.07+y*.11)
	pulse := float32(.9 + .1*math.Sin(phase))
	center := rl.NewVector2(x+size/2, y+size/2+float32(math.Sin(phase*.65))*1.5)
	radius := size * .34
	foodTint := HealthFoodColor
	coreTint := rl.NewColor(207, 255, 213, 255)
	glowRadius, glowAlpha := float32(2.8), float32(.2)
	if mana {
		foodTint = ManaFoodColor
		coreTint = rl.NewColor(193, 222, 255, 255)
		glowRadius, glowAlpha = 3.2, .3
	}
	rl.DrawCircleGradient(center, radius*glowRadius, rl.Fade(foodTint, glowAlpha*pulse), rl.Fade(foodTint, 0))
	rl.DrawCircleV(center, radius*1.18, rl.Fade(foodTint, .18*pulse))
	rl.DrawCircleV(center, radius, rl.Fade(foodTint, .94))
	rl.DrawCircleV(center, radius*.36, rl.Fade(coreTint, .7))
	if mass {
		rl.DrawCircleLinesV(center, radius*1.5, rl.Fade(tint, .35*pulse))
	}
}

var EnergyColor = color.RGBA{112, 229, 214, 255}

var HealthFoodColor = color.RGBA{101, 220, 120, 255}
var ManaFoodColor = color.RGBA{57, 140, 255, 255}
