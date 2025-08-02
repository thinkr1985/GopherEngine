package nomath

import "math"

// Vec3 represents a 3D vector in our coordinate system (X, Y, Z)------------------------------------------
type Vec3 struct {
	X, Y, Z float64
}

// Helper methods for Vec3
func (v Vec3) Add(other Vec3) Vec3 {
	return Vec3{X: v.X + other.X, Y: v.Y + other.Y, Z: v.Z + other.Z}
}

func (v Vec3) Subtract(other Vec3) Vec3 {
	return Vec3{X: v.X - other.X, Y: v.Y - other.Y, Z: v.Z - other.Z}
}

func (v Vec3) Cross(other Vec3) Vec3 {
	return Vec3{
		X: v.Y*other.Z - v.Z*other.Y,
		Y: v.Z*other.X - v.X*other.Z,
		Z: v.X*other.Y - v.Y*other.X,
	}
}

func (v Vec3) Normalize() Vec3 {
	length := math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
	if length > 0 {
		return Vec3{X: v.X / length, Y: v.Y / length, Z: v.Z / length}
	}
	return v
}
func (v Vec3) Multiply(scalar float64) Vec3 {
	return Vec3{X: v.X * scalar, Y: v.Y * scalar, Z: v.Z * scalar}
}

// Length returns the magnitude of the vector
func (v Vec3) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

// Dot returns the dot product of two vectors
func (v Vec3) Dot(other Vec3) float64 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

// Add to Vec3 methods in camera.go
func (v Vec3) LengthSquared() float64 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

// NewVec3 creates a new 3D vector
func NewVec3(x, y, z float64) Vec3 {
	return Vec3{X: x, Y: y, Z: z}
}

// In your vector.go file (or wherever your Vec3 is defined)
func (v Vec3) Equals(other Vec3) bool {
	const epsilon = 1e-6 // Small value to account for floating point precision
	return math.Abs(v.X-other.X) < epsilon &&
		math.Abs(v.Y-other.Y) < epsilon &&
		math.Abs(v.Z-other.Z) < epsilon
}

// Negate returns the negated vector
func (v Vec3) Negate() Vec3 {
	return Vec3{X: -v.X, Y: -v.Y, Z: -v.Z}
}

// Reflect calculates the reflection vector
func (v Vec3) Reflect(normal Vec3) Vec3 {
	return v.Subtract(normal.Multiply(2 * v.Dot(normal)))
}

// EqualsEpsilon compares two vectors with a specified tolerance
func (v Vec3) EqualsEpsilon(other Vec3, epsilon float64) bool {
	return math.Abs(v.X-other.X) < epsilon &&
		math.Abs(v.Y-other.Y) < epsilon &&
		math.Abs(v.Z-other.Z) < epsilon
}
