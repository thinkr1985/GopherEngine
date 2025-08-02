package nomath

type BoundingBox struct {
	Min Vec3
	Max Vec3
}

func NewBoundingBox() *BoundingBox {
	return &BoundingBox{}
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
