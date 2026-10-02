package worldmap

import "time"

const (
	ProvisionCourt  = "the-provision-hall"
	GateHall        = "gate-hall"
	ColumnHall      = "the-column-hall"
	StillwaterCourt = "the-stillwater-bridge"
	CrownArena      = "the-crown-boss-arena"
	TeleportSanctum = "the-descent"
	FeastHall       = "floor-two"
	EntranceGate    = "entrance-gate"
	DescentPortal   = "descent"
	ReturnPortal    = "return"
	WestArchive     = "west-archive"
	WestGallery     = "west-gallery"
	EastArmory      = "east-armory"
	EastGallery     = "east-gallery"
	SecretLibrary   = "secret-library"
	OrbVault        = "orb-vault"
	WingChamber     = "wing-chamber"
	WingDoor        = "wing-door"
	WingStatue      = "wing-statue"
)

func DefaultLayout() Layout {
	return Layout{
		Rooms: []Room{
			provisionCourt(),
			gateHall(),
			columnHall(),
			stillwaterCourt(),
			crownArena(),
			teleportSanctum(),
			feastHall(),
			sideRoom(WestArchive, "THE SUNKEN ARCHIVE", Cells(-32, -106, 20, 20), ""),
			sideRoom(WestGallery, "THE WEST GALLERY", Cells(-32, -134, 20, 20), "column-inscription"),
			sideRoom(EastArmory, "THE BROKEN ARMORY", Cells(48, -112, 20, 20), "armory-inscription"),
			sideRoom(EastGallery, "THE EAST GALLERY", Cells(48, -140, 20, 20), ""),
			{ID: SecretLibrary, Name: "THE FORGOTTEN LIBRARY", Area: Cells(-47, -103, 14, 17), Floor: FirstFloor, Statues: []Statue{{ID: "secret-inscription", At: At(6, 6)}}, FeedingAreas: []FeedingArea{{ID: "library-rest", At: At(6, 12), Radius: 2, Food: FoodSource{Capacity: 12, Respawn: 30 * time.Second}}}},
			orbVault(),
			{ID: WingChamber, Name: "THE SEALED AERIE", Area: Cells(-15, SurfaceCells-9, 15, 8), Floor: FirstFloor, Statues: []Statue{{ID: WingStatue, At: At(2, 3)}}},
		},
		SecretDoors: []SecretDoor{
			{ID: WingDoor, RoomID: WingChamber, Area: Cells(-1, SurfaceCells-6, 1, 2), Interactable: true},
			{ID: "library-wall", RoomID: SecretLibrary, Area: Cells(-33, -89, 1, 2)},
			{ID: "vault-wall", RoomID: OrbVault, Area: Cells(68, -131, 1, 3)},
		},
		Passages: []Passage{
			{ID: "aerie-entrance", From: Location{At: At(-1, SurfaceCells-6)}, To: Location{At: At(-1, SurfaceCells-5)}, Width: 1},
			{ID: "archive-approach", From: Location{At: At(-12, -94)}, To: Location{At: At(-6, -94)}, Width: 4},
			{ID: "west-loop", From: Location{At: At(-24, -114)}, To: Location{At: At(-24, -106)}, Width: 4},
			{ID: "west-crown", From: Location{At: At(-12, -124)}, To: Location{At: At(5, -124)}, Width: 4},
			{ID: "armory-approach", From: Location{At: At(41, -98)}, To: Location{At: At(48, -98)}, Width: 4},
			{ID: "east-loop", From: Location{At: At(56, -120)}, To: Location{At: At(56, -112)}, Width: 4},
			{ID: "east-crown", From: Location{At: At(30, -128)}, To: Location{At: At(48, -128)}, Width: 4},
			{ID: "library-secret", From: Location{At: At(-33, -89)}, To: Location{At: At(-32, -89)}, Width: 2},
			{ID: "vault-secret", From: Location{At: At(68, -131)}, To: Location{At: At(69, -131)}, Width: 3},
			{
				ID:    "surface-entrance",
				From:  Location{Room: ProvisionCourt, At: At(13, 32)},
				To:    Location{Room: ProvisionCourt, At: At(13, 56)},
				Width: 6,
			},
			{
				ID:    "sanctum-entrance",
				From:  Location{Room: TeleportSanctum, At: At(6, 18)},
				To:    Location{Room: CrownArena, At: At(10, -1)},
				Width: 4,
			},
		},
		Entrances: []Location{
			{Room: ProvisionCourt, At: At(15, 55)},
		},
	}
}

