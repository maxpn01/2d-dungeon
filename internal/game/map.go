package game

import (
	"image/color"
	"math"
	"raylib-2d-sandbox/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Cell struct {
	size  rl.Vector2
	color color.RGBA
}

type EdgePos struct {
	start float32
	end   float32
}

type Map struct {
	safeZones   []simulation.SafeZone
	exploration mapExploration
	background  color.RGBA
	sizeInCells rl.Vector2
	cellSize    rl.Vector2
	size        rl.Vector2
	edgePosX    EdgePos
	edgePosY    EdgePos
	cells       [][]Cell
}

func NewMap(size, cellSize rl.Vector2, bgColor, accentColor color.RGBA) *Map {
	cells := make([][]Cell, int(size.Y))
	for row := range cells {
		cells[row] = make([]Cell, int(size.X))
	}

	for row := 0; row < int(size.Y); row++ {
		for col := 0; col < int(size.X); col++ {
			c := bgColor
			if isEdgeCell(row, col, int(size.X), int(size.Y)) || isCenterCell(row, col, int(size.X), int(size.Y)) {
				c = accentColor
			}
			cells[row][col] = Cell{
				size:  cellSize,
				color: c,
			}
		}
	}

	return &Map{
		background:  bgColor,
		sizeInCells: size,
		cellSize:    cells[0][0].size,
		size:        rl.NewVector2(size.X*cellSize.X, size.Y*cellSize.Y),
		edgePosX:    EdgePos{start: cellSize.X, end: size.X*cellSize.X - cellSize.X}, // to account for border cells
		edgePosY:    EdgePos{start: cellSize.Y, end: size.Y*cellSize.Y - cellSize.Y}, // to account for border cells
		cells:       cells,
	}
}

func (m *Map) draw(ctx RenderContext) {
	rl.DrawRectangle(0, 0, int32(m.size.X), int32(m.size.Y), m.background)
	colStart, rowStart, colEnd, rowEnd := m.visibleCellRange(ctx.Camera, ctx.WindowWidth, ctx.WindowHeight)
	for row := rowStart; row < rowEnd; row++ {
		for col := colStart; col < colEnd; col++ {
			cell := m.cells[row][col]
			if cell.color == m.background {
				continue
			}
			x := float32(col) * cell.size.X
			y := float32(row) * cell.size.Y
			rl.DrawRectangle(int32(x), int32(y), int32(cell.size.X), int32(cell.size.Y), cell.color)
		}
	}
	drawSafeZones(m.safeZones)
	m.drawDungeon(ctx)
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

func isEdgeCell(row, col, width, height int) bool {
	return row == 0 || row == height-1 || col == 0 || col == width-1
}

func isCenterCell(row, col, width, height int) bool {
	centerRow := height / 2
	centerCol := width / 2
	radius := 2
	return row >= centerRow-radius && row <= centerRow+radius &&
		col >= centerCol-radius && col <= centerCol+radius
}
