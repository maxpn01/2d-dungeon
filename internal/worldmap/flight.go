package worldmap

import (
	"math"

	"dungeon-raylib-go/internal/parameters"
	"dungeon-raylib-go/internal/physics"
)

func (m *Map) CanFly(body physics.Rect) bool {
	if m.vaultGateClosed {
		room, ok := m.Room(OrbVault)
		if body.Overlaps(m.VaultGateBarrier()) || ok && body.Overlaps(room.Bounds().Expanded(parameters.Tile)) {
			return false
		}
	}
	if !body.Valid() || !insideWorld(body.X, body.Y, body.W, body.H) {
		return false
	}
	flightCell := func(x, y int) int {
		if m.flightCells[[2]int{x, y}] {
			return 1
		}
		return 0
	}

	if body.Overlaps(Cells(0, 0, parameters.SurfaceCells, parameters.SurfaceCells).Bounds()) {
		return walkableSanctuary(body.X, body.Y, body.W, body.H, flightCell, false)
	}
	first := WorldCell(body.X, body.Y)
	last := WorldCell(math.Nextafter(body.X+body.W, math.Inf(-1)), math.Nextafter(body.Y+body.H, math.Inf(-1)))
	for y := first[1]; y <= last[1]; y++ {
		for x := first[0]; x <= last[0]; x++ {
			if !m.flightCells[[2]int{x, y}] || m.solid[[2]int{x, y}] {
				return false
			}
		}
	}
	return true
}

func (m *Map) buildFlightCells() map[[2]int]bool {
	cells := make(map[[2]int]bool)
	reserve := func(area CellRect, margin int) {
		for y := area.Y - margin; y < area.Bottom()+margin; y++ {
			for x := area.X - margin; x < area.Right()+margin; x++ {
				cells[[2]int{x, y}] = true
			}
		}
	}
	reserve(Cells(0, 0, parameters.SurfaceCells, parameters.SurfaceCells), 0)
	for _, room := range m.layout.Rooms {
		margin := 1

		reserve(room.Area, margin)
	}
	for _, passage := range m.passages {
		reserve(passage, 1)
	}
	exterior := make(map[[2]int]bool)
	queue := make([][2]int, 0)
	left, top, right, bottom := parameters.MinCellX-2, parameters.MinCellY-2, parameters.MaxCellX+1, parameters.MaxCellY+1
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
	for y := parameters.MinCellY; y < parameters.MaxCellY; y++ {
		for x := parameters.MinCellX; x < parameters.MaxCellX; x++ {
			cell := [2]int{x, y}
			if !exterior[cell] {
				cells[cell] = true
			}
		}
	}
	return cells
}

func (m *Map) NearestLanding(x, y, size float64) (float64, float64, bool) {
	body := physics.Square(x, y, size)
	if !body.Valid() {
		return x, y, false
	}
	bestX, bestY, bestDistance := x, y, math.Inf(1)
	consider := func(bounds physics.Rect) {
		if bounds.W < size || bounds.H < size {
			return
		}
		candidate := body.ClampedTo(bounds)
		dx, dy := candidate.X-x, candidate.Y-y
		distance := dx*dx + dy*dy
		if distance > bestDistance || distance == bestDistance && (candidate.X > bestX || candidate.X == bestX && candidate.Y >= bestY) {
			return
		}
		if m.CanOccupy(candidate) {
			bestX, bestY, bestDistance = candidate.X, candidate.Y, distance
		}
	}
	consider(physics.Rect{X: parameters.SurfaceMin, Y: parameters.SurfaceMin, W: parameters.SurfaceMax - parameters.SurfaceMin, H: parameters.SurfaceMax - parameters.SurfaceMin})

	considerCell := func(cx, cy int) {
		for height := 1; height <= 2; height++ {
			for width := 1; width <= 2; width++ {
				consider(Cells(cx, cy, width, height).Bounds())
			}
		}
	}
	for cell := range m.floors {
		considerCell(cell[0], cell[1])
	}
	for cy := 0; cy < parameters.SurfaceCells; cy++ {
		for cx := 0; cx < parameters.SurfaceCells; cx++ {
			considerCell(cx, cy)
		}
	}
	return bestX, bestY, !math.IsInf(bestDistance, 1)
}

func (m *Map) FlightLandingStep(x, y, tx, ty, size float64) (float64, float64, bool) {
	air := physics.World{CanOccupy: m.CanFly, MaxStep: parameters.Tile / 5}
	if _, blocked := air.MovementLimit(physics.Square(x, y, size), physics.Vector{X: tx - x, Y: ty - y}); !blocked {
		return tx, ty, true
	}
	return m.nextStep(x, y, tx, ty, size, m.CanFly)
}

func (m *Map) MoveFlying(x, y, dx, dy, size float64) (float64, float64) {
	air := physics.World{CanOccupy: m.CanFly, MaxStep: parameters.Tile / 5}
	motion := air.MoveAndSlide(physics.Square(x, y, size), physics.Vector{X: dx, Y: dy})
	return motion.Bounds.X, motion.Bounds.Y
}
