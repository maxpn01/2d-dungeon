package physics

import "math"

type Vector struct {
	X, Y float64
}

func (v Vector) Valid() bool {
	return finite(v.X) && finite(v.Y)
}

func (v Vector) LengthSquared() float64 {
	return v.X*v.X + v.Y*v.Y
}

func (v Vector) Length() float64 {
	return math.Hypot(v.X, v.Y)
}

func (v Vector) DistanceSquared(other Vector) float64 {
	return (Vector{X: v.X - other.X, Y: v.Y - other.Y}).LengthSquared()
}

func (v Vector) Distance(other Vector) float64 {
	return (Vector{X: v.X - other.X, Y: v.Y - other.Y}).Length()
}

func (v Vector) Add(other Vector) Vector {
	return Vector{X: v.X + other.X, Y: v.Y + other.Y}
}

func (v Vector) Sub(other Vector) Vector {
	return Vector{X: v.X - other.X, Y: v.Y - other.Y}
}

func (v Vector) Scale(factor float64) Vector {
	return Vector{X: v.X * factor, Y: v.Y * factor}
}

func (v Vector) Normalized() Vector {
	length := v.Length()
	if !v.Valid() || !finite(length) || length == 0 {
		return Vector{}
	}
	return Vector{X: v.X / length, Y: v.Y / length}
}
