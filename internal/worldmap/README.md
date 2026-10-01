# Editing the map

Start in `default.go`. `DefaultLayout()` lists the rooms and connections in map
order. Each room has its own small function containing its name, bounds, roles,
and contents. The client and server build this definition at startup; restart
both after editing it.

## Coordinates

`Cells(x, y, width, height)` describes a rectangle in tiles, with exclusive right
and bottom edges. `At(x, y)` describes a tile position. One tile is `Tile` world
units. Negative Y is north of the surface.

Only `Room.Area` is in global map coordinates. Everything inside a room is
relative to its top-left corner: props, pools, feeding areas, gates, and portals.
Moving `Room.Area` moves all of those contents automatically.

```go
func columnHall() Room {
    return Room{
        ID: ColumnHall,
        Name: "THE COLUMN HALL",
        Area: Cells(9, -72, 18, 8),
        Floor: FirstFloor,
        Roles: []RoomRole{ExplorationRoom, Refuge},
        Props: []Prop{
            Cover("west-veil", Cells(1, 2, 3, 4)),
            Cover("east-veil", Cells(14, 2, 3, 4)),
        },
    }
}
```

A cover blocks sight and allows movement. A pool removes floor tiles; its drawing
scales to its authored dimensions. Contents must fit inside their room.
`Roles` use named values instead of arbitrary strings. `BossArena` supplies the
simulation's arena bounds; `PortalSanctum` selects its decorative treatment.
`Floor` supplies floor detection and portal travel labels.

## Food and guards

A `FeedingArea` defines one position shared by food and its optional guard:

```go
FeedingArea{
    ID: "west-garden",
    At: At(7, 8),
    Radius: 4,
    Style: TealGarden,
    Food: FoodSource{Capacity: 7, Respawn: 30 * time.Second},
    Guard: &Guard{PatrolRadius: 4, Level: 1, Respawn: 2 * time.Minute},
}
```

Food positions use tile centers. `Radius` is the food placement radius in tiles;
`PatrolRadius` creates a square guard patrol around the same tile. Guard `Respawn`
is the delay after death; food `Respawn` is the interval between patch refills. Set `Guard` to
nil to remove the guard; set food capacity to zero to disable its food.
`guardedGarden` holds the shared settings for the four existing gardens.
Simulation compiles these definitions into world-coordinate patches and dens,
with spawning and patrol constraints supplied by their owning room.

Guards also accept `Damage` and `AttackRate` (shots per second). Zero uses the
default mob value: 1 damage and 2 shots per second. The first-floor dungeon has
four additional level-2 Square guards and nearby replenishing food patches.
HP and MP are derived from level with the same +5 capacity growth as players.
Guards spawn and respawn with full HP/MP: level 1 has 10/10, level 2 has 15/15.
Entrance guards are level 1; dungeon guards past the gate are level 2.

## Inscription statues

Add room-relative statue positions to a room:

```go
Statues: []Statue{
    {ID: "column-inscription", At: At(9, 5)},
},
```

Statue IDs are unique across the map. Positions must fit in their room and be
reachable. Statues are decorative, walkable objects; interact within 100 world
units with a clear line of sight. `Map.Statues()` returns compiled positions.
Rooms support `SetStatue` and `RemoveStatue` alongside the other content helpers.

The simulation owns shared offers, claims, and reward cycles. A statue becomes
unavailable to everyone when one player selects a card. That player's death or
departure releases their claimed statues, allowing a fresh offer for either
player. Respawn preserves the rest of the world.

Perk tuning lives in `internal/parameters/parameters.go`: `PerkRankWeights`,
`PerkBonuses`, and `PerkPenalties`. Rank columns are E, D, C, B, A, S. Family rows
are damage, attack rate, MP capacity, HP capacity, movement speed, Shift boost,
Heavy Hand, and Overdrive. Percentage values are fractions of the base stat;
HP/MP capacity values are flat bonuses. Each rank replaces the previous rank's
effect rather than adding another copy. Heavy Hand adds damage and subtracts
movement speed; Overdrive adds boost strength and increases boost MP cost.

## Connections, gates, and portals

Passages specify `From`, `To`, and `Width`. Each endpoint is a `Location` with a
room ID and a room-relative `At`; omit the room ID for an absolute map position.
Endpoints can sit outside a room to connect its boundary. They must align on
one axis. Both endpoint tiles are included; width extends right for a vertical
passage and down for a horizontal one. Room moves update referenced endpoints;
if they no longer align, validation asks you to adjust the connection.

Gates use a room-relative tile-boundary position and a width in tiles. The game
currently operates one interactive gate, identified by `EntranceGate`.

Portals use a room-relative tile-boundary position and a destination portal ID.
Moving a portal moves its drawing and teleport destination. Adding a linked
floor does not require manual reachability seeds: validation follows portal
links from `Layout.Entrances`. Each entrance is a named-room `Location` or an
absolute location.

Room IDs are unique across rooms. Feeding area, gate, and portal IDs are unique
across the map; prop and pool IDs are unique within their room. Passage IDs are
unique across passages. Named IDs for existing landmarks live at the top of
`default.go`.

## Editing through the API

Directly editing the room functions is the easiest way to change the game map.
For programmatic changes:

```go
layout := worldmap.DefaultLayout()
room, ok := layout.Room(worldmap.ColumnHall)
if !ok {
    return fmt.Errorf("missing column hall")
}
room.SetProp(worldmap.Cover("middle-veil", worldmap.Cells(7, 2, 3, 4)))
room.RemoveProp("west-veil")
layout.SetRoom(room)
world, err := worldmap.New(layout)
if err != nil {
    return err
}
```

Rooms support `Set`/`Remove` for props, pools, feeding areas, gates, and portals.
Layouts support `Set`/`Remove` for rooms and passages. Set replaces an existing
ID or appends a new item. Remove reports whether an item existed. Removing a
room removes its owned contents; update any passages, entrances, or portal
links that reference it as well.

`New` copies the layout and validates bounds, IDs, roles, settings, references,
portal landings, and reachability. Built maps are read-only. `Layout()` and the
collection queries return independent copies. A newly built map is a separate
instance; it does not replace the running default or synchronize over multiplayer.

## Runtime queries

`Cell`, `Cover`, `Pool`, and `Opaque` query tile coordinates. `Room`, `RoomAt`,
`RoomWithRole`, `RoomContains`, `FloorAt`, `Gate`, `Passage`, and `Portal` query
compiled geometry. `Pools()`, `Patches()`, `Dens()`, and `Portals()` return world
coordinates; `Rooms()` returns the editable room definitions with local contents.

`CanOccupy`, `Walkable`, `MoveBody`, `Move`, and `MovementLimit` use physics to
check placement and movement. `Visible`, `Sight`, `Trace`, and `NextStep` supply
visibility and pathfinding. `NewExploration(radius)` creates independent
exploration state; use `Reveal`, `Explored`, `Visible`, and `ActorVisible` on it.

Tile size, surface bounds, the spawn altar, and general combat tuning remain
outside the editable dungeon layout. The game supports one boss arena and one
interactive entrance gate. Their visual styles and encounter behavior remain in
`game` and `simulation`.
