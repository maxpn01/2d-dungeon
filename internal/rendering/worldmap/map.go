package worldmap

import (
	"math"

	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/rendering/primitives"
	"raylib-2d-sandbox/internal/simulation"
	mapdata "raylib-2d-sandbox/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type edgePos struct {
	start float32
	end   float32
}

type Map struct {
	wallHits    map[string]int
	world       *mapdata.Map
	gate        simulation.EntranceGateState
	floor       mapdata.Floor
	gatePrompt  bool
	nearGateID  string
	safeZones   []simulation.SafeZone
	Exploration Exploration
	sizeInCells rl.Vector2
	cellSize    rl.Vector2
	size        rl.Vector2
	edgePosX    edgePos
	edgePosY    edgePos
}

func NewMap() *Map {
	cellSize := rl.NewVector2(10, 10)
	size := rl.NewVector2(mapdata.SurfaceCells*mapdata.Tile, mapdata.SurfaceCells*mapdata.Tile)
	return &Map{world: mapdata.Default().WithSecrets(nil),
		sizeInCells: rl.NewVector2(size.X/cellSize.X, size.Y/cellSize.Y),
		cellSize:    cellSize,
		size:        size,
		edgePosX:    edgePos{start: cellSize.X, end: size.X - cellSize.X},
		edgePosY:    edgePos{start: cellSize.Y, end: size.Y - cellSize.Y},
	}
}

func (m *Map) Draw(ctx primitives.Context, snapshot protocol.Snapshot, local protocol.Player) {
	m.world = mapdata.Default().WithSecrets(snapshot.OpenSecrets)
	m.wallHits = snapshot.SecretWallHits
	m.safeZones, m.gate = snapshot.SafeZones, local.EntranceGate
	m.floor = m.world.FloorAt(local.X+local.Size/2, local.Y+local.Size/2)
	m.gatePrompt = !local.Dead && simulation.NearEntranceGateFor(local.EntranceID, local.X, local.Y)
	m.nearGateID = mapdata.EntranceGate
	north, _ := mapdata.Default().Gate(mapdata.NorthID(mapdata.EntranceGate))
	if math.Abs(local.Y+local.Size/2-float64(north.Y)) <= 110 {
		m.nearGateID = mapdata.NorthID(mapdata.EntranceGate)
	}
	if m.floor == mapdata.FirstFloor {
		m.drawSanctuary(ctx)
	}
	drawSafeZones(m.safeZones)
	if m.floor == mapdata.FirstFloor {
		m.drawSanctuaryAltar(ctx)
	}
	m.drawDungeon(ctx)
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

func (m *Map) sanctuaryTile(x, y int, northBorder mapdata.CellRect) bool {
	if m.world.SanctuaryCell(x, y) {
		return true
	}
	if !northBorder.Contains(mapdata.At(x, y)) || m.world.Cell(x, y) != 0 {
		return false
	}
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if m.world.Cell(x+dx, y+dy) != 0 && !m.world.SanctuaryCell(x+dx, y+dy) {
				return false
			}
		}
	}
	return true
}
