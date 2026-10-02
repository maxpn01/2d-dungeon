package hud

import (
	"fmt"

	"raylib-2d-sandbox/internal/rendering/primitives"
	"raylib-2d-sandbox/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func DrawLocation(x, y float32, width, height int32) {
	name, hint := "SPAWN ZONE", "Head north through the dungeon entrance."
	if y < 0 {
		name, hint = "THE HOLLOW / FLOOR I", "Entrance lies to the south"
		passage, ok := worldmap.Default().Passage("surface-entrance")
		if ok && passage.ContainsWorld(float64(x), float64(y)) {
			name, hint = "I / ENTRANCE HALL", "The provision hall opens to the north."
		}
		for _, r := range worldmap.Default().Rooms() {
			if x >= float32(r.Area.X*40) && x < float32((r.Area.X+r.Area.Width)*40) && y >= float32(r.Area.Y*40) && y < float32((r.Area.Y+r.Area.Height)*40) {
				name = floorLabel(r.Floor) + " / " + r.Name
				if r.HasRole(worldmap.Private) {
					hint = "The statues guard the north gate. Press E nearby to open it."
				}
				if r.HasRole(worldmap.Refuge) {
					hint = "Arcane veils confuse creatures. They do not stop attacks."
				}
				if len(r.Portals) > 0 {
					hint = "Press E inside the circle to interact."
				}
				if r.ID == worldmap.OrbVault {
					hint = "Defeat the level-5 sentinels, then press E at the pedestal."
				}
				break
			}
		}

	}
	if room, ok := worldmap.Default().RoomAt(float64(x), float64(y)); ok && room.Floor > worldmap.FirstFloor {
		name = floorLabel(room.Floor) + " / " + room.Name
		if len(room.Portals) > 0 {
			hint = "Press E in the circle to return to floor I."
		}
	}

	rl.DrawRectangle(0, height-70, width, 70, rl.Fade(rl.Black, .8))
	primitives.DrawText(name, 24, height-57, 20, primitives.RuneColor)
	primitives.DrawText(hint, 24, height-29, 16, rl.LightGray)
}

func floorLabel(floor worldmap.Floor) string {
	switch floor {
	case worldmap.FirstFloor:
		return "I"
	case worldmap.SecondFloor:
		return "II"
	default:
		return fmt.Sprint(floor)
	}
}
