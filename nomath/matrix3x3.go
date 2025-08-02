package nomath

// Mat3 represents a 3x3 matrix
type Mat3 [9]float64

func (m Mat3) MultiplyVec3(v Vec3) Vec3 {
	return Vec3{
		X: m[0]*v.X + m[1]*v.Y + m[2]*v.Z,
		Y: m[3]*v.X + m[4]*v.Y + m[5]*v.Z,
		Z: m[6]*v.X + m[7]*v.Y + m[8]*v.Z,
	}
}
