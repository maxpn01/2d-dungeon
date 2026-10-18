package worldmap

import (
	"time"

	"2d-dungeon/internal/parameters"
)

const (
	ProvisionCourt       = "the-provision-hall"
	GateHall             = "gate-hall"
	ColumnHall           = "the-column-hall"
	StillwaterCourt      = "the-stillwater-bridge"
	CrownArena           = "the-crown-boss-arena"
	TeleportSanctum      = "the-descent"
	Forest               = "floor-two"
	EntranceGate         = "entrance-gate"
	DescentPortal        = "descent"
	ReturnPortal         = "return"
	ForestExitPortal     = "forest-exit"
	EmberArrivalPortal   = "ember-arrival"
	EmberHollow          = "floor-three"
	WestArchive          = "west-archive"
	WestGallery          = "west-gallery"
	EastArmory           = "east-armory"
	EastGallery          = "east-gallery"
	SecretLibrary        = "secret-library"
	OrbVault             = "orb-vault"
	VaultCourt           = "vault-court"
	VaultGate            = "vault-gate"
	WingChamber          = "wing-chamber"
	WingDoor             = "wing-door"
	WingStatue           = "wing-statue"
	SurfaceEntrance      = "surface-entrance"
	VaultProcessional    = "vault-processional"
	VaultCourtEntrance   = "vault-court-entrance"
	VaultGalleryApproach = "vault-gallery-approach"
	AerieEntrance        = "aerie-entrance"
	LibrarySecret        = "library-secret"
	LibraryWall          = "library-wall"
	SecretInscription    = "secret-inscription"
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
			forest(),
			emberHollow(),
			sideRoom(WestArchive, "THE SUNKEN ARCHIVE", Cells(-32, -106, 20, 20), "archive-inscription"),
			sideRoom(WestGallery, "THE WEST GALLERY", Cells(-32, -134, 20, 20), "column-inscription"),
			sideRoom(EastArmory, "THE BROKEN ARMORY", Cells(48, -112, 20, 20), "armory-inscription"),
			sideRoom(EastGallery, "THE EAST GALLERY", Cells(48, -140, 20, 20), ""),
			{ID: SecretLibrary, Name: "THE FORGOTTEN LIBRARY", Area: Cells(-53, -109, 20, 23), Floor: FirstFloor,
				Statues: []Statue{{ID: SecretInscription, At: At(9, 11)}},
				Props: []Prop{
					{ID: "moon-stacks", Kind: BookshelfProp, Area: Cells(4, 5, 7, 1)},
					{ID: "upper-stacks", Kind: BookshelfProp, Area: Cells(13, 3, 1, 5)},
					{ID: "sun-stacks", Kind: BookshelfProp, Area: Cells(14, 12, 1, 6)},
					{ID: "star-stacks", Kind: BookshelfProp, Area: Cells(4, 16, 7, 1)},
				}},
			vaultCourt(),
			orbVault(),
			{ID: WingChamber, Name: "THE SEALED AERIE", Area: Cells(-15, parameters.SurfaceCells-9, 14, 8), Floor: FirstFloor, Statues: []Statue{{ID: WingStatue, At: At(2, 3)}}},
		},
		SecretDoors: []SecretDoor{
			{ID: WingDoor, RoomID: WingChamber, Area: Cells(-1, parameters.SurfaceCells-6, 1, 2), Interactable: true},
			{ID: LibraryWall, RoomID: SecretLibrary, Area: Cells(-33, -89, 1, 2)},
		},
		Passages: []Passage{
			{ID: AerieEntrance, From: Location{At: At(-1, parameters.SurfaceCells-6)}, To: Location{At: At(-1, parameters.SurfaceCells-5)}, Width: 1},
			{ID: "archive-approach", From: Location{At: At(-12, -94)}, To: Location{At: At(-6, -94)}, Width: 4},
			{ID: "west-loop", From: Location{At: At(-24, -114)}, To: Location{At: At(-24, -106)}, Width: 4},
			{ID: "west-crown", From: Location{At: At(-12, -124)}, To: Location{At: At(5, -124)}, Width: 4},
			{ID: "armory-approach", From: Location{At: At(41, -98)}, To: Location{At: At(48, -98)}, Width: 4},
			{ID: "east-loop", From: Location{At: At(56, -120)}, To: Location{At: At(56, -112)}, Width: 4},
			{ID: "east-crown", From: Location{At: At(30, -128)}, To: Location{At: At(48, -128)}, Width: 4},
			{ID: LibrarySecret, From: Location{At: At(-33, -89)}, To: Location{At: At(-32, -89)}, Width: 2},
			{ID: VaultGalleryApproach, From: Location{At: At(68, -128)}, To: Location{At: At(78, -128)}, Width: 5},
			{ID: VaultProcessional, From: Location{At: At(74, -164)}, To: Location{At: At(74, -128)}, Width: 5},
			{ID: VaultCourtEntrance, From: Location{At: At(74, -164)}, To: Location{At: At(80, -164)}, Width: 5},
			{ID: "vault-entrance", From: Location{At: At(90, -178)}, To: Location{At: At(90, -177)}, Width: 7},
			{
				ID:    SurfaceEntrance,
				From:  Location{Room: ProvisionCourt, At: At(13, 32)},
				To:    Location{Room: ProvisionCourt, At: At(13, 56)},
				Width: 6,
			},
			{
				ID:    "sanctum-entrance",
				From:  Location{Room: TeleportSanctum, At: At(4, 18)},
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
		Guard:  &Guard{Type: SpawnGuard, PatrolRadius: 4, Level: 1, Respawn: 2 * time.Minute},
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
		Guard: &Guard{Type: DungeonShooter, PatrolRadius: radius, Level: 2, Respawn: time.Minute},
	}
}

