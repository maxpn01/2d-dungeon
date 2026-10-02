package worldmap

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
)

type secretPlacement struct {
	host string
	id   string
	room CellRect
	door CellRect
}

func secretBase(multiplayer bool) Layout {
	layout := DefaultLayout()
	if multiplayer {
		layout = MultiplayerLayout()
	}
	layout.RemoveRoom(OrbVault)
	layout.RemoveRoom(SecretLibrary)
	layout.RemoveRoom(NorthID(SecretLibrary))
	layout.RemovePassage("vault-secret")
	layout.RemovePassage("library-secret")
	layout.RemovePassage(NorthID("library-secret"))
	layout.SecretDoors = slices.DeleteFunc(layout.SecretDoors, func(door SecretDoor) bool {
		return door.RoomID == OrbVault || door.RoomID == SecretLibrary || door.RoomID == NorthID(SecretLibrary)
	})
	return layout
}

func secretPlacements(layout Layout, room Room, span int, side string) []secretPlacement {
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
	reserve(Cells(0, 0, SurfaceCells, SurfaceCells))
	var placements []secretPlacement
	world := Cells(MinCellX+1, MinCellY+1, MaxCellX-MinCellX-2, SurfaceCells-MinCellY-2)
	for _, host := range layout.Rooms {
		if !host.HasRole(ExplorationRoom) || host.HasRole(Private) || host.Floor != FirstFloor {
			continue
		}
		north := strings.HasPrefix(host.ID, "north-")
		if side == "south" && north || side == "north" && !north {
			continue
		}
		for wall := range 4 {
			length := host.Area.Height
			if wall >= 2 {
				length = host.Area.Width
			}
			for offset := 3; offset <= length-span-3; offset++ {
				area, door, edge := secretAttachment(host.Area, room.Area.Width, room.Area.Height, span, wall, offset)
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

func secretRoom(layout Layout, roomID, doorID, passageID, side, id string, span int) (Layout, string, error) {
	template := DefaultLayout()
	if side == "north" {
		template = MultiplayerLayout()
	}
	room, _ := template.Room(roomID)
	placements := secretPlacements(layout, room, span, side)
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

func dungeonLayout(multiplayer bool, id string) (Layout, string, error) {
	parts := []string{"", "", ""}
	if id != "" {
		parts = strings.Split(id, "|")
		if len(parts) != 3 || parts[0] == "" || parts[2] == "" || multiplayer && (parts[1] == "" || parts[1] == "-") || !multiplayer && parts[1] != "-" {
			return Layout{}, "", fmt.Errorf("invalid dungeon layout %q", id)
		}
	}
	layout := secretBase(multiplayer)
	var err error
	layout, parts[0], err = secretRoom(layout, SecretLibrary, "library-wall", "library-secret", "south", parts[0], 2)
	if err != nil {
		return Layout{}, "", err
	}
	if multiplayer {
		layout, parts[1], err = secretRoom(layout, NorthID(SecretLibrary), NorthID("library-wall"), NorthID("library-secret"), "north", parts[1], 2)
		if err != nil {
			return Layout{}, "", err
		}
	} else {
		parts[1] = "-"
	}
	layout, parts[2], err = secretRoom(layout, OrbVault, "vault-wall", "vault-secret", "", parts[2], 3)
	if err != nil {
		return Layout{}, "", err
	}
	return layout, strings.Join(parts, "|"), nil
}

func activateDungeonMap(layout Layout, id string) error {
	m, err := New(layout)
	if err != nil {
		return err
	}
	m.dungeonLayout = id
	if Multiplayer() {
		multiplayerDungeonMap.Store(m)
	} else {
		soloDungeonMap.Store(m)
	}
	return nil
}

func SetDungeonLayout(id string) error {
	if Default().LayoutID() == id && id != "" {
		return nil
	}
	if id == "" {
		return fmt.Errorf("missing dungeon layout")
	}
	layout, _, err := dungeonLayout(Multiplayer(), id)
	if err != nil {
		return err
	}
	return activateDungeonMap(layout, id)
}

func RandomizeDungeon() string {
	layout, id, err := dungeonLayout(Multiplayer(), "")
	if err != nil {
		panic(err)
	}
	if err = activateDungeonMap(layout, id); err != nil {
		panic(err)
	}
	return id
}
