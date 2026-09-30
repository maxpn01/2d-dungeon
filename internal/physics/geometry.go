package physics

import "math"

type Rect struct {
	X, Y, W, H float64
}

func Square(x, y, size float64) Rect {
	return Rect{X: x, Y: y, W: size, H: size}
}

func (r Rect) Valid() bool {
	return finite(r.X) && finite(r.Y) && finite(r.W) && finite(r.H) && r.W > 0 && r.H > 0 && finite(r.X+r.W) && finite(r.Y+r.H)
}

func (r Rect) Translated(delta Vector) Rect {
	r.X += delta.X
	r.Y += delta.Y
	return r
}

func (r Rect) Center() Vector {
	return Vector{X: r.X + r.W/2, Y: r.Y + r.H/2}
}

func (r Rect) Expanded(margin float64) Rect {
	return Rect{X: r.X - margin, Y: r.Y - margin, W: r.W + 2*margin, H: r.H + 2*margin}
}

func (r Rect) Contains(other Rect) bool {
	return r.Valid() && other.Valid() && other.X >= r.X && other.Y >= r.Y && other.X+other.W <= r.X+r.W && other.Y+other.H <= r.Y+r.H
}

func (r Rect) Distance(other Rect) float64 {
	if !r.Valid() || !other.Valid() {
		return math.Inf(1)
	}
	dx := max(0, max(other.X-(r.X+r.W), r.X-(other.X+other.W)))
	dy := max(0, max(other.Y-(r.Y+r.H), r.Y-(other.Y+other.H)))
	return (Vector{X: dx, Y: dy}).Length()
}

func (r Rect) ClampedTo(bounds Rect) Rect {
	if !r.Valid() || !bounds.Valid() {
		return r
	}
	if r.W > bounds.W {
		r.X = bounds.X + (bounds.W-r.W)/2
	} else {
		r.X = max(bounds.X, min(bounds.X+bounds.W-r.W, r.X))
	}
	if r.H > bounds.H {
		r.Y = bounds.Y + (bounds.H-r.H)/2
	} else {
		r.Y = max(bounds.Y, min(bounds.Y+bounds.H-r.H, r.Y))
	}
	return r
}

func (r Rect) Overlaps(other Rect) bool {
	return r.Valid() && other.Valid() && r.X < other.X+other.W && r.X+r.W > other.X && r.Y < other.Y+other.H && r.Y+r.H > other.Y
}

func (r Rect) OverlapsOctagon(center Vector, radius float64) bool {
	if !r.Valid() || !center.Valid() || !finite(radius) || radius <= 0 {
		return false
	}
	dx := max(0, math.Abs(r.X+r.W/2-center.X)-r.W/2)
	dy := max(0, math.Abs(r.Y+r.H/2-center.Y)-r.H/2)
	extent := radius * math.Cos(math.Pi/8)
	diagonal := radius * (math.Cos(math.Pi/8) + math.Sin(math.Pi/8))
	return dx < extent && dy < extent && dx+dy < diagonal
}

func finite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}
