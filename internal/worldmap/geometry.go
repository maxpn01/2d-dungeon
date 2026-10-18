package worldmap

import (
	"math"

	"dungeon-raylib-go/internal/parameters"
	"dungeon-raylib-go/internal/physics"
)

const SpawnAltarCenter = parameters.SurfaceCells * parameters.Tile / 2
const SpawnAltarRadius = 49

func (m *Map) PrivateGateCrossed(fromX, fromY, toX, toY, size float64) bool {
	for id, gate := range m.gates {
		if id != EntranceGate {
			continue
		}
		fromCenter, toCenter := fromY+size/2, toY+size/2
		if (fromCenter < float64(gate.Y)) == (toCenter < float64(gate.Y)) {
			continue
		}
		left, right := min(fromX, toX), max(fromX, toX)+size
		if left < float64(gate.Right) && right > float64(gate.Left) {
			return true
		}
	}
	return false
}

func (m *Map) Pool(x, y int) bool {
	for _, pool := range m.pools {
		if pool.Area.Contains(At(x, y)) {
			return true
		}
	}
	return false
}

func (m *Map) build() map[[2]int]int {
	floors := make(map[[2]int]int)
	carve := func(area CellRect, id int) {
		for y := area.Y; y < area.Bottom(); y++ {
			for x := area.X; x < area.Right(); x++ {
				floors[[2]int{x, y}] = id
			}
		}
	}
	for _, area := range m.passages {
		carve(area, 1)
	}
	m.terrain = make(map[[2]int]Terrain)
	for i, room := range m.layout.Rooms {
		if len(room.Terrain) == 0 {
			carve(room.Area, i+2)
			continue
		}
		for y, row := range room.Terrain {
			for x := range len(row) {
				kind := Terrain(row[x])
				if kind == Rock {
					continue
				}
				cell := [2]int{room.Area.X + x, room.Area.Y + y}
				floors[cell] = i + 2
				m.terrain[cell] = kind
			}
		}
	}
	for _, pool := range m.pools {
		for y := pool.Area.Y; y < pool.Area.Bottom(); y++ {
			for x := pool.Area.X; x < pool.Area.Right(); x++ {
				delete(floors, [2]int{x, y})
			}
		}
	}
	return floors
}

func (m *Map) Cell(x, y int) int           { return m.floors[[2]int{x, y}] }
func (m *Map) Terrain(x, y int) Terrain    { return m.terrain[[2]int{x, y}] }
func (m *Map) TerrainBlocks(x, y int) bool { return m.terrain[[2]int{x, y}].Blocks() }
func (m *Map) Bounds() (left, top, right, bottom int) {
	left, right, bottom = 0, parameters.SurfaceCells, parameters.SurfaceCells
	for c := range m.floors {
		left, top = min(left, c[0]-1), min(top, c[1]-1)
		right = max(right, c[0]+2)
		bottom = max(bottom, c[1]+2)
	}
	return left * parameters.Tile, top * parameters.Tile, right * parameters.Tile, bottom * parameters.Tile
}

func (m *Map) Walkable(x, y, w, h float64) bool {
	if m.vaultGateClosed && (physics.Rect{X: x, Y: y, W: w, H: h}).Overlaps(m.VaultGateBarrier()) {
		return false
	}
	return m.walkableLayout(x, y, w, h)
}

func (m *Map) ShotPassable(x, y, w, h float64) bool {
	if m.vaultGateClosed && (physics.Rect{X: x, Y: y, W: w, H: h}).Overlaps(m.VaultGateBarrier()) {
		return false
	}
	return m.passable(x, y, w, h, Terrain.StopsShots)
}

func (m *Map) walkableLayout(x, y, w, h float64) bool {
	return m.passable(x, y, w, h, Terrain.Blocks)
}

