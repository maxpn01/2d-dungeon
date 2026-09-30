package physics

func (body Rect) CollisionFraction(delta Vector, target Rect, targetDelta Vector) (fraction float64, hit bool) {
	if !body.Valid() || !target.Valid() || !delta.Valid() || !targetDelta.Valid() || !body.Translated(delta).Valid() || !target.Translated(targetDelta).Valid() {
		return 0, false
	}
	relative := Vector{delta.X - targetDelta.X, delta.Y - targetDelta.Y}
	if !relative.Valid() {
		return 0, false
	}
	enter, exit := 0.0, 1.0
	for _, axis := range [2][4]float64{
		{body.X, relative.X, target.X - body.W, target.X + target.W},
		{body.Y, relative.Y, target.Y - body.H, target.Y + target.H},
	} {
		start, movement, lo, hi := axis[0], axis[1], axis[2], axis[3]
		if movement == 0 {
			if start <= lo || start >= hi {
				return 0, false
			}
			continue
		}
		a, b := (lo-start)/movement, (hi-start)/movement
		if a > b {
			a, b = b, a
		}
		enter, exit = max(enter, a), min(exit, b)
		if enter >= exit {
			return 0, false
		}
	}
	return enter, enter < 1 && exit > 0
}
