package worldmap

import "math"

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
			if m.Opaque(x+sx, y) || m.Opaque(x, y+sy) {
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
		if m.Opaque(x, y) {
			return false
		}
	}
	return true
}

func WorldCell(x, y float64) [2]int {
	return [2]int{int(math.Floor(x / Tile)), int(math.Floor(y / Tile))}
}

func (m *Map) Visible(ax, ay, bx, by float64) bool {
	if ay >= SurfaceMin && by >= SurfaceMin {
		return true
	}
	return m.Sight(WorldCell(ax, ay), WorldCell(bx, by))
}

func (m *Map) NextStep(x, y, tx, ty, size float64) (float64, float64, bool) {
	start := WorldCell(x+size/2, y+size/2)
	goal := WorldCell(tx+size/2, ty+size/2)
	if start == goal {
		return tx, ty, true
	}
	queue := [][2]int{goal}
	seen := map[[2]int]bool{goal: true}
	next := map[[2]int][2]int{}
	gx, gy := CellPosition(goal[0], goal[1], size)
	if !m.Walkable(gx, gy, size, size) {
		return x, y, false
	}
	for len(queue) > 0 && len(seen) < 20000 {
		c := queue[0]
		queue = queue[1:]
		for _, d := range steps {
			n := [2]int{c[0] + d[0], c[1] + d[1]}
			if seen[n] {
				continue
			}
			px, py := CellPosition(n[0], n[1], size)
			if !m.Walkable(px, py, size, size) {
				continue
			}
			seen[n] = true
			next[n] = c
			if n == start {
				v := next[start]
				px, py = CellPosition(v[0], v[1], size)
				return px, py, true
			}
			queue = append(queue, n)
		}
	}
	return x, y, false
}
