package worldmap

import (
	"container/heap"
	"math"

	"dungeon-raylib-go/internal/parameters"
	"dungeon-raylib-go/internal/physics"
)

var steps = [][2]int{{0, -1}, {-1, 0}, {1, 0}, {0, 1}}

func (m *Map) Sight(from, to [2]int) bool {
	if from == to {
		return true
	}
	if m.Opaque(from[0], from[1]) || m.Opaque(to[0], to[1]) {
		return false
	}
	return m.Trace(from, to)
}

func (m *Map) Trace(from, to [2]int) bool {
	return traceCells(from, to, m.Opaque)
}

func traceCells(from, to [2]int, opaque func(int, int) bool) bool {
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
			if opaque(x+sx, y) || opaque(x, y+sy) {
				return false
			}
			x += sx
			y += sy
			ix++
			iy++
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
		if opaque(x, y) {
			return false
		}
	}
	return true
}

func WorldCell(x, y float64) [2]int {
	return [2]int{int(math.Floor(x / parameters.Tile)), int(math.Floor(y / parameters.Tile))}
}

func (m *Map) Visible(ax, ay, bx, by float64) bool {
	if ay >= parameters.SurfaceMin && by >= parameters.SurfaceMin {
		return true
	}
	return m.Sight(WorldCell(ax, ay), WorldCell(bx, by))
}

func (m *Map) NextStep(x, y, tx, ty, size float64) (float64, float64, bool) {
	return m.nextStep(x, y, tx, ty, size, m.CanOccupy)
}

func (m *Map) NextStepOutsideCover(x, y, tx, ty, size float64) (float64, float64, bool) {
	return m.nextStep(x, y, tx, ty, size, func(body physics.Rect) bool {
		return m.CanOccupy(body) && !m.OverlapsCover(body)
	})
}

func (m *Map) nextStep(x, y, tx, ty, size float64, canOccupy func(physics.Rect) bool) (float64, float64, bool) {
	start := WorldCell(x+size/2, y+size/2)
	goal := WorldCell(tx+size/2, ty+size/2)
	if start == goal {
		return tx, ty, true
	}
	gx, gy := CellPosition(goal[0], goal[1], size)
	if !canOccupy(physics.Square(gx, gy, size)) {
		return x, y, false
	}
	estimate := func(c [2]int) int { return abs(c[0]-start[0]) + abs(c[1]-start[1]) }
	open := &pathQueue{{cell: goal, cost: estimate(goal)}}
	distance := map[[2]int]int{goal: 0}
	next := map[[2]int][2]int{}
	for open.Len() > 0 && len(distance) < 20000 {
		c := heap.Pop(open).(pathNode).cell
		if c == start {
			v := next[start]
			px, py := CellPosition(v[0], v[1], size)
			return px, py, true
		}
		for _, d := range steps {
			n := [2]int{c[0] + d[0], c[1] + d[1]}
			cost := distance[c] + 1
			if known, ok := distance[n]; ok && known <= cost {
				continue
			}
			px, py := CellPosition(n[0], n[1], size)
			if n != start && !canOccupy(physics.Square(px, py, size)) {
				continue
			}
			distance[n] = cost
			next[n] = c
			heap.Push(open, pathNode{cell: n, cost: cost + estimate(n)})
		}
	}
	return x, y, false
}

type pathNode struct {
	cell [2]int
	cost int
}

type pathQueue []pathNode

func (q pathQueue) Len() int           { return len(q) }
func (q pathQueue) Less(i, j int) bool { return q[i].cost < q[j].cost }
func (q pathQueue) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }
func (q *pathQueue) Push(v any)        { *q = append(*q, v.(pathNode)) }
func (q *pathQueue) Pop() any {
	old := *q
	v := old[len(old)-1]
	*q = old[:len(old)-1]
	return v
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
