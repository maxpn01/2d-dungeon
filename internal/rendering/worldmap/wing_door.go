package worldmap

import (
	"math"
	"slices"

	"raylib-2d-sandbox/internal/simulation"
	mapdata "raylib-2d-sandbox/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (m *Map) drawWingDoor(snapshot simulation.Snapshot) {
	for _, door := range m.world.SecretDoors() {
		if !door.Interactable {
			continue
		}
		bounds := door.SanctuaryBorder()
		horizontal := bounds.W > bounds.H
		center := rl.NewVector2(float32(bounds.X+bounds.W/2), float32(bounds.Y+bounds.H/2))
		opened := slices.Contains(snapshot.OpenSecrets, door.ID)
		tint := rl.NewColor(158, 115, 225, 255)
		pulse := float32(.75 + .1*math.Sin(rl.GetTime()*1.4))
		if !opened {
			rl.DrawCircleGradient(center, 32, rl.Fade(tint, .08*pulse), rl.Blank)
		}
		progress := float32(snapshot.WingDoorOpen)
		progress = progress * progress * (3 - 2*progress)
		for i := 0; i < 2; i++ {
			direction := float32(-1)
			if i == 1 {
				direction = 1
			}
			shift := float32(i*mapdata.Tile) + direction*progress*mapdata.Tile
			x, y, w, h, index := int(bounds.X), int(bounds.Y+float64(shift)), mapdata.SurfaceMin, mapdata.Tile, door.Area.Y+i
			if horizontal {
				x, y, w, h, index = int(bounds.X+float64(shift)), int(bounds.Y), mapdata.Tile, mapdata.SurfaceMin, door.Area.X+i
			}
			drawSanctuaryBorder(x, y, w, h, index)
			at := rl.NewVector2(float32(x+w/2), float32(y+h/2))
			if !opened {
				rl.DrawCircleGradient(at, 26, rl.Fade(tint, .13*pulse), rl.Blank)
			}
			from, to := rl.NewVector2(at.X, at.Y-1.5), rl.NewVector2(at.X, at.Y+1.5)
			if horizontal {
				from, to = rl.NewVector2(at.X-1.5, at.Y), rl.NewVector2(at.X+1.5, at.Y)
			}
			rl.DrawLineV(from, to, rl.Fade(tint, .2*pulse))
		}
	}
}
