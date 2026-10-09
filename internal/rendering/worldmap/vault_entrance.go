package worldmap

import (
	"math"
	"raylib-2d-sandbox/internal/physics"
	mapdata "raylib-2d-sandbox/internal/worldmap"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (m *Map) drawVaultApproach(room mapdata.Room, rune rl.Color) {
	start, entry := m.world.VaultApproach()
	points := []mapdata.Cell{entry, mapdata.At(entry.X, start.Y), start, m.world.VaultOrbCell()}
	if entry.X == m.world.VaultOrbCell().X {
		points = []mapdata.Cell{entry, m.world.VaultOrbCell()}
	} else if entry.Y == m.world.VaultOrbCell().Y {
		points = []mapdata.Cell{entry, m.world.VaultOrbCell()}
	}
	bounds := room.Bounds()
	rects := make([]physics.Rect, 0, 3)
	for i := 1; i < len(points); i++ {
		p, q := points[i-1].Center(), points[i].Center()
		if p == q {
			continue
		}
		r := physics.Rect{X: min(p.X, q.X) - 28, Y: min(p.Y, q.Y) - 28, W: math.Abs(p.X-q.X) + 56, H: math.Abs(p.Y-q.Y) + 56}
		left, top := max(r.X, bounds.X), max(r.Y, bounds.Y)
		r = physics.Rect{X: left, Y: top, W: min(r.X+r.W, bounds.X+bounds.W) - left, H: min(r.Y+r.H, bounds.Y+bounds.H) - top}
		rects = append(rects, r)
		rl.DrawRectangleRec(rl.NewRectangle(float32(r.X), float32(r.Y), float32(r.W), float32(r.H)), rl.NewColor(49, 52, 73, 255))
	}
	for i, r := range rects {
		for _, edge := range [][4]float64{{r.X, r.Y, r.X + r.W, r.Y}, {r.X, r.Y + r.H, r.X + r.W, r.Y + r.H}, {r.X, r.Y, r.X, r.Y + r.H}, {r.X + r.W, r.Y, r.X + r.W, r.Y + r.H}} {
			cuts := []float64{0, 1}
			for _, other := range rects {
				lo, hi := other.X, other.X+other.W
				base, span := edge[0], edge[2]-edge[0]
				if span == 0 {
					lo, hi = other.Y, other.Y+other.H
					base, span = edge[1], edge[3]-edge[1]
				}
				cuts = append(cuts, max(0, min(1, (lo-base)/span)), max(0, min(1, (hi-base)/span)))
			}
			slices.Sort(cuts)
			for k := 1; k < len(cuts); k++ {
				from, to := cuts[k-1], cuts[k]
				if from == to {
					continue
				}
				mx, my := edge[0]+(edge[2]-edge[0])*(from+to)/2, edge[1]+(edge[3]-edge[1])*(from+to)/2
				hidden := mx == bounds.X || mx == bounds.X+bounds.W || my == bounds.Y || my == bounds.Y+bounds.H
				for j, other := range rects {
					if i != j && mx > other.X && mx < other.X+other.W && my > other.Y && my < other.Y+other.H {
						hidden = true
						break
					}
				}
				if !hidden {
					rl.DrawLineEx(rl.NewVector2(float32(edge[0]+(edge[2]-edge[0])*from), float32(edge[1]+(edge[3]-edge[1])*from)), rl.NewVector2(float32(edge[0]+(edge[2]-edge[0])*to), float32(edge[1]+(edge[3]-edge[1])*to)), 2, rl.Fade(rune, .4))
				}
			}
		}
	}
}

func (m *Map) drawVaultBorder(room mapdata.Room, rune rl.Color) {
	b := room.Bounds()
	left, right, top, bottom := float32(b.X+18), float32(b.X+b.W-18), float32(b.Y+18), float32(b.Y+b.H-18)
	line := func(x, y, ex, ey float32) {
		rl.DrawLineEx(rl.NewVector2(x, y), rl.NewVector2(ex, ey), 2, rl.Fade(rune, .4))
	}
	gate, ok := m.world.Gate(mapdata.VaultGate)
	if !ok {
		return
	}
	line(left, top, right, top)
	line(left, top, left, bottom)
	line(right, top, right, bottom)
	line(left, bottom, float32(gate.Left)-8, bottom)
	line(float32(gate.Right)+8, bottom, right, bottom)
}
