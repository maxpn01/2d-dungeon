package game

import (
	"raylib-2d-sandbox/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func drawSafeZones(zones []simulation.SafeZone) {
	for _, zone := range zones {
		center := rl.NewVector2(float32(zone.X), float32(zone.Y))
		radius := float32(zone.Radius)
		rl.DrawCircleV(center, radius, rl.NewColor(16, 48, 35, 255))
		rl.DrawRing(center, radius-2, radius, 0, 360, 96, rl.NewColor(90, 205, 135, 255))
	}
}

func drawSafeProtection(x, y, size, mass float64, zones []simulation.SafeZone) {
	// The zone itself is the only visual cue; entering it does not add a halo.
}
