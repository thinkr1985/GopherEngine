package nomath

import "math"

type Mat4 [16]float64

func IdentityMatrix() Mat4 {
	return Mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

func isAffine(m Mat4) bool {
	return m[3] == 0 && m[7] == 0 && m[11] == 0 && m[15] == 1
}

func TranslationMatrix(x, y, z float64) Mat4 {
	return Mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		x, y, z, 1,
	}
}

func ScaleMatrix(x, y, z float64) Mat4 {
	return Mat4{
		x, 0, 0, 0,
		0, y, 0, 0,
		0, 0, z, 0,
		0, 0, 0, 1,
	}
}

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

// Optimized inverse assuming the matrix is affine (last row: [0 0 0 1])
func (m Mat4) Inverse() Mat4 {
	// Upper-left 3x3 (rotation/scale)
	r0 := Vec3{m[0], m[1], m[2]}
	r1 := Vec3{m[4], m[5], m[6]}
	r2 := Vec3{m[8], m[9], m[10]}

	// Transpose of rotation matrix (inverse if orthonormal)
	rt0 := Vec3{r0.X, r1.X, r2.X}
	rt1 := Vec3{r0.Y, r1.Y, r2.Y}
	rt2 := Vec3{r0.Z, r1.Z, r2.Z}

	// Translation
	t := Vec3{m[12], m[13], m[14]}

	// Inverted translation
	negT := Vec3{-t.X, -t.Y, -t.Z}

	// New translation = -R^T * T
	newT := Vec3{
		rt0.X*negT.X + rt0.Y*negT.Y + rt0.Z*negT.Z,
		rt1.X*negT.X + rt1.Y*negT.Y + rt1.Z*negT.Z,
		rt2.X*negT.X + rt2.Y*negT.Y + rt2.Z*negT.Z,
	}

	return Mat4{
		rt0.X, rt0.Y, rt0.Z, 0,
		rt1.X, rt1.Y, rt1.Z, 0,
		rt2.X, rt2.Y, rt2.Z, 0,
		newT.X, newT.Y, newT.Z, 1,
	}
}

// Transpose returns the transpose of the matrix
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
func (a Mat4) Multiply(b Mat4) Mat4 {
	if isAffine(a) && isAffine(b) {
		// Optimized affine multiplication
		return Mat4{
			// First column
			a[0]*b[0] + a[4]*b[1] + a[8]*b[2],
			a[1]*b[0] + a[5]*b[1] + a[9]*b[2],
			a[2]*b[0] + a[6]*b[1] + a[10]*b[2],
			0,

			// Second column
			a[0]*b[4] + a[4]*b[5] + a[8]*b[6],
			a[1]*b[4] + a[5]*b[5] + a[9]*b[6],
			a[2]*b[4] + a[6]*b[5] + a[10]*b[6],
			0,

			// Third column
			a[0]*b[8] + a[4]*b[9] + a[8]*b[10],
			a[1]*b[8] + a[5]*b[9] + a[9]*b[10],
			a[2]*b[8] + a[6]*b[9] + a[10]*b[10],
			0,

			// Fourth column (translation)
			a[0]*b[12] + a[4]*b[13] + a[8]*b[14] + a[12],
			a[1]*b[12] + a[5]*b[13] + a[9]*b[14] + a[13],
			a[2]*b[12] + a[6]*b[13] + a[10]*b[14] + a[14],
			1,
		}
	}

	// General full matrix multiplication
	return Mat4{
		a[0]*b[0] + a[4]*b[1] + a[8]*b[2] + a[12]*b[3],
		a[1]*b[0] + a[5]*b[1] + a[9]*b[2] + a[13]*b[3],
		a[2]*b[0] + a[6]*b[1] + a[10]*b[2] + a[14]*b[3],
		a[3]*b[0] + a[7]*b[1] + a[11]*b[2] + a[15]*b[3],

		a[0]*b[4] + a[4]*b[5] + a[8]*b[6] + a[12]*b[7],
		a[1]*b[4] + a[5]*b[5] + a[9]*b[6] + a[13]*b[7],
		a[2]*b[4] + a[6]*b[5] + a[10]*b[6] + a[14]*b[7],
		a[3]*b[4] + a[7]*b[5] + a[11]*b[6] + a[15]*b[7],

		a[0]*b[8] + a[4]*b[9] + a[8]*b[10] + a[12]*b[11],
		a[1]*b[8] + a[5]*b[9] + a[9]*b[10] + a[13]*b[11],
		a[2]*b[8] + a[6]*b[9] + a[10]*b[10] + a[14]*b[11],
		a[3]*b[8] + a[7]*b[9] + a[11]*b[10] + a[15]*b[11],

		a[0]*b[12] + a[4]*b[13] + a[8]*b[14] + a[12]*b[15],
		a[1]*b[12] + a[5]*b[13] + a[9]*b[14] + a[13]*b[15],
		a[2]*b[12] + a[6]*b[13] + a[10]*b[14] + a[14]*b[15],
		a[3]*b[12] + a[7]*b[13] + a[11]*b[14] + a[15]*b[15],
	}
}

func (m Mat4) MultiplyVec4(v Vec4) Vec4 {
	return Vec4{
		X: v.X*m[0] + v.Y*m[4] + v.Z*m[8] + v.W*m[12],
		Y: v.X*m[1] + v.Y*m[5] + v.Z*m[9] + v.W*m[13],
		Z: v.X*m[2] + v.Y*m[6] + v.Z*m[10] + v.W*m[14],
		W: v.X*m[3] + v.Y*m[7] + v.Z*m[11] + v.W*m[15],
	}
}

func (m Mat4) TransformVec3(v Vec3) Vec3 {
	return Vec3{
		X: m[0]*v.X + m[4]*v.Y + m[8]*v.Z + m[12],
		Y: m[1]*v.X + m[5]*v.Y + m[9]*v.Z + m[13],
		Z: m[2]*v.X + m[6]*v.Y + m[10]*v.Z + m[14],
	}
}

func (m Mat4) TransformDirection(v Vec3) Vec3 {
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

func Ortho(left, right, bottom, top, near, far float64) Mat4 {
	return Mat4{
		2 / (right - left), 0, 0, 0,
		0, 2 / (top - bottom), 0, 0,
		0, 0, -2 / (far - near), 0,
		-(right + left) / (right - left),
		-(top + bottom) / (top - bottom),
		-(far + near) / (far - near),
		1,
	}
}

// MultiplyVec3 multiplies a 4x4 matrix with a 3D vector (treats it as vec4 with w=0)
func (m *Mat4) MultiplyVec3(v Vec3) Vec3 {
	return Vec3{
		X: m[0]*v.X + m[4]*v.Y + m[8]*v.Z,
		Y: m[1]*v.X + m[5]*v.Y + m[9]*v.Z,
		Z: m[2]*v.X + m[6]*v.Y + m[10]*v.Z,
	}
}

// MultiplyVec3WithW multiplies a 4x4 matrix with a 3D vector (with specified w component)
func (m *Mat4) MultiplyVec3WithW(v Vec3, w float64) Vec3 {
	return Vec3{
		X: m[0]*v.X + m[4]*v.Y + m[8]*v.Z + m[12]*w,
		Y: m[1]*v.X + m[5]*v.Y + m[9]*v.Z + m[13]*w,
		Z: m[2]*v.X + m[6]*v.Y + m[10]*v.Z + m[14]*w,
	}
}
