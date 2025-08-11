package nomath

import "math"

// Vec4 represents a 4D vector (X, Y, Z, W) for homogeneous coordinates---------------------------------------
type Vec4 struct {
	X, Y, Z, W float64
}

func (v Vec4) Multiply(scalar float64) Vec4 {
	return Vec4{
		X: v.X * scalar,
		Y: v.Y * scalar,
		Z: v.Z * scalar,
		W: v.W * scalar,
	}
}

// Add vector addition for Vec4
func (v Vec4) Add(other Vec4) Vec4 {
	return Vec4{
		X: v.X + other.X,
		Y: v.Y + other.Y,
		Z: v.Z + other.Z,
		W: v.W + other.W,
	}
}

// NewVec4 creates a new 4D vector
func NewVec4(x, y, z, w float64) Vec4 {
	return Vec4{X: x, Y: y, Z: z, W: w}
}

// ToVec3 converts a Vec4 to Vec3 by perspective division
func (v Vec4) ToVec3() Vec3 {
	if v.W != 0 {
		return Vec3{X: v.X / v.W, Y: v.Y / v.W, Z: v.Z / v.W}
	}
	return Vec3{X: v.X, Y: v.Y, Z: v.Z}
}
func (v Vec4) EqualsEpsilon(other Vec4, epsilon float64) bool {
	return math.Abs(v.X-other.X) < epsilon &&
		math.Abs(v.Y-other.Y) < epsilon &&
		math.Abs(v.Z-other.Z) < epsilon &&
		math.Abs(v.W-other.W) < epsilon
}

func (v Vec3) ToVec4(w float64) Vec4 {
	return Vec4{X: v.X, Y: v.Y, Z: v.Z, W: w}
}

// Add to your vector4x4.go file
func (v Vec4) Normalize() Vec4 {
	// For plane normalization, we only care about XYZ components
	xyzLength := math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
	if xyzLength > 0 {
		invLength := 1.0 / xyzLength
		return Vec4{
			X: v.X * invLength,
			Y: v.Y * invLength,
			Z: v.Z * invLength,
			W: v.W * invLength, // Also normalize W for consistency
		}
	}
	return v
}

// Optimized version specifically for frustum planes
func (v Vec4) NormalizePlane() Vec4 {
	// Special normalization for plane equations (where W is distance)
	xyzLength := math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
	if xyzLength > 0 {
		invLength := 1.0 / xyzLength
		return Vec4{
			X: v.X * invLength,
			Y: v.Y * invLength,
			Z: v.Z * invLength,
			W: v.W * invLength, // Important: normalize distance too!
		}
	}
	return v
}

func (v Vec4) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z + v.W*v.W)
}

// Sub performs vector subtraction: v - other
func (v Vec4) Sub(other Vec4) Vec4 {
	return Vec4{
		X: v.X - other.X,
		Y: v.Y - other.Y,
		Z: v.Z - other.Z,
		W: v.W - other.W,
	}
}

// Divide divides all components of the vector by a scalar
func (v Vec4) Divide(scalar float64) Vec4 {
	if scalar != 0 {
		inv := 1.0 / scalar
		return Vec4{
			X: v.X * inv,
			Y: v.Y * inv,
			Z: v.Z * inv,
			W: v.W * inv,
		}
	}
	return v // Or panic/log if divide-by-zero should be caught
}

func (v Vec4) DivideW() Vec4 {
	if v.W == 0 {
		return v // or panic/log
	}
	return Vec4{
		X: v.X / v.W,
		Y: v.Y / v.W,
		Z: v.Z / v.W,
		W: 1.0, // convention: after perspective divide
	}
}
