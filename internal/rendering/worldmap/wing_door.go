package worldmap

import (
	"math"
	"slices"

	"raylib-2d-sandbox/internal/protocol"
	mapdata "raylib-2d-sandbox/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (m *Map) drawWingDoor(snapshot protocol.Snapshot) {
	for _, door := range m.world.SecretDoors() {
		if door.ID != mapdata.WingDoor {
			continue
		}
		bounds := door.Area.Bounds()
		center := rl.NewVector2(5, float32(bounds.Y+bounds.H/2))
		opened := slices.Contains(snapshot.OpenSecrets, door.ID)
		tint := rl.NewColor(158, 115, 225, 255)
		pulse := float32(.75 + .1*math.Sin(rl.GetTime()*1.4))
		if !opened {
			rl.DrawCircleGradient(center, 32, rl.Fade(tint, .08*pulse), rl.Blank)
		}
		progress := float32(snapshot.WingDoorOpen)
		progress = progress * progress * (3 - 2*progress)
		for i := 0; i < 2; i++ {
			y := float32(bounds.Y) + float32(i*mapdata.Tile)
			direction := float32(-1)
			if i == 1 {
				direction = 1
			}
			y += direction * progress * mapdata.Tile
			drawSanctuaryBorder(0, int(y), mapdata.SurfaceMin, mapdata.Tile, door.Area.Y+i)
			if !opened {
				rl.DrawCircleGradient(rl.NewVector2(5, y+20), 26, rl.Fade(tint, .13*pulse), rl.Blank)
			}
			rl.DrawLine(5, int32(y+19), 5, int32(y+22), rl.Fade(tint, .2*pulse))
			if opened && progress < 1 {
				rl.DrawLineEx(rl.NewVector2(10, y+20), rl.NewVector2(10, y+20-direction*10), 1, rl.Fade(tint, .25))
			}
		}
	}
}
