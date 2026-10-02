package worldmap

import (
	"math"

	"raylib-2d-sandbox/internal/physics"
)

func (m *Map) CanFly(body physics.Rect) bool {
	if !body.Valid() || body.X < MinCellX*Tile || (body.X+body.W) > MaxCellX*Tile || body.Y < North || (body.Y+body.H) > SurfaceMax {
		return false
	}
	if room, ok := m.Room(NorthSanctuary); ok && body.Overlaps(room.Bounds()) {
		return walkableSanctuary(body.X, 2*MirrorAxis*Tile-body.Y-body.H, body.W, body.H, func(cx, cy int) int {
			if m.flightCell(cx, 2*MirrorAxis-cy-1) {
				return 1
			}
			return 0
		}, false)
	}
	if body.Y >= SurfaceMin && body.X >= SurfaceMin && body.X+body.W <= SurfaceMax {
		return true
	}
	for y := int(math.Floor(body.Y / Tile)); y <= int(math.Floor(math.Nextafter((body.Y+body.H)/Tile, math.Inf(-1)))); y++ {
		for x := int(math.Floor(body.X / Tile)); x <= int(math.Floor(math.Nextafter((body.X+body.W)/Tile, math.Inf(-1)))); x++ {
			if y >= 0 && x >= 0 && x < SurfaceCells {
				if body.Y < SurfaceMin && y == 0 && !m.flightCell(x, -1) || body.X < SurfaceMin && x == 0 && !m.flightCell(-1, y) || body.X+body.W > SurfaceMax && x == SurfaceCells-1 && !m.flightCell(SurfaceCells, y) {
					return false
				}
			}
			if !m.flightCell(x, y) {
				return false
			}
		}
	}
	return true
}

func (m *Map) buildFlightCells() map[[2]int]bool {
	cells := make(map[[2]int]bool)
	floor := func(x, y int) bool {
		return (x >= 0 && x < SurfaceCells && y >= 0 && y < SurfaceCells) || m.Cell(x, y) != 0
	}
	left, top, right, bottom := MinCellX-1, MinCellY-1, MaxCellX, SurfaceCells
	for y := top; y <= bottom; y++ {
		for x := left; x <= right; x++ {
			if floor(x, y) || m.Pool(x, y) || (floor(x-1, y) && floor(x+1, y)) || (floor(x, y-1) && floor(x, y+1)) {
				cells[[2]int{x, y}] = true
			}
		}
	}
	exterior := make(map[[2]int]bool)
	queue := make([][2]int, 0)
	add := func(x, y int) {
		cell := [2]int{x, y}
		if x < left || x > right || y < top || y > bottom || cells[cell] || exterior[cell] {
			return
		}
		exterior[cell] = true
		queue = append(queue, cell)
	}
	for x := left; x <= right; x++ {
		add(x, top)
		add(x, bottom)
	}
	for y := top; y <= bottom; y++ {
		add(left, y)
		add(right, y)
	}
	for i := 0; i < len(queue); i++ {
		cell := queue[i]
		for _, step := range steps {
			add(cell[0]+step[0], cell[1]+step[1])
		}
	}
	for y := top; y <= bottom; y++ {
		for x := left; x <= right; x++ {
			cell := [2]int{x, y}
			if !exterior[cell] {
				cells[cell] = true
			}
		}
	}
	return cells
}

func (m *Map) flightCell(x, y int) bool { return m.flightCells[[2]int{x, y}] }

func (m *Map) FlightLandingStep(x, y, tx, ty, size float64) (float64, float64, bool) {
	air := physics.World{CanOccupy: m.CanFly, MaxStep: Tile / 5}
	if _, blocked := air.MovementLimit(physics.Square(x, y, size), physics.Vector{X: tx - x, Y: ty - y}); !blocked {
		return tx, ty, true
	}
	return m.nextStep(x, y, tx, ty, size, m.CanFly)
}

func (m *Map) MoveFlying(x, y, dx, dy, size float64) (float64, float64) {
	air := physics.World{CanOccupy: m.CanFly, MaxStep: Tile / 5}
	motion := air.MoveAndSlide(physics.Square(x, y, size), physics.Vector{X: dx, Y: dy})
	return motion.Bounds.X, motion.Bounds.Y
}