func provisionCourt() Room {
	return Room{
		ID:    ProvisionCourt,
		Name:  "THE PROVISION COURT",
		Area:  Cells(2, -56, 32, 32),
		Floor: FirstFloor,
		Roles: []RoomRole{Entrance, Private},
		FeedingAreas: []FeedingArea{
			guardedGarden("northwest", At(7, 8), TealGarden),
			guardedGarden("northeast", At(24, 8), VioletGarden),
			guardedGarden("southwest", At(7, 24), VioletGarden),
			guardedGarden("southeast", At(24, 24), TealGarden),
		},
		Gates: []Gate{
			{ID: EntranceGate, At: At(12, 0), Width: 8},
		},
	}
}

func guardedGarden(id string, at Cell, style FeedingStyle) FeedingArea {
	return FeedingArea{
		ID:     id,
		At:     at,
		Radius: 4,
		Style:  style,
		Food:   FoodSource{Capacity: 7, Respawn: 30 * time.Second},
		Guard:  &Guard{PatrolRadius: 4, Level: 1, Respawn: 2 * time.Minute},
	}
}

func gateHall() Room {
	return Room{
		ID:    GateHall,
		Name:  "THE GATE HALL",
		Area:  Cells(14, -64, 8, 8),
		Floor: FirstFloor,
		Roles: []RoomRole{Entrance},
	}
}

func columnHall() Room {
	return Room{
		ID:    ColumnHall,
		Name:  "THE COLUMN HALL",
		Area:  Cells(5, -80, 26, 16),
		Floor: FirstFloor,
		Roles: []RoomRole{ExplorationRoom, Refuge},
		Props: []Prop{
			Cover("west-veil", Cells(3, 5, 3, 4)),
			Cover("east-veil", Cells(20, 5, 3, 4)),
		},
		FeedingAreas: []FeedingArea{dungeonEncounter("column-guard", At(13, 4), 1, 3)},
	}
}

func stillwaterCourt() Room {
	return Room{
		ID:    StillwaterCourt,
		Name:  "THE STILLWATER COURT",
		Area:  Cells(-6, -108, 48, 28),
		Floor: FirstFloor,
		Roles: []RoomRole{ExplorationRoom},
		Pools: []Pool{
			{ID: "pool-1", Area: Cells(14, 6, 5, 4)},
			{ID: "pool-2", Area: Cells(29, 6, 5, 4)},
			{ID: "pool-3", Area: Cells(14, 18, 5, 4)},
			{ID: "pool-4", Area: Cells(29, 18, 5, 4)},
		},
		FeedingAreas: []FeedingArea{
			dungeonEncounter("stillwater-west", At(6, 7), 3, 5),
			dungeonEncounter("stillwater-east", At(41, 18), 3, 5),
		},
	}
}

func crownArena() Room {
	return Room{
		ID:    CrownArena,
		Name:  "THE CROWN",
		Area:  Cells(5, -138, 26, 30),
		Floor: FirstFloor,
		Roles: []RoomRole{ExplorationRoom, Refuge, BossArena},
		Props: []Prop{
			Cover("west-veil", Cells(3, 16, 3, 4)),
			Cover("east-veil", Cells(20, 16, 3, 4)),
		},
		Statues:      []Statue{{ID: "crown-inscription", At: At(13, 4)}},
		FeedingAreas: []FeedingArea{dungeonEncounter("crown-guard", At(13, 10), 3, 6)},
	}
}

