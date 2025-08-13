package nomath

import (
	"math"
	"sync"
)

// Coordinate system: Y+ up, Z+ forward, X+ left (right-handed)
type Transform struct {
	Position       Vec3 // Position in world space
	Rotation       Vec3 // Euler angles in radians (order: YXZ - yaw, pitch, roll)
	Scale          Vec3 // Scale factors
	ModelMatrix    Mat4
	RotationMatrix Mat4
	Forward        Vec3
	Right          Vec3
	Up             Vec3
	Dirty          bool // track whether transform changed
	Mutex          sync.RWMutex
}

// NewTransform creates a new Transform with default values
func NewTransform() *Transform {
	return &Transform{
		Position:       Vec3{X: 0, Y: 0, Z: 0},
		Rotation:       Vec3{X: 0, Y: 0, Z: 0},
		Scale:          Vec3{X: 1, Y: 1, Z: 1},
		ModelMatrix:    IdentityMatrix(),
		RotationMatrix: IdentityMatrix(),
		Right:          NewVec3(0, 0, 0),
		Forward:        NewVec3(0, 0, 0),
		Up:             NewVec3(0, 0, 0),
	}
}

func (t *Transform) GetMatrix() Mat4 {
	t.UpdateModelMatrix()
	return t.ModelMatrix
}

func (t *Transform) GetModelMatrix() Mat4 {
	t.Mutex.RLock()
	defer t.Mutex.RUnlock()
	return t.GetMatrix()
}

func (t *Transform) GetRotationMatrix() Mat4 {
	t.UpdateModelMatrix()
	return t.RotationMatrix
}

// SetPosition sets the position
func (t *Transform) SetPosition(pos Vec3) {
	if !t.Position.Equals(pos) {
		t.Position = pos
		t.Dirty = true
	}
}

// SetRotation sets the rotation (Euler angles in radians)
// Order: Y (yaw), X (pitch), Z (roll)
func (t *Transform) SetRotation(rot Vec3) {
	rot.X = wrapAngle(rot.X) // Pitch
	rot.Y = wrapAngle(rot.Y) // Yaw
	rot.Z = wrapAngle(rot.Z) // Roll

	if !t.Rotation.Equals(rot) {
		t.Rotation = rot
		t.Dirty = true
	}
}

// SetScale sets the scale with validation
func (t *Transform) SetScale(scale Vec3) {
	// Prevent zero or negative scale
	scale.X = math.Max(scale.X, 0.0001)
	scale.Y = math.Max(scale.Y, 0.0001)
	scale.Z = math.Max(scale.Z, 0.0001)

	if !t.Scale.Equals(scale) {
		t.Scale = scale
		t.Dirty = true
	}
}

// Translate moves the transform by the specified offset
func (t *Transform) Translate(offset Vec3) {
	t.Position = t.Position.Add(offset)
	t.Dirty = true
}

// Rotate adds rotation to the current Euler angles
func (t *Transform) Rotate(rotation Vec3) {
	t.Rotation = t.Rotation.Add(rotation)
	t.Rotation.X = wrapAngle(t.Rotation.X)
	t.Rotation.Y = wrapAngle(t.Rotation.Y)
	t.Rotation.Z = wrapAngle(t.Rotation.Z)
	t.Dirty = true
}
func (t *Transform) GetForward() Vec3 {
	t.UpdateModelMatrix()
	return t.Forward
}

func (t *Transform) GetRight() Vec3 {
	t.UpdateModelMatrix()
	return t.Right
}

func (t *Transform) GetUp() Vec3 {
	t.UpdateModelMatrix()
	return t.Up
}

func (t *Transform) getDirectionFromRotation(x, y, z float64) Vec3 {
	// direction := Vec4{X: x, Y: y, Z: z, W: 0}
	// transformed := t.RotationMatrix.MultiplyVec4(direction)
	// return transformed.ToVec3().Normalize()
	return t.RotationMatrix.TransformDirection(Vec3{x, y, z}).Normalize()
}

// GetWorldPosition returns the transformed position
func (t *Transform) GetWorldPosition() Vec3 {
	return t.Position
}

// GetWorldRotation returns the transformed rotation
func (t *Transform) GetWorldRotation() Vec3 {
	return t.Rotation
}

// GetWorldScale returns the transformed scale
func (t *Transform) GetWorldScale() Vec3 {
	return t.Scale
}

// LookAtMatrix creates a view matrix looking at target
func LookAtMatrix(eye, target, up Vec3) Mat4 {
	forward := target.Subtract(eye).Normalize()
	right := forward.Cross(up).Normalize()
	up = right.Cross(forward).Normalize()

	return Mat4{
		right.X, up.X, -forward.X, 0,
		right.Y, up.Y, -forward.Y, 0,
		right.Z, up.Z, -forward.Z, 0,
		-right.Dot(eye), -up.Dot(eye), forward.Dot(eye), 1,
	}
}

// LookAt makes the transform point toward a target position
func (t *Transform) LookAt(target Vec3, worldUp Vec3) {
	forward := target.Subtract(t.Position).Normalize()
	right := worldUp.Cross(forward).Normalize()
	up := forward.Cross(right).Normalize()

	// Adjust for X+ being left
	right = right.Multiply(-1)

	// Create rotation matrix from basis vectors
	rotMat := Mat4{
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
	if !t.Dirty {
		return
	}
	// Create individual transformation matrices
	translation := TranslationMatrix(
		t.Position.X,
		t.Position.Y,
		t.Position.Z,
	)

	// Fixed rotation order: Yaw (Y) -> Pitch (X) -> Roll (Z)
	rotation := IdentityMatrix().
		Multiply(RotationYMatrix(t.Rotation.Y)).
		Multiply(RotationXMatrix(t.Rotation.X)).
		Multiply(RotationZMatrix(t.Rotation.Z))

	t.RotationMatrix = rotation

	t.Forward = t.getDirectionFromRotation(0, 0, -1)
	t.Right = t.getDirectionFromRotation(1, 0, 0)
	t.Up = t.getDirectionFromRotation(0, 1, 0)

	scale := ScaleMatrix(
		t.Scale.X,
		t.Scale.Y,
		t.Scale.Z,
	)

	// Correct multiplication order: Scale -> Rotation -> Translation
	t.ModelMatrix = IdentityMatrix().
		Multiply(scale).
		Multiply(rotation).
		Multiply(translation)

	t.Dirty = false
}

type SerializableTransform struct {
	Position Vec3 `json:"position"`
	Rotation Vec3 `json:"rotation"`
	Scale    Vec3 `json:"scale"`
}

func (t *Transform) ToSerializable() SerializableTransform {
	return SerializableTransform{
		Position: t.Position,
		Rotation: t.Rotation,
		Scale:    t.Scale,
	}
}
