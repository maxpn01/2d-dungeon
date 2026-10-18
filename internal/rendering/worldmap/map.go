package worldmap

import (
	"math"

	"dungeon-raylib-go/internal/parameters"
	"dungeon-raylib-go/internal/rendering/primitives"
	"dungeon-raylib-go/internal/simulation"
	mapdata "dungeon-raylib-go/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type edgePos struct {
	start float32
	end   float32
}

type Map struct {
	libraryFrame libraryRenderFrame
	wallHits     map[string]int
	world        *mapdata.Map
	gate         simulation.EntranceGateState
	floor        mapdata.Floor
	gatePrompt   bool
	nearGateID   string
	Exploration  Exploration
	sizeInCells  rl.Vector2
	cellSize     rl.Vector2
	size         rl.Vector2
	edgePosX     edgePos
	edgePosY     edgePos
}

func NewMap() *Map {
	cellSize := rl.NewVector2(10, 10)
	size := rl.NewVector2(parameters.SurfaceCells*parameters.Tile, parameters.SurfaceCells*parameters.Tile)
	return &Map{world: mapdata.Default().WithSecrets(nil),
		sizeInCells: rl.NewVector2(size.X/cellSize.X, size.Y/cellSize.Y),
		cellSize:    cellSize,
		size:        size,
		edgePosX:    edgePos{start: cellSize.X, end: size.X - cellSize.X},
		edgePosY:    edgePos{start: cellSize.Y, end: size.Y - cellSize.Y},
	}
}

func (m *Map) Prepare(snapshot simulation.Snapshot, local simulation.Player) {
	m.Exploration.SetSecrets(snapshot.OpenSecrets, snapshot.Orb.GateOpen)
	m.Exploration.Reveal(rl.NewVector2(float32(local.X+parameters.PlayerSize/2), float32(local.Y+parameters.PlayerSize/2)), local.HasWings, local.Flight.Flying)
	m.world = mapdata.Default().WithSecrets(snapshot.OpenSecrets)
	m.wallHits = snapshot.SecretWallHits
	m.gate = local.EntranceGate
	m.floor = m.world.BodyFloor(local.X, local.Y, parameters.PlayerSize)
	m.gatePrompt = !local.Dead && simulation.NearEntranceGateFor(local.EntranceID, local.X, local.Y)
	m.nearGateID = mapdata.EntranceGate
}

func (m *Map) DrawForestDecals(snapshot simulation.Snapshot, local simulation.Player) {
	drawStatues(snapshot.Statues, local.ActiveStatue, local.X, local.Y, local.Dead, &m.Exploration)
}

func (m *Map) Draw(ctx primitives.Context, snapshot simulation.Snapshot, local simulation.Player) {
	m.Prepare(snapshot, local)
	if m.floor == mapdata.FirstFloor {
		m.drawSanctuary(ctx)
		m.drawSanctuaryAltar(ctx)
	}
	m.drawDungeon(ctx)
	m.drawAstralRoute()
	m.drawVaultGate(snapshot.Orb, local)
	m.drawVaultSeals(snapshot.Orb)
	m.drawVaultTraps(snapshot)
	m.drawHauntedLibrary(snapshot, local)
	m.drawWingDoor(snapshot)
	drawOrbPedestal(snapshot.Orb, local, &m.Exploration)
	drawStatues(snapshot.Statues, local.ActiveStatue, local.X, local.Y, local.Dead, &m.Exploration)
}

func (m *Map) visibleCellRange(camera rl.Camera2D, windowWidth, windowHeight int32) (colStart, rowStart, colEnd, rowEnd int) {
	if camera.Rotation != 0 {
		return 0, 0, int(m.sizeInCells.X), int(m.sizeInCells.Y)
	}

	zoom := camera.Zoom
	if zoom == 0 {
		zoom = 1
	}

	screenW := float32(windowWidth)
	screenH := float32(windowHeight)

	worldMinX := camera.Target.X - camera.Offset.X/zoom
	worldMinY := camera.Target.Y - camera.Offset.Y/zoom
	worldMaxX := camera.Target.X + (screenW-camera.Offset.X)/zoom
	worldMaxY := camera.Target.Y + (screenH-camera.Offset.Y)/zoom

	colStart = clampInt(int(math.Floor(float64(worldMinX/m.cellSize.X))), 0, int(m.sizeInCells.X))
	rowStart = clampInt(int(math.Floor(float64(worldMinY/m.cellSize.Y))), 0, int(m.sizeInCells.Y))
	colEnd = clampInt(int(math.Ceil(float64(worldMaxX/m.cellSize.X))), 0, int(m.sizeInCells.X))
	rowEnd = clampInt(int(math.Ceil(float64(worldMaxY/m.cellSize.Y))), 0, int(m.sizeInCells.Y))

	return colStart, rowStart, colEnd, rowEnd
}
func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func (m *Map) showsRoom(id string) bool {
	room, ok := m.world.Room(id)
	return ok && room.Floor == m.floor
}