func teleportSanctum() Room {
	return Room{
		ID:    TeleportSanctum,
		Name:  "THE TELEPORT SANCTUM",
		Area:  Cells(11, -160, 12, 18),
		Floor: FirstFloor,
		Roles: []RoomRole{PortalSanctum},
		Portals: []Portal{
			{ID: DescentPortal, At: At(6, 9), Destination: ReturnPortal},
		},
	}
}

func emberHollow() Room {
	return Room{
		ID:    EmberHollow,
		Name:  "THE EMBER HOLLOW",
		Area:  Cells(-4, -252, 36, 18),
		Floor: ThirdFloor,
		Portals: []Portal{
			{ID: EmberArrivalPortal, At: At(18, 14), Destination: ForestExitPortal},
		},
		Terrain: []string{
			"....................................",
			"....................................",
			"....................................",
			"....................................",
			"....................................",
			"....................................",
			"....................................",
			"..............BBBBBBBB..............",
			"....................................",
			"....................................",
			"....................................",
			"......BBBBB..............BBBBB......",
			"....................................",
			"##................................##",
			"###..............................###",
			"####################################",
			"####################################",
			"####################################",
		},
	}
}

func sideRoom(id, name string, area CellRect, statue string) Room {
	r := Room{ID: id, Name: name, Area: area, Floor: FirstFloor, Roles: []RoomRole{ExplorationRoom},
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

func vaultCourt() Room {
	r := Room{ID: VaultCourt, Name: "THE ASTRAL GATE", Area: Cells(80, -177, 27, 34), Floor: FirstFloor, Gates: []Gate{{ID: VaultGate, At: At(10, 0), Width: 7}}}
	for i, at := range []Cell{At(7, 7), At(19, 7)} {
		encounter := dungeonEncounter([]string{"vault-gate-west", "vault-gate-east"}[i], at, 2, 0)
		encounter.Guard.Type = GateHeavy
		encounter.Guard.Level = 10
		r.FeedingAreas = append(r.FeedingAreas, encounter)
	}
	return r
}

func orbVault() Room {
	r := Room{ID: OrbVault, Name: "THE ASTRAL ORB VAULT", Area: Cells(80, -216, 27, 38), Floor: FirstFloor}
	for i, cell := range vaultSentinels(r, At(r.Area.X+r.Area.Width/2, r.Area.Bottom()-3)) {
		at := At(cell.X-r.Area.X, cell.Y-r.Area.Y)
		encounter := dungeonEncounter([]string{"vault-sentinel-west-north", "vault-sentinel-west-south", "vault-sentinel-east-north", "vault-sentinel-east-south"}[i], at, 2, 0)
		encounter.Guard.Type = Sentinel
		encounter.Guard.Level = 2 + parameters.SentinelLevelOffset
		r.FeedingAreas = append(r.FeedingAreas, encounter)
	}
	return r
}
