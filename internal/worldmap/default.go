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
		},
		Passages: []Passage{
			{
				ID:    "surface-entrance",
				From:  Location{Room: ProvisionCourt, At: At(13, 32)},
				To:    Location{Room: ProvisionCourt, At: At(13, 56)},
				Width: 6,
			},
			{
				ID:    "sanctum-entrance",
				From:  Location{Room: TeleportSanctum, At: At(4, 12)},
				To:    Location{Room: CrownArena, At: At(7, -1)},
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
		Area:  Cells(9, -72, 18, 8),
		Floor: FirstFloor,
		Roles: []RoomRole{ExplorationRoom, Refuge},
		Props: []Prop{
			Cover("west-veil", Cells(1, 2, 3, 4)),
			Cover("east-veil", Cells(14, 2, 3, 4)),
		},
		Statues:      []Statue{{ID: "column-inscription", At: At(9, 5)}},
		FeedingAreas: []FeedingArea{dungeonEncounter("column-guard", At(9, 2), 1, 3)},
	}
}

func stillwaterCourt() Room {
	return Room{
		ID:    StillwaterCourt,
		Name:  "THE STILLWATER COURT",
		Area:  Cells(0, -92, 36, 20),
		Floor: FirstFloor,
		Roles: []RoomRole{ExplorationRoom},
		Pools: []Pool{
			{ID: "pool-1", Area: Cells(11, 5, 5, 4)},
			{ID: "pool-2", Area: Cells(20, 5, 5, 4)},
			{ID: "pool-3", Area: Cells(11, 13, 5, 4)},
			{ID: "pool-4", Area: Cells(20, 13, 5, 4)},
		},
		Statues: []Statue{{ID: "stillwater-inscription", At: At(5, 10)}},
		FeedingAreas: []FeedingArea{
			dungeonEncounter("stillwater-west", At(5, 5), 3, 5),
			dungeonEncounter("stillwater-east", At(30, 10), 3, 5),
		},
	}
}

func crownArena() Room {
	return Room{
		ID:    CrownArena,
		Name:  "THE CROWN",
		Area:  Cells(9, -112, 18, 20),
		Floor: FirstFloor,
		Roles: []RoomRole{ExplorationRoom, Refuge, BossArena},
		Props: []Prop{
			Cover("west-veil", Cells(1, 10, 3, 4)),
			Cover("east-veil", Cells(14, 10, 3, 4)),
		},
		Statues:      []Statue{{ID: "crown-inscription", At: At(9, 16)}},
		FeedingAreas: []FeedingArea{dungeonEncounter("crown-guard", At(9, 6), 3, 6)},
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
		Area:  Cells(12, -128, 12, 12),
		Floor: FirstFloor,
		Roles: []RoomRole{PortalSanctum},
		Portals: []Portal{
			{ID: DescentPortal, At: At(6, 6), Destination: ReturnPortal},
		},
	}
}

func feastHall() Room {
	return Room{
		ID:    FeastHall,
		Name:  "THE FEAST",
		Area:  Cells(10, -156, 16, 16),
		Floor: SecondFloor,
		Portals: []Portal{
			{ID: ReturnPortal, At: At(8, 11), Destination: DescentPortal},
		},
	}
}
