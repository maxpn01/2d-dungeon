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
	MaxCellY     = parameters.MaxCellY
	SurfaceCells = parameters.SurfaceCells
	SurfaceMin   = parameters.SurfaceMin
	SurfaceMax   = parameters.SurfaceMax
	North        = parameters.North
	ActorSize    = parameters.ActorSize
)

const SpawnAltarCenter = SurfaceCells * Tile / 2
const SpawnAltarRadius = 49

func (m *Map) PrivateGateCrossed(fromX, fromY, toX, toY, size float64) bool {
	for id, gate := range m.gates {
		if id != EntranceGate && id != NorthID(EntranceGate) {
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
	left, right, bottom = 0, SurfaceCells, SurfaceCells
	for c := range m.floors {
		left, top = min(left, c[0]-1), min(top, c[1]-1)
		right = max(right, c[0]+2)
		bottom = max(bottom, c[1]+2)
	}
	return left * Tile, top * Tile, right * Tile, bottom * Tile
}

func (m *Map) Walkable(x, y, w, h float64) bool {
	if m.vaultGateClosed && (physics.Rect{X: x, Y: y, W: w, H: h}).Overlaps(m.VaultGateBarrier()) {
		return false
	}
	return m.walkableLayout(x, y, w, h)
}

func (m *Map) walkableLayout(x, y, w, h float64) bool {
	if !(physics.Rect{X: x, Y: y, W: w, H: h}).Valid() || x < MinCellX*Tile || x+w > MaxCellX*Tile || y < North || y+h > MaxCellY*Tile {
		return false
	}
	body := physics.Rect{X: x, Y: y, W: w, H: h}
	if room, ok := m.Room(NorthSanctuary); ok && body.Overlaps(room.Bounds()) {
		return walkableSanctuary(x, 2*MirrorAxis*Tile-y-h, w, h, func(cx, cy int) int { return m.Cell(cx, 2*MirrorAxis-cy-1) }, true)
	}
	if body.Overlaps(physics.Rect{X: 0, Y: 0, W: SurfaceCells * Tile, H: SurfaceCells * Tile}) {
		return walkableSanctuary(x, y, w, h, m.Cell, true)
	}
	for cy := int(math.Floor(y / Tile)); cy <= int(math.Floor(math.Nextafter((y+h)/Tile, math.Inf(-1)))); cy++ {
		for cx := int(math.Floor(x / Tile)); cx <= int(math.Floor(math.Nextafter((x+w)/Tile, math.Inf(-1)))); cx++ {
			if m.Cell(cx, cy) == 0 {
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
	if y >= SurfaceMin && y+h <= SurfaceMax && x >= SurfaceMin && x+w <= SurfaceMax {
		return true
	}
	for cy := int(math.Floor(y / Tile)); cy <= int(math.Floor(math.Nextafter((y+h)/Tile, math.Inf(-1)))); cy++ {
		for cx := int(math.Floor(x / Tile)); cx <= int(math.Floor(math.Nextafter((x+w)/Tile, math.Inf(-1)))); cx++ {
			if cy >= 0 && cy < SurfaceCells && cx >= 0 && cx < SurfaceCells {
				if y < SurfaceMin && cy == 0 && cell(cx, -1) == 0 || x < SurfaceMin && cx == 0 && cell(-1, cy) == 0 || x+w > SurfaceMax && cx == SurfaceCells-1 && cell(SurfaceCells, cy) == 0 || y+h > SurfaceMax && cy == SurfaceCells-1 && cell(cx, SurfaceCells) == 0 {
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

func overlapsSpawnAltar(x, y, w, h float64) bool {
	return (physics.Rect{X: x, Y: y, W: w, H: h}).OverlapsOctagon(
		physics.Vector{X: SpawnAltarCenter, Y: SpawnAltarCenter}, SpawnAltarRadius,
	)
}
