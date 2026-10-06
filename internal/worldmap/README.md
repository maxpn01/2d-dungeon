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
    Guard: &Guard{Type: SpawnGuard, PatrolRadius: 4, Level: 1, Respawn: 2 * time.Minute},
}
```

Food positions use tile centers. `Radius` is the food placement radius in tiles;
`PatrolRadius` creates a square guard patrol around the same tile. Guard `Respawn`
is the delay after death; food `Respawn` is the interval between patch refills. Set `Guard` to
nil to remove the guard; set food capacity to zero to disable its food.
`guardedGarden` holds the shared settings for the four existing gardens.
Simulation compiles these definitions into world-coordinate patches and dens,
with spawning and patrol constraints supplied by their owning room.

Players start with 2 shot damage, 4 shots per second, and 550 world units per
second movement speed. Guards use separate `MobDamageMultiplier`,
`MobAttackRateMultiplier`, and `MobMovementSpeedMultiplier` values from
`internal/parameters/parameters.go`. These multiply the starting stats only.
Both players and guards gain 5 maximum HP, 5 maximum MP, 0.25 damage, and
0.25 shots per second per level. Movement speed stays constant before perks.
Guards spawn and respawn with full HP/MP.

Guard definitions carry a `Type`: `SpawnGuard`, `DungeonShooter`,
`DungeonPursuer`, `DungeonHeavy`, or `Sentinel`. The type is compiled into its
den and retained in simulation stats and multiplayer snapshots. Spawn guards
keep the entrance behavior. Dungeon shooters maintain a 260–350 unit firing distance and strafe with
periodic direction changes, checking walls and territory before moving.
Distance changes use a midpoint dead band, and wall-induced strafe reversals
have a short cooldown to prevent rapid direction switching. Shooter combat
movement stays outside cover; losing sight preserves a short movement
commitment without firing, instead of immediately switching to patrol.
Ordinary direction changes accelerate smoothly, while shot dodges stay reactive. Navigation commits
to grid waypoints and slows for arrival rather than overshooting corners.
Pursuers move faster than an unboosted starting player with the default tuning
and use quick melee strikes. Heavy guards move slowly, have more HP, dodge
less often, and use slower, stronger melee strikes; they stomp toward
targets on a three-second cooldown. During its 0.28-second windup, the guard
tracks the player before locking its aim for the final 0.1 seconds. The stomp
sends a 42-unit-wide shockwave and debris toward the player at 700 units per
second. The wave travels up to 260 units, stops
at walls, and deals damage and a 0.8-second stun when it hits a player. Their square bodies share green coloring with different symbols;
sentinels remain triangular. Role tuning lives in `internal/parameters/parameters.go`.

The central halls contain shooters. Each dungeon initialization shuffles four
side-room encounter pairs: two pursuers, two heavies, a shooter with a heavy,
and a shooter with a pursuer. Each pair's spawn positions are shuffled too.
The encounter seed is stored in the dungeon layout ID so clients reconstruct
the same assignments; mirrored multiplayer wings receive matching encounters.
Guard roles remain fixed across respawns until the next dungeon initialization.
Dungeon melee weapons and their warning accents are green.
Secret-wall runes appear only within `SecretRuneRevealDistance` (90 units) of
the player; inscription text appears within `SecretRuneTextDistance` (65 units).

Entrance guards start at level 1; ordinary dungeon guards start at level 2.
Each guard spawn tracks its own deaths. Every `AIDeathsPerLevel` deaths raises
that guard's next spawn level by one. At each multiple of `AIPerkLevelInterval`
(default 4), it receives a random perk or rank upgrade using the same offers
and rank weights as player statues. Earned perks persist across respawns.
There is no special guard level cap; the shared `MaxLevel` still applies.
A rematch resets guard progression along with the rest of the dungeon.

Vault sentinels are Triangles. When they awaken, their level is the average
ordinary first-floor dungeon guard level, rounded to the nearest whole level,
plus `SentinelLevelOffset` (default 15). The average excludes private entrance
guards, sentinels, and temporary frenzy levels, and includes defeated guards
at their last spawn level. Sentinel levels stay fixed during the encounter.
Each sentinel receives one random perk or rank upgrade at level 14, then
another every four levels (18, 22, and so on), using the same selection as
ordinary guards. Awakening at a higher level grants all reached milestones.
The first threshold is configurable through `SentinelFirstPerkLevel`; the
interval uses `AIPerkLevelInterval`.
Sentinels intercept observed player movement rather than orbiting their current
position. The nearest ready sentinel leads pursuit, with up to
`SentinelConcurrentAttackers` (default 2) attacking together; nearby flankers
can strike without waiting for their turn. They charge across clear ground
for up to `SentinelDashBurst` (default 0.28 seconds), shortening the charge
to stop near their predicted target. Charges have a separate cooldown from
swords, allowing melee follow-ups. `SentinelAttackPrediction` (default 0.75)
controls how much observed movement they anticipate.
Sword strikes have a 0.3-second windup and a short forward arc. Sentinels
advance at `SentinelSwordAdvanceMultiplier` (default 0.8) during windup,
adjusting aim for only the first `SentinelSwordTrackingDuration` (default
0.18 seconds), with turning limited by `SentinelSwordTurnRate` (default 2.5
radians per second). The final 0.12 seconds commit to the telegraphed direction,
followed by stationary swing recovery. Defeating all four unlocks the orb,
and they stay defeated for the match. Frenzy raises ordinary guards to at
least `AIFrenzyLevel` while preserving their HP/MP percentages, then restores
their original level when the hunt ends. Higher-level guards are not weakened.

Taking the orb starts a five-second warning before guardians enter frenzy.
The awakening announcement lasts seven seconds, with a three-second camera
shake and matching low dungeon rumble at pickup. The shake and rumble fade
together; the HUD stays steady and the rumble follows the sound mute setting.

`MobAimAccuracy` controls random angular spread, not a guaranteed hit rate.
The maximum spread in degrees is `(1 - MobAimAccuracy) * 90`. Player shots
remain exact. Killing a mob drops HP pickups totaling 50% of its maximum HP
at spawn, including health perks, subject to the shared loot limit. Temporary
frenzy levels do not increase this reward. MP loot still uses 50% of the mob's
remaining MP.

Player respawn immediately refills their private provision court food patches
and resets that court's guards to fresh HP/MP and positions while preserving
their death counts and earned perks. Loot and shots belonging to those replaced
guards are cleared. The opponent's private route and the shared dungeon keep
their existing state.

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

Dungeon guards respawn after 60 seconds. The base Shift boost multiplier is
1.40625 (25% below its previous 1.875 value), giving a starting player about
773.4 world units per second while boosting.

Inscription choices are cached per player and statue cycle. Reopening a statue
refreshes an offer if it contains perks already owned at the same or higher rank.
Only unowned perks or higher ranks are offered; fully upgraded builds receive no
cards. Recharged statues generate choices when the next player interacts.

Normal guards keep their 400-unit detection radius after being hit. A hit starts
a 10-second investigation toward the shot's recorded origin; seeing the shooter
does not move that destination or extend the timer. Fresh hits can start a new
investigation while the guard is active. Investigations are limited to 900 units
from the guard's home, and private guards remain inside their own court. Reaching
the boundary or exhausting the investigation timer sends the guard home without
reacquiring targets or accepting new alerts until it arrives. Orb-carrier hunts
retain their broader pursuit rules; returning guards may travel back from beyond
the normal leash after such a hunt ends.

Guards aim for 0.4 seconds before their first shot, showing an amber ring. Losing
line of sight cancels an unfinished aim. Once aiming completes, brief cover does
not restart the delay; one second without a visible target resets it. Switching
targets requires a fresh aim. Sustained fire keeps the normal attack rate.

The initial multiplayer lobby
waits for both players to press Start game before showing the briefing over the
loaded dungeon. Both must dismiss it before the shared countdown begins;
the simulation remains paused throughout the briefing and countdown. Disconnecting
during the briefing cancels readiness and returns the remaining player to the lobby.

A round finishes when the first player reaches the next floor. The winner and
round time stay fixed while exploration continues. After three seconds, a results
popup appears over the game. Continue exploring dismisses it; Round results
reopens it later. Both players must request a rematch before the three-second countdown.
A rematch creates a fresh randomized dungeon, resets progression and objectives,
and clears client exploration, effects, and interpolation while preserving player
identities, characters, chat, and input sequence continuity. Leaving during the
countdown cancels it and clears readiness. Solo results offer Play again.
