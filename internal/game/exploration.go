package game

import (
	"math"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/physics"
	"raylib-2d-sandbox/internal/terrain"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const explorationRadius = parameters.ExplorationRadius
const entityDetectionRadius = parameters.EntityDetectionRadius

type mapExploration struct {
	explored map[[2]int]bool
	visible  map[[2]int]bool
	origin   [2]int
	focus    rl.Vector2
}

func explorationCell(p rl.Vector2) [2]int {
	return [2]int{int(math.Floor(float64(p.X) / terrain.Tile)), int(math.Floor(float64(p.Y) / terrain.Tile))}
}

func (e *mapExploration) reveal(p rl.Vector2) {
	e.focus = p
	origin := explorationCell(p)
	if e.explored != nil && origin == e.origin {
		return
	}
	if e.explored == nil {
		e.explored = make(map[[2]int]bool)
	}
	e.origin = origin
	e.visible = make(map[[2]int]bool)
	const radius = explorationRadius / terrain.Tile
	for y := origin[1] - radius; y <= origin[1]+radius; y++ {
		for x := origin[0] - radius; x <= origin[0]+radius; x++ {
			dx, dy := x-origin[0], y-origin[1]
			if (physics.Vector{X: float64(dx), Y: float64(dy)}).LengthSquared() > radius*radius || x < terrain.MinCellX || x >= terrain.MaxCellX || y < terrain.MinCellY || y >= terrain.SurfaceCells {
				continue
			}
			cell := [2]int{x, y}
			if explorationSight(origin, cell) {
				e.visible[cell] = true
				e.explored[cell] = true
			}
		}
	}
}

func explorationSight(from, to [2]int) bool { return terrain.Trace(from, to) }

func (e *mapExploration) actorVisible(x, y, size float64) bool {
	cell := terrain.WorldCell(x+size/2, y+size/2)
	return e.visible[cell] && terrain.Sight(e.origin, cell)
}

func (e *mapExploration) entityVisible(x, y, size float64) bool {
	center := (physics.Rect{X: x, Y: y, W: size, H: size}).Center()
	return center.DistanceSquared(physics.Vector{X: float64(e.focus.X), Y: float64(e.focus.Y)}) <= entityDetectionRadius*entityDetectionRadius
}
