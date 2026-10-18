package worldmap

import (
	"fmt"
	"math/rand/v2"

	"dungeon-raylib-go/internal/parameters"
	"dungeon-raylib-go/internal/physics"
)

func wingRoom(layout Layout, id string) (Layout, string, error) {
	room, _ := DefaultLayout().Room(WingChamber)
	doorID, passageID := WingDoor, AerieEntrance

	wall, offset := "", 0
	if id == "" {
		wall = []string{"left", "right"}[rand.IntN(2)]
		offset = parameters.SurfaceCells/2 + rand.IntN(parameters.SurfaceCells/2-4)
		id = fmt.Sprintf("%s:%d", wall, offset)
	}
	found := false
	var door CellRect
	for _, candidate := range []string{"left", "right"} {
		first, last := parameters.SurfaceCells/2, parameters.SurfaceCells-5
		for at := first; at <= last; at++ {
			if id != fmt.Sprintf("%s:%d", candidate, at) {
				continue
			}
			wall, offset, found = candidate, at, true
		}
	}
	if !found {
		return layout, "", fmt.Errorf("invalid wing room placement %q", id)
	}
	switch wall {
	case "left":
		room.Area = Cells(-15, offset-3, 14, 8)
		room.Statues[0].At = At(2, 3)
		door = Cells(-1, offset, 1, 2)
	case "right":
		room.Area = Cells(parameters.SurfaceCells+1, offset-3, 14, 8)
		room.Statues[0].At = At(11, 3)
		door = Cells(parameters.SurfaceCells, offset, 1, 2)
	}

	layout.SetRoom(room)
	layout.SecretDoors = append(layout.SecretDoors, SecretDoor{ID: doorID, RoomID: room.ID, Area: door, Interactable: true})
	layout.SetPassage(Passage{ID: passageID, From: Location{At: door.Origin()}, To: Location{At: At(door.Right()-1, door.Bottom()-1)}, Width: 1})
	return layout, id, nil
}

func (door SecretDoor) SanctuaryBorder() physics.Rect {
	area := door.Area.Bounds()
	if door.Area.Width > door.Area.Height {
		y := float64(parameters.SurfaceMax)
		return physics.Rect{X: area.X, Y: y, W: area.W, H: parameters.SurfaceMin}
	}
	x := 0.0
	if door.Area.X > 0 {
		x = parameters.SurfaceMax
	}
	return physics.Rect{X: x, Y: area.Y, W: parameters.SurfaceMin, H: area.H}
}

func IsWingStatue(id string) bool { return id == WingStatue }