func (m *Map) passable(x, y, w, h float64, blocks func(Terrain) bool) bool {
	if !(physics.Rect{X: x, Y: y, W: w, H: h}).Valid() || !insideWorld(x, y, w, h) {
		return false
	}
	body := physics.Rect{X: x, Y: y, W: w, H: h}

	if body.Overlaps(physics.Rect{X: 0, Y: 0, W: parameters.SurfaceCells * parameters.Tile, H: parameters.SurfaceCells * parameters.Tile}) {
		return walkableSanctuary(x, y, w, h, m.Cell, true)
	}
	for cy := int(math.Floor(y / parameters.Tile)); cy <= int(math.Floor(math.Nextafter((y+h)/parameters.Tile, math.Inf(-1)))); cy++ {
		for cx := int(math.Floor(x / parameters.Tile)); cx <= int(math.Floor(math.Nextafter((x+w)/parameters.Tile, math.Inf(-1)))); cx++ {
			if m.Cell(cx, cy) == 0 || m.solid[[2]int{cx, cy}] || blocks(m.terrain[[2]int{cx, cy}]) {
				return false
			}
		}
	}
	return true
}

func walkableSanctuary(x, y, w, h float64, cell func(int, int) int, altar bool) bool {
	if altar && overlapsSpawnAltar(x, y, w, h) {
		return false
	}
	if y >= parameters.SurfaceMin && y+h <= parameters.SurfaceMax && x >= parameters.SurfaceMin && x+w <= parameters.SurfaceMax {
		return true
	}
	for cy := int(math.Floor(y / parameters.Tile)); cy <= int(math.Floor(math.Nextafter((y+h)/parameters.Tile, math.Inf(-1)))); cy++ {
		for cx := int(math.Floor(x / parameters.Tile)); cx <= int(math.Floor(math.Nextafter((x+w)/parameters.Tile, math.Inf(-1)))); cx++ {
			if cy >= 0 && cy < parameters.SurfaceCells && cx >= 0 && cx < parameters.SurfaceCells {
				last := parameters.SurfaceCells - 1
				top := y < parameters.SurfaceMin && cy == 0 && cell(cx, -1) == 0
				left := x < parameters.SurfaceMin && cx == 0 && cell(-1, cy) == 0
				right := x+w > parameters.SurfaceMax && cx == last && cell(parameters.SurfaceCells, cy) == 0
				bottom := y+h > parameters.SurfaceMax && cy == last && cell(cx, parameters.SurfaceCells) == 0
				if top || left || right || bottom {
					return false
				}
			} else if cell(cx, cy) == 0 {
				return false
			}
		}
	}
	return true
}

func (m *Map) Move(x, y, dx, dy, w, h float64) (float64, float64) {
	motion := m.MoveBody(physics.Rect{X: x, Y: y, W: w, H: h}, physics.Vector{X: dx, Y: dy})
	return motion.Bounds.X, motion.Bounds.Y
}

func (m *Map) MoveBody(body physics.Rect, delta physics.Vector) physics.Motion {
	return m.collisionWorld.MoveAndSlide(body, delta)
}

func (m *Map) MovementLimit(x, y, dx, dy, w, h float64) (float64, bool) {
	return m.collisionWorld.MovementLimit(physics.Rect{X: x, Y: y, W: w, H: h}, physics.Vector{X: dx, Y: dy})
}

func (m *Map) ShotLimit(x, y, dx, dy, w, h float64) (float64, bool) {
	return m.shotWorld.MovementLimit(physics.Rect{X: x, Y: y, W: w, H: h}, physics.Vector{X: dx, Y: dy})
}

func overlapsSpawnAltar(x, y, w, h float64) bool {
	return (physics.Rect{X: x, Y: y, W: w, H: h}).OverlapsOctagon(
		physics.Vector{X: SpawnAltarCenter, Y: SpawnAltarCenter}, SpawnAltarRadius,
	)
}

func insideWorld(x, y, w, h float64) bool {
	return x >= parameters.MinCellX*parameters.Tile && x+w <= parameters.MaxCellX*parameters.Tile && y >= parameters.North && y+h <= parameters.MaxCellY*parameters.Tile
}
