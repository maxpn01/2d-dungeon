package physics

import "math"

const (
	defaultMaxStep    = 8.0
	maxSubsteps       = 65536
	contactIterations = 24
)

type World struct {
	CanOccupy func(Rect) bool
	MaxStep   float64
}

type Motion struct {
	Bounds             Rect
	BlockedX, BlockedY bool
	Rejected           bool
}

func (m Motion) Blocked() bool {
	return m.BlockedX || m.BlockedY
}

func (w World) MoveAndSlide(body Rect, delta Vector) Motion {
	result := Motion{Bounds: body}
	steps, valid := w.substeps(body, delta)
	if !valid {
		result.Rejected = true
		return result
	}
	step := Vector{delta.X / float64(steps), delta.Y / float64(steps)}
	for range steps {
		var blocked bool
		result.Bounds, blocked = w.advance(result.Bounds, Vector{X: step.X})
		result.BlockedX = result.BlockedX || blocked
		result.Bounds, blocked = w.advance(result.Bounds, Vector{Y: step.Y})
		result.BlockedY = result.BlockedY || blocked
	}
	return result
}

func (w World) MovementLimit(body Rect, delta Vector) (fraction float64, blocked bool) {
	steps, valid := w.substeps(body, delta)
	if !valid || !w.CanOccupy(body) {
		return 0, true
	}
	for i := 1; i <= steps; i++ {
		fraction := float64(i) / float64(steps)
		if w.CanOccupy(body.Translated(Vector{delta.X * fraction, delta.Y * fraction})) {
			continue
		}
		lo, hi := float64(i-1)/float64(steps), fraction
		for range contactIterations {
			mid := (lo + hi) / 2
			if w.CanOccupy(body.Translated(Vector{delta.X * mid, delta.Y * mid})) {
				lo = mid
			} else {
				hi = mid
			}
		}
		return lo, true
	}
	return 1, false
}

func (w World) advance(body Rect, delta Vector) (Rect, bool) {
	if delta.X == 0 && delta.Y == 0 {
		return body, false
	}
	next := body.Translated(delta)
	if w.CanOccupy(next) {
		return next, false
	}
	if !w.CanOccupy(body) {
		return body, true
	}
	lo, hi := 0.0, 1.0
	for range contactIterations {
		mid := (lo + hi) / 2
		if w.CanOccupy(body.Translated(Vector{delta.X * mid, delta.Y * mid})) {
			lo = mid
		} else {
			hi = mid
		}
	}
	return body.Translated(Vector{delta.X * lo, delta.Y * lo}), true
}

func (w World) substeps(body Rect, delta Vector) (int, bool) {
	if w.CanOccupy == nil || !body.Valid() || !delta.Valid() || !body.Translated(delta).Valid() {
		return 0, false
	}
	step := w.MaxStep
	if !finite(step) || step <= 0 {
		step = defaultMaxStep
	}
	step = min(step, min(body.W, body.H)/2)
	count := max(1, math.Ceil(max(math.Abs(delta.X), math.Abs(delta.Y))/step))
	if !finite(count) || count > maxSubsteps {
		return 0, false
	}
	return int(count), true
}
