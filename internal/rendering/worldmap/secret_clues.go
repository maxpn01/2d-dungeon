package worldmap

import (
	"math"
	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/rendering/primitives"
	mapdata "raylib-2d-sandbox/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (m *Map) drawSecretInscriptions() {
	for _, door := range m.world.SecretDoors() {
		if door.Interactable || m.world.Cell(door.Area.X, door.Area.Y) != 0 {
			continue
		}
		cx, cy := door.Area.X+door.Area.Width/2, door.Area.Y+door.Area.Height/2
		candidates := []mapdata.Cell{mapdata.At(door.Area.X-1, cy), mapdata.At(door.Area.Right(), cy), mapdata.At(cx, door.Area.Y-1), mapdata.At(cx, door.Area.Bottom())}
		for _, cell := range candidates {
			at := cell.Center()
			if m.world.Cell(cell.X, cell.Y) == 0 || m.world.CellFloor(cell.X, cell.Y) != m.floor || m.world.SecretHidden(at.X, at.Y) || !m.Exploration.Visible([2]int{cell.X, cell.Y}) {
				continue
			}
			center := rl.NewVector2(float32(at.X), float32(at.Y))
			distance := math.Hypot(float64(m.Exploration.focus.X-center.X), float64(m.Exploration.focus.Y-center.Y))
			if distance > parameters.SecretRuneRevealDistance {
				continue
			}
			tint := rl.NewColor(133, 149, 163, 255)
			rl.DrawPolyLines(center, 4, 7, 45, rl.Fade(tint, .55))
			rl.DrawLineEx(rl.NewVector2(center.X-11, center.Y+11), rl.NewVector2(center.X+11, center.Y+11), 1, rl.Fade(tint, .4))
			if distance < parameters.SecretRuneTextDistance {
				text := "Light sleeps behind the stone."
				if m.wallHits[door.ID] > 0 {
					text = "The stone yields. Strike again."
				}
				width := primitives.MeasureText(text, 13)
				rl.DrawRectangle(int32(center.X)-width/2-6, int32(center.Y)+19, width+12, 22, rl.Fade(rl.NewColor(9, 15, 22, 255), .88))
				primitives.DrawText(text, int32(center.X)-width/2, int32(center.Y)+23, 13, tint)
			}
			break
		}
	}
}
