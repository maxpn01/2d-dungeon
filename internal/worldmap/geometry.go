package worldmap

import (
	"math"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/physics"
)

const Tile = parameters.Tile
const (
	MinCellX     = parameters.MinCellX
	MaxCellX     = parameters.MaxCellX
	MinCellY     = parameters.MinCellY
	SurfaceCells = parameters.SurfaceCells
	SurfaceMin   = parameters.SurfaceMin
	SurfaceMax   = parameters.SurfaceMax
	North        = parameters.North
	ActorSize    = parameters.ActorSize
)

const SpawnAltarCenter = SurfaceCells * Tile / 2
const SpawnAltarRadius = 49

func (m *Map) PrivateGateCrossed(fromX, fromY, toX, toY, size float64) bool {
	gate := m.EntranceGate()
	if gate.Right <= gate.Left {
		return false
	}
	fromCenter, toCenter := fromY+size/2, toY+size/2
	if (fromCenter < float64(gate.Y)) == (toCenter < float64(gate.Y)) {
		return false
	}
	left, right := min(fromX, toX), max(fromX, toX)+size
	return left < float64(gate.Right) && right > float64(gate.Left)
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
	for i, room := range m.layout.Rooms {
		carve(room.Area, i+2)
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

func (m *Map) Cell(x, y int) int { return m.floors[[2]int{x, y}] }
func (m *Map) Bounds() (left, top, right, bottom int) {
	left, right = 0, SurfaceCells
	for c := range m.floors {
		left, top = min(left, c[0]-1), min(top, c[1]-1)
		right = max(right, c[0]+2)
	}
	return left * Tile, top * Tile, right * Tile, SurfaceCells * Tile
}

func (m *Map) Walkable(x, y, w, h float64) bool {
	return m.walkableLayout(x, y, w, h)
}

func (m *Map) walkableLayout(x, y, w, h float64) bool {
	if !(physics.Rect{X: x, Y: y, W: w, H: h}).Valid() || x < MinCellX*Tile || x+w > MaxCellX*Tile || y < North || y+h > SurfaceMax {
		return false
	}
	if y >= SurfaceMin {
		return x >= SurfaceMin && x+w <= SurfaceMax && !overlapsSpawnAltar(x, y, w, h)
	}
	for cy := int(math.Floor(y / Tile)); cy <= int(math.Floor(math.Nextafter((y+h)/Tile, math.Inf(-1)))); cy++ {
		for cx := int(math.Floor(x / Tile)); cx <= int(math.Floor(math.Nextafter((x+w)/Tile, math.Inf(-1)))); cx++ {
			if cy >= 0 {
				if y < SurfaceMin && m.Cell(cx, -1) == 0 {
					return false
				}
				continue
			}
			if m.Cell(cx, cy) == 0 {
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

func overlapsSpawnAltar(x, y, w, h float64) bool {
	return (physics.Rect{X: x, Y: y, W: w, H: h}).OverlapsOctagon(
		physics.Vector{X: SpawnAltarCenter, Y: SpawnAltarCenter}, SpawnAltarRadius,
	)
}
