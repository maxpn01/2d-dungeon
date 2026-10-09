package worldmap

import (
	"slices"
	"strings"
)

const MirrorAxis = -160
const NorthSanctuary = "north-sanctuary"

func NorthID(id string) string    { return "north-" + id }
func IsWingStatue(id string) bool { return id == WingStatue || id == NorthID(WingStatue) }

func SpawnCenter(slot uint64) (float64, float64) {
	y := float64(SpawnAltarCenter)
	if slot == 1 {
		y = 2*MirrorAxis*Tile - y
	}
	return SpawnAltarCenter, y
}

func mirrorArea(area CellRect) CellRect {
	area.Y = 2*MirrorAxis - area.Bottom()
	return area
}

func MultiplayerLayout() Layout {
	layout := DefaultLayout()
	original := cloneLayout(layout)
	for _, room := range original.Rooms {
		switch room.ID {
		case TeleportSanctum:
			room.Area = Cells(11, -171, 12, 22)
			room.Portals[0].At = At(6, 11)
			layout.SetRoom(room)
			continue
		case FeastHall:
			room.Area = Cells(70, -350, 24, 24)
			layout.SetRoom(room)
			continue
		case OrbVault, VaultCourt:
			continue
		}
		if room.HasRole(BossArena) {
			room.Roles = slices.DeleteFunc(room.Roles, func(role RoomRole) bool { return role == BossArena })
		}
		room.ID = NorthID(room.ID)
		room.Name = "NORTH / " + room.Name
		room.Area = mirrorArea(room.Area)
		for i := range room.Props {
			room.Props[i].Area.Y = room.Area.Height - room.Props[i].Area.Bottom()
		}
		for i := range room.Pools {
			room.Pools[i].Area.Y = room.Area.Height - room.Pools[i].Area.Bottom()
		}
		for i := range room.FeedingAreas {
			room.FeedingAreas[i].ID = NorthID(room.FeedingAreas[i].ID)
			room.FeedingAreas[i].At.Y = room.Area.Height - 1 - room.FeedingAreas[i].At.Y
		}
		for i := range room.Statues {
			room.Statues[i].ID = NorthID(room.Statues[i].ID)
			room.Statues[i].At.Y = room.Area.Height - 1 - room.Statues[i].At.Y
		}
		for i := range room.Gates {
			room.Gates[i].ID = NorthID(room.Gates[i].ID)
			room.Gates[i].At.Y = room.Area.Height - room.Gates[i].At.Y
		}
		layout.Rooms = append(layout.Rooms, room)
	}
	layout.Rooms = append(layout.Rooms, Room{ID: NorthSanctuary, Name: "THE NORTHERN SANCTUARY", Area: Cells(0, 2*MirrorAxis-SurfaceCells, SurfaceCells, SurfaceCells), Floor: FirstFloor})
	for _, door := range original.SecretDoors {
		door.ID, door.RoomID, door.Area = NorthID(door.ID), NorthID(door.RoomID), mirrorArea(door.Area)
		layout.SecretDoors = append(layout.SecretDoors, door)
	}
	for _, passage := range original.Passages {
		if strings.HasPrefix(passage.ID, "vault-") && passage.ID != "vault-gallery-approach" {
			continue
		}
		area, _ := original.passageArea(passage)
		area = mirrorArea(area)
		from, to, width := At(area.X, area.Y), At(area.Right()-1, area.Y), area.Height
		if passage.From.At.X == passage.To.At.X && passage.From.Room == passage.To.Room || area.Height > area.Width {
			to, width = At(area.X, area.Bottom()-1), area.Width
		}
		layout.Passages = append(layout.Passages, Passage{ID: NorthID(passage.ID), From: Location{At: from}, To: Location{At: to}, Width: width})
	}
	layout.Passages = append(layout.Passages,
		Passage{ID: "north-vault-processional", From: Location{At: At(74, -196)}, To: Location{At: At(74, -164)}, Width: 5},
	)
	layout.SetPassage(Passage{ID: "sanctum-entrance", From: Location{Room: TeleportSanctum, At: At(4, 22)}, To: Location{Room: CrownArena, At: At(10, -1)}, Width: 4})
	layout.SetPassage(Passage{ID: NorthID("sanctum-entrance"), From: Location{Room: NorthID(CrownArena), At: At(10, 30)}, To: Location{Room: TeleportSanctum, At: At(4, -1)}, Width: 4})
	layout.Entrances = append(layout.Entrances, Location{Room: NorthSanctuary, At: At(17, 34)})
	return layout
}
