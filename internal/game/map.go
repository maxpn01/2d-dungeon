package game

import (
	"math"
	"raylib-2d-sandbox/internal/simulation"
	"raylib-2d-sandbox/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type EdgePos struct {
	start float32
	end   float32
}

type mapView struct {
	gate        simulation.EntranceGateState
	gatePrompt  bool
	safeZones   []simulation.SafeZone
	exploration mapExploration
	sizeInCells rl.Vector2
	cellSize    rl.Vector2
	size        rl.Vector2
	edgePosX    EdgePos
	edgePosY    EdgePos
}

func newMapView() *mapView {
	cellSize := rl.NewVector2(10, 10)
	size := rl.NewVector2(worldmap.SurfaceCells*worldmap.Tile, worldmap.SurfaceCells*worldmap.Tile)
	return &mapView{
		sizeInCells: rl.NewVector2(size.X/cellSize.X, size.Y/cellSize.Y),
		cellSize:    cellSize,
		size:        size,
		edgePosX:    EdgePos{start: cellSize.X, end: size.X - cellSize.X},
		edgePosY:    EdgePos{start: cellSize.Y, end: size.Y - cellSize.Y},
	}
}

func (m *mapView) draw(ctx RenderContext) {
	m.drawSanctuary(ctx)
	drawSafeZones(m.safeZones)
	m.drawSanctuaryAltar(ctx)
	m.drawDungeon(ctx)
}

func (m *mapView) visibleCellRange(camera rl.Camera2D, windowWidth, windowHeight int32) (colStart, rowStart, colEnd, rowEnd int) {
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
