package nomath

import "math"

// Mat4 represents a 4x4 matrix in column-major order-----------------------------------------------
type Mat4 [16]float64

// IdentityMatrix returns a 4x4 identity matrix
func IdentityMatrix() Mat4 {
	return Mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

// TranslationMatrix creates a translation matrix
func TranslationMatrix(x, y, z float64) Mat4 {
	return Mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		x, y, z, 1,
	}
}

// ScaleMatrix creates a scaling matrix
func ScaleMatrix(x, y, z float64) Mat4 {
	return Mat4{
		x, 0, 0, 0,
		0, y, 0, 0,
		0, 0, z, 0,
		0, 0, 0, 1,
	}
}

// RotationXMatrix creates a rotation matrix around X axis (pitch)
func RotationXMatrix(angle float64) Mat4 {
	c := math.Cos(angle)
	s := math.Sin(angle)
	return Mat4{
		1, 0, 0, 0,
		0, c, -s, 0,
		0, s, c, 0,
		0, 0, 0, 1,
	}
}

// RotationYMatrix creates a rotation matrix around Y axis (yaw)
func RotationYMatrix(angle float64) Mat4 {
	c := math.Cos(angle)
	s := math.Sin(angle)
	return Mat4{
		c, 0, s, 0,
		0, 1, 0, 0,
		-s, 0, c, 0,
		0, 0, 0, 1,
	}
}

