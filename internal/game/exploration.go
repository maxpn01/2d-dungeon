package game

import (
	"math"

	"raylib-2d-sandbox/internal/terrain"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const explorationRadius = 400

type mapExploration struct {
	explored map[[2]int]bool
	visible  map[[2]int]bool
	origin   [2]int
}

func explorationCell(p rl.Vector2) [2]int {
	return [2]int{int(math.Floor(float64(p.X) / terrain.Tile)), int(math.Floor(float64(p.Y) / terrain.Tile))}
}

func (e *mapExploration) reveal(p rl.Vector2) {
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
			if dx*dx+dy*dy > radius*radius || x < 0 || x >= 50 || y < terrain.North/terrain.Tile || y >= 50 {
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

func explorationOpaque(x, y int) bool {
	return y < 0 && terrain.Cell(x, y) == 0
}

func explorationSight(from, to [2]int) bool {
	x, y := from[0], from[1]
	dx, dy := to[0]-x, to[1]-y
	nx, ny := int(math.Abs(float64(dx))), int(math.Abs(float64(dy)))
	sx, sy := 1, 1
	if dx < 0 {
		sx = -1
	}
	if dy < 0 {
		sy = -1
	}
	for ix, iy := 0, 0; ix < nx || iy < ny; {
		decision := (1+2*ix)*ny - (1+2*iy)*nx
		if decision == 0 {
			if explorationOpaque(x+sx, y) || explorationOpaque(x, y+sy) {
				return false
			}
			x, y = x+sx, y+sy
			ix, iy = ix+1, iy+1
		} else if decision < 0 {
			x += sx
			ix++
		} else {
			y += sy
			iy++
		}
		if x == to[0] && y == to[1] {
			return true
		}
		if explorationOpaque(x, y) {
			return false
		}
	}
	return true
}
