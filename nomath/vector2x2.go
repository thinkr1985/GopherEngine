package nomath

import "math"

// Vec2 represents a 2D vector (U, V) for texture coordinates
type Vec2 struct {
	U, V float64
}

func (v Vec2) Add(other Vec2) Vec2 {
	return Vec2{U: v.U + other.U, V: v.V + other.V}
}

func (v Vec2) Subtract(other Vec2) Vec2 {
	return Vec2{U: v.U - other.U, V: v.V - other.V}
}

func (v Vec2) Multiply(scalar float64) Vec2 {
	return Vec2{U: v.U * scalar, V: v.V * scalar}
}

func (v Vec2) Length() float64 {
	return math.Sqrt(v.U*v.U + v.V*v.V)
}

func (v Vec2) Normalize() Vec2 {
	length := v.Length()
	if length > 0 {
		return Vec2{U: v.U / length, V: v.V / length}
	}
	return v
}
