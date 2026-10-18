- do not waste time on adding tests yet
- do not add comments unless asked

## Project
- roguelike dungeon crawler in Go with raylib (`github.com/gen2brain/raylib-go`). Module name: `dungeon-raylib-go`.
- Run: `go run ./cmd`. Developer playground: `go run ./cmd --testing`. Check that it compiles: `go build ./...`.
- Forest screenshots without playing: `go run ./cmd/forestshot <outDir> <cellX> <cellY> ...`, with cells relative to the forest's top-left corner. `SHOT_HYDRA=<ticks>` screenshots a simulated hydra fight instead (see `docs/v2-forest.md`).

## Code map
- `internal/worldmap`: map layout data (rooms, passages, feeding areas, dens, portals, statues) and floor geometry. `default.go` defines the dungeon.
- `internal/simulation`: game logic that runs at a fixed tick rate (players, mobs, AI, combat, eating, perks, portals).
- `internal/rendering`: drawing and audio. `rendering/worldmap` draws each floor's environment.
- `internal/ui`: HUD, menus and overlays.
- `internal/game`: the main loop that ties simulation, rendering and input together.
- `internal/parameters/parameters.go`: balance numbers. New tunable values go here; reference them as `parameters.X` instead of re-exporting them in other packages.
- `_archive/`: old prototype code, used only as reference. It is not part of the build.

## Current work
- v2 (forest second floor) design, decisions and milestones: `docs/v2-forest.md`. Read it before working on v2, and update its decisions and milestones when they change.
