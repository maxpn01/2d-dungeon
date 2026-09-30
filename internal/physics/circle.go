package physics

type Circle struct {
	Center Vector
	Radius float64
}

func (c Circle) Valid() bool {
	return c.Center.Valid() && finite(c.Radius) && c.Radius > 0
}

func (c Circle) Contains(other Circle) bool {
	return c.Valid() && other.Valid() && c.Center.Distance(other.Center)+other.Radius <= c.Radius
}
