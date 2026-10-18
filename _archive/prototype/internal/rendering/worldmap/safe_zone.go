package worldmap

import (
	"math"

	"dungeon-raylib-go/internal/rendering/primitives"
	"dungeon-raylib-go/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func drawSafeZones(zones []simulation.SafeZone) {
	for _, zone := range zones {
		center := rl.NewVector2(float32(zone.X), float32(zone.Y))
		radius := float32(zone.Radius)
		pulse := float32(.7 + .1*math.Sin(rl.GetTime()*1.2+zone.X))
		rl.DrawCircleV(center, radius, rl.NewColor(19, 43, 40, 255))
		rl.DrawCircleGradient(center, radius, rl.Fade(primitives.RuneColor, .045*pulse), rl.Blank)
		rl.DrawRing(center, radius-2, radius, 0, 360, 96, rl.NewColor(90, 205, 135, 255))
		rl.DrawRing(center, radius-7, radius-6, 0, 360, 96, rl.Fade(sanctuaryGold, .4))
		for i := 0; i < 8; i++ {
			angle := float32(i) * 45
			rl.DrawRing(center, radius-14, radius-13, angle+7, angle+33, 12, rl.Fade(primitives.RuneColor, .5*pulse))
			drawSanctuaryRune(sanctuaryPoint(center, float64(i)*math.Pi/4, float64(radius-23)), 4, rl.Fade(primitives.RuneColor, .65*pulse))
		}
	}
}