// RotationZMatrix creates a rotation matrix around Z axis (roll)
func RotationZMatrix(angle float64) Mat4 {
	c := math.Cos(angle)
	s := math.Sin(angle)
	return Mat4{
		c, -s, 0, 0,
		s, c, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

// Inverse returns the inverse of the matrix.
// If the matrix is non-invertible, it returns the identity matrix.
func (m Mat4) Inverse() Mat4 {
	inv := Mat4{}

	inv[0] = m[5]*m[10]*m[15] - m[5]*m[11]*m[14] - m[9]*m[6]*m[15] + m[9]*m[7]*m[14] + m[13]*m[6]*m[11] - m[13]*m[7]*m[10]
	inv[4] = -m[4]*m[10]*m[15] + m[4]*m[11]*m[14] + m[8]*m[6]*m[15] - m[8]*m[7]*m[14] - m[12]*m[6]*m[11] + m[12]*m[7]*m[10]
	inv[8] = m[4]*m[9]*m[15] - m[4]*m[11]*m[13] - m[8]*m[5]*m[15] + m[8]*m[7]*m[13] + m[12]*m[5]*m[11] - m[12]*m[7]*m[9]
	inv[12] = -m[4]*m[9]*m[14] + m[4]*m[10]*m[13] + m[8]*m[5]*m[14] - m[8]*m[6]*m[13] - m[12]*m[5]*m[10] + m[12]*m[6]*m[9]

	inv[1] = -m[1]*m[10]*m[15] + m[1]*m[11]*m[14] + m[9]*m[2]*m[15] - m[9]*m[3]*m[14] - m[13]*m[2]*m[11] + m[13]*m[3]*m[10]
	inv[5] = m[0]*m[10]*m[15] - m[0]*m[11]*m[14] - m[8]*m[2]*m[15] + m[8]*m[3]*m[14] + m[12]*m[2]*m[11] - m[12]*m[3]*m[10]
	inv[9] = -m[0]*m[9]*m[15] + m[0]*m[11]*m[13] + m[8]*m[1]*m[15] - m[8]*m[3]*m[13] - m[12]*m[1]*m[11] + m[12]*m[3]*m[9]
	inv[13] = m[0]*m[9]*m[14] - m[0]*m[10]*m[13] - m[8]*m[1]*m[14] + m[8]*m[2]*m[13] + m[12]*m[1]*m[10] - m[12]*m[2]*m[9]

	inv[2] = m[1]*m[6]*m[15] - m[1]*m[7]*m[14] - m[5]*m[2]*m[15] + m[5]*m[3]*m[14] + m[13]*m[2]*m[7] - m[13]*m[3]*m[6]
	inv[6] = -m[0]*m[6]*m[15] + m[0]*m[7]*m[14] + m[4]*m[2]*m[15] - m[4]*m[3]*m[14] - m[12]*m[2]*m[7] + m[12]*m[3]*m[6]
	inv[10] = m[0]*m[5]*m[15] - m[0]*m[7]*m[13] - m[4]*m[1]*m[15] + m[4]*m[3]*m[13] + m[12]*m[1]*m[7] - m[12]*m[3]*m[5]
	inv[14] = -m[0]*m[5]*m[14] + m[0]*m[6]*m[13] + m[4]*m[1]*m[14] - m[4]*m[2]*m[13] - m[12]*m[1]*m[6] + m[12]*m[2]*m[5]

	inv[3] = -m[1]*m[6]*m[11] + m[1]*m[7]*m[10] + m[5]*m[2]*m[11] - m[5]*m[3]*m[10] - m[9]*m[2]*m[7] + m[9]*m[3]*m[6]
	inv[7] = m[0]*m[6]*m[11] - m[0]*m[7]*m[10] - m[4]*m[2]*m[11] + m[4]*m[3]*m[10] + m[8]*m[2]*m[7] - m[8]*m[3]*m[6]
	inv[11] = -m[0]*m[5]*m[11] + m[0]*m[7]*m[9] + m[4]*m[1]*m[11] - m[4]*m[3]*m[9] - m[8]*m[1]*m[7] + m[8]*m[3]*m[5]
	inv[15] = m[0]*m[5]*m[10] - m[0]*m[6]*m[9] - m[4]*m[1]*m[10] + m[4]*m[2]*m[9] + m[8]*m[1]*m[6] - m[8]*m[2]*m[5]

	det := m[0]*inv[0] + m[1]*inv[4] + m[2]*inv[8] + m[3]*inv[12]

	if det == 0 {
		// Non-invertible matrix, return identity as fallback
		return IdentityMatrix()
	}

	invDet := 1.0 / det
	for i := 0; i < 16; i++ {
		inv[i] *= invDet
	}

	return inv
}

func (m Mat4) Transpose() Mat4 {
	return Mat4{
		m[0], m[4], m[8], m[12],
		m[1], m[5], m[9], m[13],
		m[2], m[6], m[10], m[14],
		m[3], m[7], m[11], m[15],
	}
}

func (m Mat4) ToEulerAnglesYXZ() Vec3 {
	// Matrix indices for column-major 16-element array:
	// 0  4  8  12
	// 1  5  9  13
	// 2  6 10  14
	// 3  7 11  15

	var angles Vec3

	// Yaw (Y rotation)
	angles.Y = math.Asin(-m[2]) // m[2] is m[2][0] in row-major

	// Handle gimbal lock
	if math.Cos(angles.Y) > 0.0001 {
		// Pitch (X rotation)
		angles.X = math.Atan2(m[6], m[10]) // m[6] is m[1][2], m[10] is m[2][2]
		// Roll (Z rotation)
		angles.Z = math.Atan2(m[1], m[0]) // m[1] is m[0][1], m[0] is m[0][0]
	} else {
		// Pitch (X rotation)
		angles.X = math.Atan2(-m[9], m[5]) // m[9] is m[2][1], m[5] is m[1][1]
		// Roll (Z rotation)
		angles.Z = 0
	}

	return angles
}

// Multiply optimized version (unrolled loops where possible)
func (m Mat4) Multiply(other Mat4) Mat4 {
	return Mat4{
		m[0]*other[0] + m[4]*other[1] + m[8]*other[2] + m[12]*other[3],
		m[1]*other[0] + m[5]*other[1] + m[9]*other[2] + m[13]*other[3],
		m[2]*other[0] + m[6]*other[1] + m[10]*other[2] + m[14]*other[3],
		m[3]*other[0] + m[7]*other[1] + m[11]*other[2] + m[15]*other[3],

		m[0]*other[4] + m[4]*other[5] + m[8]*other[6] + m[12]*other[7],
		m[1]*other[4] + m[5]*other[5] + m[9]*other[6] + m[13]*other[7],
		m[2]*other[4] + m[6]*other[5] + m[10]*other[6] + m[14]*other[7],
		m[3]*other[4] + m[7]*other[5] + m[11]*other[6] + m[15]*other[7],

		m[0]*other[8] + m[4]*other[9] + m[8]*other[10] + m[12]*other[11],
		m[1]*other[8] + m[5]*other[9] + m[9]*other[10] + m[13]*other[11],
		m[2]*other[8] + m[6]*other[9] + m[10]*other[10] + m[14]*other[11],
		m[3]*other[8] + m[7]*other[9] + m[11]*other[10] + m[15]*other[11],

		m[0]*other[12] + m[4]*other[13] + m[8]*other[14] + m[12]*other[15],
		m[1]*other[12] + m[5]*other[13] + m[9]*other[14] + m[13]*other[15],
		m[2]*other[12] + m[6]*other[13] + m[10]*other[14] + m[14]*other[15],
		m[3]*other[12] + m[7]*other[13] + m[11]*other[14] + m[15]*other[15],
	}
}

// MultiplyVec4 optimized version (unrolled loops)
func (m Mat4) MultiplyVec4(v Vec4) Vec4 {
	return Vec4{
		X: v.X*m[0] + v.Y*m[4] + v.Z*m[8] + v.W*m[12],
		Y: v.X*m[1] + v.Y*m[5] + v.Z*m[9] + v.W*m[13],
		Z: v.X*m[2] + v.Y*m[6] + v.Z*m[10] + v.W*m[14],
		W: v.X*m[3] + v.Y*m[7] + v.Z*m[11] + v.W*m[15],
	}
}

// TransformVec3 transforms a Vec3 using the 3x3 portion of the matrix (ignoring translation).
// func (m Mat4) TransformVec3(v Vec3) Vec3 {
// 	return Vec3{
// 		X: m[0]*v.X + m[4]*v.Y + m[8]*v.Z,
// 		Y: m[1]*v.X + m[5]*v.Y + m[9]*v.Z,
// 		Z: m[2]*v.X + m[6]*v.Y + m[10]*v.Z,
// 	}
// }

func (m Mat4) TransformVec3(v Vec3) Vec3 {
	return Vec3{
		X: m[0]*v.X + m[4]*v.Y + m[8]*v.Z,
		Y: m[1]*v.X + m[5]*v.Y + m[9]*v.Z,
		Z: m[2]*v.X + m[6]*v.Y + m[10]*v.Z,
	}
}

// Optimized MultiplyVec4 for batch transformations
func (m Mat4) MultiplyVec4Batch(vectors []Vec4) []Vec4 {
	results := make([]Vec4, len(vectors))
	for i, v := range vectors {
		results[i] = Vec4{
			X: v.X*m[0] + v.Y*m[4] + v.Z*m[8] + v.W*m[12],
			Y: v.X*m[1] + v.Y*m[5] + v.Z*m[9] + v.W*m[13],
			Z: v.X*m[2] + v.Y*m[6] + v.Z*m[10] + v.W*m[14],
			W: v.X*m[3] + v.Y*m[7] + v.Z*m[11] + v.W*m[15],
		}
	}
	return results
}
