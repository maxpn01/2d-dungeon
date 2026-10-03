package worldmap

import "raylib-2d-sandbox/internal/physics"

type Exploration struct {
	world    *Map
	radius   int
	explored map[[2]int]bool
	visible  map[[2]int]bool
	origin   [2]int
}

func (m *Map) NewExploration(radius float64) *Exploration {
	cells := 0
	if radius > 0 {
		cells = int(min(radius/Tile, float64(MaxCellX-MinCellX+MaxCellY-MinCellY)))
	}
	return &Exploration{world: m, radius: cells, explored: make(map[[2]int]bool)}
}

func (e *Exploration) SetWorld(world *Map) {
	if e.world != world {
		e.world = world
		e.visible = nil
	}
}

func (e *Exploration) Reveal(position physics.Vector) {
	if !position.Valid() {
		return
	}
	origin := WorldCell(position.X, position.Y)
	if e.visible != nil && origin == e.origin {
		return
	}
	e.origin = origin
	e.visible = make(map[[2]int]bool)
	for y := max(MinCellY, origin[1]-e.radius); y <= min(MaxCellY-1, origin[1]+e.radius); y++ {
		for x := max(MinCellX, origin[0]-e.radius); x <= min(MaxCellX-1, origin[0]+e.radius); x++ {
			dx, dy := x-origin[0], y-origin[1]
			if dx*dx+dy*dy > e.radius*e.radius {
				continue
			}
			cell := [2]int{x, y}
			if e.world.CellFloor(x, y) == e.world.CellFloor(origin[0], origin[1]) && e.world.Trace(origin, cell) {
				e.visible[cell] = true
				e.explored[cell] = true
			}
		}
	}
}

func (e *Exploration) Explored(cell [2]int) bool {
	return e.explored[cell] && e.world.CellFloor(cell[0], cell[1]) == e.world.CellFloor(e.origin[0], e.origin[1])
}
func (e *Exploration) Visible(cell [2]int) bool {
	return e.visible[cell] && e.world.CellFloor(cell[0], cell[1]) == e.world.CellFloor(e.origin[0], e.origin[1])
}
func (e *Exploration) ActorVisible(body physics.Rect) bool {
	center := body.Center()
	cell := WorldCell(center.X, center.Y)
	return e.Visible(cell) && e.world.Sight(e.origin, cell)
}
