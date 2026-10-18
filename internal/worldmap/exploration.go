package worldmap

import (
	"2d-dungeon/internal/parameters"
	"2d-dungeon/internal/physics"
)

type Exploration struct {
	world    *Map
	radius   int
	flying   bool
	margin   int
	explored map[[2]int]bool
	visible  map[[2]int]bool
	walls    map[[2]int]bool
	origin   [2]int
}

func (m *Map) NewExploration(radius float64) *Exploration {
	cells := 0
	if radius > 0 {
		cells = int(min(radius/parameters.Tile, float64(parameters.MaxCellX-parameters.MinCellX+parameters.MaxCellY-parameters.MinCellY)))
	}
	return &Exploration{world: m, radius: cells, explored: make(map[[2]int]bool)}
}

func (e *Exploration) SetWorld(world *Map) {
	if e.world != world {
		e.world = world
		e.visible = nil
	}
}

func (e *Exploration) SetVision(radius float64, flying bool, margin int) {
	cells := int(radius / parameters.Tile)
	if e.radius != cells || e.flying != flying || e.margin != margin {
		e.radius, e.flying, e.margin = cells, flying, margin
		e.visible = nil
	}
}

func (e *Exploration) Overview() (physics.Rect, bool) {
	area := Cells(-e.margin, -e.margin, parameters.SurfaceCells+2*e.margin, parameters.SurfaceCells+2*e.margin)
	return area.Bounds(), e.flying && area.Contains(At(e.origin[0], e.origin[1]))
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
	left, top, right, bottom := origin[0]-e.radius, origin[1]-e.radius, origin[0]+e.radius, origin[1]+e.radius
	overview, overlooking := e.Overview()
	if overlooking {
		left, top = min(left, -e.margin), min(top, -e.margin)
		right, bottom = max(right, parameters.SurfaceCells+e.margin-1), max(bottom, parameters.SurfaceCells+e.margin-1)
	}
	limitBottom := parameters.MaxCellY - 1
	if overlooking {
		limitBottom = max(limitBottom, parameters.SurfaceCells+e.margin-1)
	}
	for y := max(parameters.MinCellY, top); y <= min(limitBottom, bottom); y++ {
		for x := max(parameters.MinCellX, left); x <= min(parameters.MaxCellX-1, right); x++ {
			dx, dy := x-origin[0], y-origin[1]
			if dx*dx+dy*dy > e.radius*e.radius && !(overlooking && overview.Contains(physics.Square(float64(x*parameters.Tile), float64(y*parameters.Tile), parameters.Tile))) {
				continue
			}
			cell := [2]int{x, y}
			if e.flying || traceCells(origin, cell, e.world.fogOpaque) {
				hiddenFromAbove := e.flying && e.world.SecretHidden(float64(x*parameters.Tile+parameters.Tile/2), float64(y*parameters.Tile+parameters.Tile/2))
				if e.world.CellFloor(x, y) != e.world.CellFloor(origin[0], origin[1]) || hiddenFromAbove {
					continue
				}
				if e.flying && e.world.vaultGateClosed {
					if room, ok := e.world.Room(OrbVault); ok && room.Area.Contains(At(x, y)) {
						continue
					}
				}
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
	return e.Visible(cell) && !e.world.fogOpaque(cell[0], cell[1]) && (e.flying || traceCells(e.origin, cell, e.world.fogOpaque))
}
