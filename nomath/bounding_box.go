package nomath

import "math"

type BoundingBox struct {
	Min Vec3
	Max Vec3
}

func NewBoundingBox() *BoundingBox {
	return &BoundingBox{
		Min: Vec3{X: math.MaxFloat64, Y: math.MaxFloat64, Z: math.MaxFloat64},
		Max: Vec3{X: -math.MaxFloat64, Y: -math.MaxFloat64, Z: -math.MaxFloat64},
	}
}

// Transform applies a matrix transform to the bounding box
func (b *BoundingBox) Transform(matrix Mat4) *BoundingBox {
	if b == nil {
		return NewBoundingBox()
	}

	// Transform all 8 corners of the box
	corners := [8]Vec3{
		{X: b.Min.X, Y: b.Min.Y, Z: b.Min.Z},
		{X: b.Min.X, Y: b.Min.Y, Z: b.Max.Z},
		{X: b.Min.X, Y: b.Max.Y, Z: b.Min.Z},
		{X: b.Min.X, Y: b.Max.Y, Z: b.Max.Z},
		{X: b.Max.X, Y: b.Min.Y, Z: b.Min.Z},
		{X: b.Max.X, Y: b.Min.Y, Z: b.Max.Z},
		{X: b.Max.X, Y: b.Max.Y, Z: b.Min.Z},
		{X: b.Max.X, Y: b.Max.Y, Z: b.Max.Z},
	}

	newBox := NewBoundingBox()
	for _, corner := range corners {
		transformed := matrix.MultiplyVec3WithW(corner, 1)
		newBox.Expand(transformed)
	}
	return newBox
}

// Expand expands the bounding box to include the given point
func (b *BoundingBox) Expand(point Vec3) {
	b.Min = Min(b.Min, point)
	b.Max = Max(b.Max, point)
}

// Merge combines two bounding boxes
func (b *BoundingBox) Merge(other *BoundingBox) *BoundingBox {
	if other == nil {
		return b
	}
	return &BoundingBox{
		Min: Min(b.Min, other.Min),
		Max: Max(b.Max, other.Max),
	}
}

func (b *BoundingBox) Center() Vec3 {
	return b.Min.Add(b.Max).Multiply(0.5)
}

func (b *BoundingBox) Size() Vec3 {
	return b.Max.Subtract(b.Min)
}

func (b *BoundingBox) Contains(point Vec3) bool {
	return point.X >= b.Min.X && point.X <= b.Max.X &&
		point.Y >= b.Min.Y && point.Y <= b.Max.Y &&
		point.Z >= b.Min.Z && point.Z <= b.Max.Z
}
