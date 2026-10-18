package worldmap

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"

	"2d-dungeon/internal/parameters"
)

type secretPlacement struct {
	host string
	id   string
	room CellRect
	door CellRect
}

func secretBase() Layout {
	layout := DefaultLayout()

	layout.RemoveRoom(WingChamber)
	layout.RemovePassage(AerieEntrance)
	layout.RemoveRoom(SecretLibrary)
	layout.RemovePassage(LibrarySecret)
	layout.SecretDoors = slices.DeleteFunc(layout.SecretDoors, func(door SecretDoor) bool {
		return door.RoomID == WingChamber || door.RoomID == SecretLibrary
	})
	return layout
}

func secretPlacements(layout Layout, room Room, span int) []secretPlacement {
	occupied := map[[2]int]bool{}
	reserve := func(area CellRect) {
		for y := area.Y; y < area.Bottom(); y++ {
			for x := area.X; x < area.Right(); x++ {
				occupied[[2]int{x, y}] = true
			}
		}
	}
	for _, existing := range layout.Rooms {
		reserve(existing.Area)
	}
	for _, passage := range layout.Passages {
		area, _ := layout.passageArea(passage)
		reserve(area)
	}
	reserve(Cells(0, 0, parameters.SurfaceCells, parameters.SurfaceCells))
	var placements []secretPlacement
	world := Cells(parameters.MinCellX+1, parameters.MinCellY+1, parameters.MaxCellX-parameters.MinCellX-2, parameters.MaxCellY-parameters.MinCellY-2)
	for _, host := range layout.Rooms {
		if !host.HasRole(ExplorationRoom) || host.HasRole(Private) || host.Floor != FirstFloor {
			continue
		}

		for wall := range 4 {
			width, height := room.Area.Width, room.Area.Height
			length := host.Area.Height
			if wall >= 2 {
				length = host.Area.Width
			}
			for offset := 3; offset <= length-span-3; offset++ {
				area, door, edge := secretAttachment(host.Area, width, height, span, wall, offset)
				if !world.ContainsArea(area) || !secretSpaceFree(area, occupied) || !secretEntranceFree(host, door, edge, occupied) {
					continue
				}
				placements = append(placements, secretPlacement{host: host.ID, id: fmt.Sprintf("%s:%d:%d", host.ID, wall, offset), room: area, door: door})
			}
		}
	}
	return placements
}

func secretAttachment(host CellRect, width, height, span, wall, offset int) (CellRect, CellRect, CellRect) {
	switch wall {
	case 0:
		return Cells(host.Right()+1, host.Y+offset-(height-span)/2, width, height), Cells(host.Right(), host.Y+offset, 1, span), Cells(host.Right()-1, host.Y+offset, 1, span)
	case 1:
		return Cells(host.X-width-1, host.Y+offset-(height-span)/2, width, height), Cells(host.X-1, host.Y+offset, 1, span), Cells(host.X, host.Y+offset, 1, span)
	case 2:
		return Cells(host.X+offset-(width-span)/2, host.Y-height-1, width, height), Cells(host.X+offset, host.Y-1, span, 1), Cells(host.X+offset, host.Y, span, 1)
	default:
		return Cells(host.X+offset-(width-span)/2, host.Bottom()+1, width, height), Cells(host.X+offset, host.Bottom(), span, 1), Cells(host.X+offset, host.Bottom()-1, span, 1)
	}
}

func secretSpaceFree(area CellRect, occupied map[[2]int]bool) bool {
	for y := area.Y - 1; y <= area.Bottom(); y++ {
		for x := area.X - 1; x <= area.Right(); x++ {
			if occupied[[2]int{x, y}] {
				return false
			}
		}
	}
	for y := area.Y; y < area.Bottom(); y++ {
		if !occupied[[2]int{area.X - 2, y}] && occupied[[2]int{area.X - 3, y}] || !occupied[[2]int{area.Right() + 1, y}] && occupied[[2]int{area.Right() + 2, y}] {
			return false
		}
	}
	for x := area.X; x < area.Right(); x++ {
		if !occupied[[2]int{x, area.Y - 2}] && occupied[[2]int{x, area.Y - 3}] || !occupied[[2]int{x, area.Bottom() + 1}] && occupied[[2]int{x, area.Bottom() + 2}] {
			return false
		}
	}
	return true
}

