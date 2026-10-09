package worldmap

import "raylib-2d-sandbox/internal/physics"

type Exploration struct {
	world    *Map
	radius   int
	explored map[[2]int]bool
	visible  map[[2]int]bool
	walls    map[[2]int]bool
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
	e.walls = make(map[[2]int]bool)
	for y := max(MinCellY, origin[1]-e.radius); y <= min(MaxCellY-1, origin[1]+e.radius); y++ {
		for x := max(MinCellX, origin[0]-e.radius); x <= min(MaxCellX-1, origin[0]+e.radius); x++ {
			dx, dy := x-origin[0], y-origin[1]
			if dx*dx+dy*dy > e.radius*e.radius {
				continue
			}
			cell := [2]int{x, y}
			if traceCells(origin, cell, e.world.fogOpaque) {
				e.visible[cell] = true
				e.explored[cell] = true
			}
		}
	}
	for cell := range e.visible {
		if e.world.fogOpaque(cell[0], cell[1]) {
			continue
		}
		for _, offset := range steps {
			wall := [2]int{cell[0] + offset[0], cell[1] + offset[1]}
			if e.world.fogOpaque(wall[0], wall[1]) {
				e.walls[wall] = true
			}
		}
	}
	for wall := range e.walls {
		e.visible[wall] = true
		e.explored[wall] = true
	}
}

func (e *Exploration) WallVisible(cell [2]int) bool { return e.walls[cell] }

func (e *Exploration) Explored(cell [2]int) bool {
	return e.explored[cell] && e.world.CellFloor(cell[0], cell[1]) == e.world.CellFloor(e.origin[0], e.origin[1])
}
func (e *Exploration) Visible(cell [2]int) bool {
	return e.visible[cell]
}
func (e *Exploration) ActorVisible(body physics.Rect) bool {
	center := body.Center()
	cell := WorldCell(center.X, center.Y)
	return e.Visible(cell) && !e.world.fogOpaque(cell[0], cell[1]) && traceCells(e.origin, cell, e.world.fogOpaque)
}