func dungeonEncounter(id string, at Cell, radius, food int) FeedingArea {
	return FeedingArea{ID: id, At: at, Radius: max(1, radius), Style: TealGarden,
		Food:  FoodSource{Capacity: food, Respawn: 30 * time.Second},
		Guard: &Guard{PatrolRadius: radius, Level: 2, Damage: 1, AttackRate: 2, Respawn: 45 * time.Second},
	}
}

func teleportSanctum() Room {
	return Room{
		ID:    TeleportSanctum,
		Name:  "THE TELEPORT SANCTUM",
		Area:  Cells(9, -160, 18, 18),
		Floor: FirstFloor,
		Roles: []RoomRole{PortalSanctum},
		Portals: []Portal{
			{ID: DescentPortal, At: At(9, 9), Destination: ReturnPortal},
		},
	}
}

func feastHall() Room {
	return Room{
		ID:   FeastHall,
		Name: "THE FEAST",
		FeedingAreas: []FeedingArea{
			feastFood("feast-northwest", At(5, 5)),
			feastFood("feast-northeast", At(18, 5)),
			feastFood("feast-southwest", At(5, 18)),
			feastFood("feast-southeast", At(18, 18)),
		},
		Area:  Cells(6, -196, 24, 24),
		Floor: SecondFloor,
		Portals: []Portal{
			{ID: ReturnPortal, At: At(12, 17), Destination: DescentPortal},
		},
	}
}

func sideRoom(id, name string, area CellRect, statue string) Room {
	r := Room{ID: id, Name: name, Area: area, Floor: FirstFloor,
		Props:        []Prop{Cover("west-veil", Cells(3, 8, 2, 3)), Cover("east-veil", Cells(15, 8, 2, 3))},
		FeedingAreas: []FeedingArea{dungeonEncounter(id+"-north", At(10, 4), 2, 4), dungeonEncounter(id+"-south", At(10, 15), 2, 4)}}
	switch id {
	case WestArchive:
		r.Props = []Prop{Cover("west-shelves", Cells(3, 4, 2, 12)), Cover("east-shelves", Cells(15, 4, 2, 12))}
	case EastArmory:
		r.Props = []Prop{Cover("north-veil", Cells(3, 3, 2, 2)), Cover("south-veil", Cells(15, 15, 2, 2))}
		r.Pools = []Pool{{ID: "west-basin", Area: Cells(3, 8, 3, 3)}, {ID: "east-basin", Area: Cells(14, 8, 3, 3)}}
	}
	if statue != "" {
		r.Statues = []Statue{{ID: statue, At: At(10, 10)}}
	}
	return r
}

func orbVault() Room {
	r := Room{ID: OrbVault, Name: "THE SEALED ORB VAULT", Area: Cells(69, -140, 18, 20), Floor: FirstFloor,
		Props: []Prop{Cover("north-veil", Cells(6, 3, 2, 2)), Cover("south-veil", Cells(6, 15, 2, 2))}}
	for i, at := range []Cell{At(5, 10), At(12, 4), At(12, 15)} {
		encounter := dungeonEncounter([]string{"vault-sentinel-west", "vault-sentinel-north", "vault-sentinel-south"}[i], at, 2, 0)
		encounter.Guard.Level = 5
		encounter.Guard.Damage = 2
		encounter.Guard.AttackRate = 2.5
		r.FeedingAreas = append(r.FeedingAreas, encounter)
	}
	return r
}

func feastFood(id string, at Cell) FeedingArea {
	return FeedingArea{ID: id, At: at, Radius: 3, Style: VioletGarden, Food: FoodSource{Capacity: 20, Respawn: 20 * time.Second}}
}
