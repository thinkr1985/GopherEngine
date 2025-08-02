package core

import (
	"GopherEngine/nomath"
	"math"
)

// Transform represents position, rotation (Euler angles), and scale in 3D space
// Coordinate system: Y+ up, Z+ forward, X+ left (right-handed)
type Transform struct {
	Position    nomath.Vec3 // Position in world space
	Rotation    nomath.Vec3 // Euler angles in radians (order: YXZ - yaw, pitch, roll)
	Scale       nomath.Vec3 // Scale factors
	ModelMatrix nomath.Mat4
}

// NewTransform creates a new Transform with default values
func NewTransform() *Transform {
	return &Transform{
		Position:    nomath.Vec3{X: 0, Y: 0, Z: 0},
		Rotation:    nomath.Vec3{X: 0, Y: 0, Z: 0},
		Scale:       nomath.Vec3{X: 1, Y: 1, Z: 1},
		ModelMatrix: nomath.IdentityMatrix(),
	}
}

// SetPosition sets the position
func (t *Transform) SetPosition(pos nomath.Vec3) {
	if !t.Position.Equals(pos) {
		t.Position = pos
	}
}

// SetRotation sets the rotation (Euler angles in radians)
// Order: Y (yaw), X (pitch), Z (roll)
func (t *Transform) SetRotation(rot nomath.Vec3) {
	rot.X = wrapAngle(rot.X) // Pitch
	rot.Y = wrapAngle(rot.Y) // Yaw
	rot.Z = wrapAngle(rot.Z) // Roll

	if !t.Rotation.Equals(rot) {
		t.Rotation = rot
	}
}

// SetScale sets the scale with validation
func (t *Transform) SetScale(scale nomath.Vec3) {
	// Prevent zero or negative scale
	scale.X = math.Max(scale.X, 0.0001)
	scale.Y = math.Max(scale.Y, 0.0001)
	scale.Z = math.Max(scale.Z, 0.0001)

	if !t.Scale.Equals(scale) {
		t.Scale = scale
	}
}

// Translate moves the transform by the specified offset
func (t *Transform) Translate(offset nomath.Vec3) {
	t.Position = t.Position.Add(offset)
}

// Rotate adds rotation to the current Euler angles
func (t *Transform) Rotate(rotation nomath.Vec3) {
	t.Rotation = t.Rotation.Add(rotation)
	t.Rotation.X = wrapAngle(t.Rotation.X)
	t.Rotation.Y = wrapAngle(t.Rotation.Y)
	t.Rotation.Z = wrapAngle(t.Rotation.Z)
}

// GetForward returns the forward vector (Z+)
func (t *Transform) GetForward() nomath.Vec3 {
	return t.getDirectionFromRotation(0, 0, 1)
}

// GetUp returns the up vector (Y+)
func (t *Transform) GetUp() nomath.Vec3 {
	return t.getDirectionFromRotation(0, 1, 0)
}

// GetRight returns the right vector (X-)
func (t *Transform) GetRight() nomath.Vec3 {
	return t.getDirectionFromRotation(-1, 0, 0)
}

// getDirectionFromRotation calculates a direction vector from rotation
func (t *Transform) getDirectionFromRotation(x, y, z float64) nomath.Vec3 {
	// Rotation order: Yaw (Y), Pitch (X), Roll (Z)
	rotation := nomath.IdentityMatrix().
		Multiply(nomath.RotationYMatrix(t.Rotation.Y)). // Yaw
		Multiply(nomath.RotationXMatrix(t.Rotation.X)). // Pitch
		Multiply(nomath.RotationZMatrix(t.Rotation.Z))  // Roll

	direction := nomath.Vec4{X: x, Y: y, Z: z, W: 0}
	transformed := rotation.MultiplyVec4(direction) // Fixed typo here (was MultiplyVec4)
	return transformed.ToVec3().Normalize()
}

// GetWorldPosition returns the transformed position
func (t *Transform) GetWorldPosition() nomath.Vec3 {
	return t.Position
}

// GetWorldRotation returns the transformed rotation
func (t *Transform) GetWorldRotation() nomath.Vec3 {
	return t.Rotation
}

// GetWorldScale returns the transformed scale
func (t *Transform) GetWorldScale() nomath.Vec3 {
	return t.Scale
}

// LookAtMatrix creates a view matrix looking at target
func LookAtMatrix(eye, target, up nomath.Vec3) nomath.Mat4 {
	f := target.Subtract(eye).Normalize()
	s := f.Cross(up).Normalize()
	u := s.Cross(f)

	return nomath.Mat4{
		s.X, u.X, -f.X, 0,
		s.Y, u.Y, -f.Y, 0,
		s.Z, u.Z, -f.Z, 0,
		-s.Dot(eye), -u.Dot(eye), f.Dot(eye), 1,
	}
}

// LookAt makes the transform point toward a target position
func (t *Transform) LookAt(target nomath.Vec3, worldUp nomath.Vec3) {
	forward := target.Subtract(t.Position).Normalize()
	right := worldUp.Cross(forward).Normalize()
	up := forward.Cross(right).Normalize()

	// Adjust for X+ being left
	right = right.Multiply(-1)

	// Create rotation matrix from basis vectors
	rotMat := nomath.Mat4{
		right.X, right.Y, right.Z, 0,
		up.X, up.Y, up.Z, 0,
		forward.X, forward.Y, forward.Z, 0,
		0, 0, 0, 1,
	}

	// Convert to Euler angles (simplified - implement proper conversion)
	t.Rotation = rotMat.ToEulerAnglesYXZ()
}

// Equals checks if two transforms are approximately equal
func (t *Transform) Equals(other *Transform) bool {
	const epsilon = 0.0001
	return t.Position.EqualsEpsilon(other.Position, epsilon) &&
		t.Rotation.EqualsEpsilon(other.Rotation, epsilon) &&
		t.Scale.EqualsEpsilon(other.Scale, epsilon)
}

// wrapAngle keeps angles in the range [-π, π]
func wrapAngle(angle float64) float64 {
	angle = math.Mod(angle, 2*math.Pi)
	if angle > math.Pi {
		angle -= 2 * math.Pi
	} else if angle <= -math.Pi {
		angle += 2 * math.Pi
	}
	return angle
}

// UpdateModelMatrix updates the model matrix and marks geometry as needing update
func (t *Transform) UpdateModelMatrix() {
	// Create individual transformation matrices
	translation := nomath.TranslationMatrix(
		t.Position.X,
		t.Position.Y,
		t.Position.Z,
	)

	// Fixed rotation order: Yaw (Y) -> Pitch (X) -> Roll (Z)
	rotation := nomath.IdentityMatrix().
		Multiply(nomath.RotationYMatrix(t.Rotation.Y)). // Yaw
		Multiply(nomath.RotationXMatrix(t.Rotation.X)). // Pitch
		Multiply(nomath.RotationZMatrix(t.Rotation.Z))  // Roll

	scale := nomath.ScaleMatrix(
		t.Scale.X,
		t.Scale.Y,
		t.Scale.Z,
	)

	// Correct multiplication order: Scale -> Rotation -> Translation
	t.ModelMatrix = nomath.IdentityMatrix().
		Multiply(scale).
		Multiply(rotation).
		Multiply(translation)

}