func secretEntranceFree(host Room, door, edge CellRect, occupied map[[2]int]bool) bool {
	for y := door.Y - 1; y <= door.Bottom(); y++ {
		for x := door.X - 1; x <= door.Right(); x++ {
			if occupied[[2]int{x, y}] && !host.Area.Contains(At(x, y)) {
				return false
			}
		}
	}
	for y := edge.Y; y < edge.Bottom(); y++ {
		for x := edge.X; x < edge.Right(); x++ {
			for _, prop := range host.Props {
				if host.WorldArea(prop.Area).Contains(At(x, y)) {
					return false
				}
			}
			for _, pool := range host.Pools {
				if host.WorldArea(pool.Area).Contains(At(x, y)) {
					return false
				}
			}
		}
	}
	return true
}

func placeSecret(layout *Layout, room Room, doorID, passageID string, placement secretPlacement) {
	room.Area = placement.room
	layout.SetRoom(room)
	door := placement.door
	layout.SecretDoors = append(layout.SecretDoors, SecretDoor{ID: doorID, RoomID: room.ID, Area: door})
	layout.SetPassage(Passage{ID: passageID, From: Location{At: door.Origin()}, To: Location{At: At(door.Right()-1, door.Bottom()-1)}, Width: 1})
}

func secretRoom(layout Layout, roomID, doorID, passageID, id string, span int) (Layout, string, error) {
	template := DefaultLayout()

	room, _ := template.Room(roomID)
	placements := secretPlacements(layout, room, span)
	if id == "" {
		byRoom := map[string][]secretPlacement{}
		var rooms []string
		for _, placement := range placements {
			if len(byRoom[placement.host]) == 0 {
				rooms = append(rooms, placement.host)
			}
			byRoom[placement.host] = append(byRoom[placement.host], placement)
		}
		if len(rooms) == 0 {
			return layout, "", fmt.Errorf("no valid placement for %s", roomID)
		}
		options := byRoom[rooms[rand.IntN(len(rooms))]]
		placement := options[rand.IntN(len(options))]
		placeSecret(&layout, room, doorID, passageID, placement)
		return layout, placement.id, nil
	}
	for _, placement := range placements {
		if placement.id == id {
			placeSecret(&layout, room, doorID, passageID, placement)
			return layout, id, nil
		}
	}
	return layout, "", fmt.Errorf("invalid placement %q for %s", id, roomID)
}

func dungeonLayout(id string) (Layout, string, error) {
	parts := []string{"", "", "", "", "", ""}
	if id != "" {
		parts = strings.Split(id, "|")
		if len(parts) == 5 {
			parts = append(parts, "0000000000000000")
		}
		if len(parts) != 6 || parts[0] == "" || parts[2] != "astral-gate-v2" || parts[3] == "" || parts[4] == "" || parts[1] != "-" || parts[4] != "-" {
			return Layout{}, "", fmt.Errorf("invalid dungeon layout %q", id)
		}
	}
	if id == "" {
		parts[5] = fmt.Sprintf("%016x", rand.Uint64())
	}
	seed, seedErr := strconv.ParseUint(parts[5], 16, 64)
	if seedErr != nil {
		return Layout{}, "", fmt.Errorf("invalid guard placement seed %q", parts[5])
	}
	layout := secretBase()
	randomizeSideRoomGuards(&layout, seed)
	var err error
	layout, parts[3], err = wingRoom(layout, parts[3])
	if err != nil {
		return Layout{}, "", err
	}

	parts[4] = "-"

	layout, parts[0], err = secretRoom(layout, SecretLibrary, LibraryWall, LibrarySecret, parts[0], 2)
	if err != nil {
		return Layout{}, "", err
	}

	parts[1] = "-"

	parts[2] = "astral-gate-v2"
	return layout, strings.Join(parts, "|"), nil
}

func activateDungeonMap(layout Layout, id string) error {
	m, err := New(layout)
	if err != nil {
		return err
	}
	m.dungeonLayout = id
	dungeonMap.Store(m)
	return nil
}

func SetDungeonLayout(id string) error {
	if Default().LayoutID() == id && id != "" {
		return nil
	}
	if id == "" {
		return fmt.Errorf("missing dungeon layout")
	}
	layout, _, err := dungeonLayout(id)
	if err != nil {
		return err
	}
	return activateDungeonMap(layout, id)
}

func RandomizeDungeon() string {
	layout, id, err := dungeonLayout("")
	if err != nil {
		panic(err)
	}
	if err = activateDungeonMap(layout, id); err != nil {
		panic(err)
	}
	return id
}
