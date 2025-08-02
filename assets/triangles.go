package assets

import (
	"GopherEngine/lookdev"
	"GopherEngine/nomath"
)

type Triangle struct {
	Parent   *Geometry // Reference to parent geometry
	Material *lookdev.Material

	V0  *nomath.Vec3 // Vertex positions
	V1  *nomath.Vec3 // Vertex positions
	V2  *nomath.Vec3 // Vertex positions
	N0  *nomath.Vec3 // Vertex normals
	N1  *nomath.Vec3 // Vertex normals
	N2  *nomath.Vec3 // Vertex normals
	UV0 *nomath.Vec2 // Texture coordinates
	UV1 *nomath.Vec2 // Texture coordinates
	UV2 *nomath.Vec2 // Texture coordinates

}

func NewTriangle(
	geometry *Geometry, material *lookdev.Material,
	v0, v1, v2, n0, n1, n2 *nomath.Vec3,
	uv0, uv1, uv2 *nomath.Vec2) *Triangle {
	return &Triangle{
		Parent:   geometry,
		Material: material,
		V0:       v0,
		V1:       v1,
		V2:       v2,
		N0:       n0,
		N1:       n1,
		N2:       n2,
		UV0:      uv0,
		UV1:      uv1,
		UV2:      uv2,
	}
}

func (t *Triangle) Centroid() nomath.Vec3 {
	return (*t.V0).Add(*t.V1).Add(*t.V2).Multiply(1.0 / 3.0)
}

func (t *Triangle) Area() float64 {
	edge1 := t.V1.Subtract(*t.V0)
	edge2 := t.V2.Subtract(*t.V0)
	return edge1.Cross(edge2).Length() * 0.5
}

func (t *Triangle) Normal() nomath.Vec3 {
	edge1 := t.V1.Subtract(*t.V0)
	edge2 := t.V2.Subtract(*t.V0)
	return edge1.Cross(edge2).Normalize()
}

func (t *Triangle) InterpolatedNormal(u, v, w float64) nomath.Vec3 {
	n := t.N0.Multiply(u).Add(t.N1.Multiply(v)).Add(t.N2.Multiply(w))
	return n.Normalize()
}

func (t *Triangle) InterpolatedUV(u, v, w float64) nomath.Vec2 {
	return nomath.Vec2{
		U: t.UV0.U*u + t.UV1.U*v + t.UV2.U*w,
		V: t.UV0.V*u + t.UV1.V*v + t.UV2.V*w,
	}
}

func (t *Triangle) Barycentric(p nomath.Vec2) (u, v, w float64) {
	// Convert Vec3 to 2D for screen-space interpolation
	a := nomath.Vec2{U: t.V0.X, V: t.V0.Y}
	b := nomath.Vec2{U: t.V1.X, V: t.V1.Y}
	c := nomath.Vec2{U: t.V2.X, V: t.V2.Y}

	denom := (b.V-c.V)*(a.U-c.U) + (c.U-b.U)*(a.V-c.V)
	u = ((b.V-c.V)*(p.U-c.U) + (c.U-b.U)*(p.V-c.V)) / denom
	v = ((c.V-a.V)*(p.U-c.U) + (a.U-c.U)*(p.V-c.V)) / denom
	w = 1 - u - v
	return
}

func (t *Triangle) ContainsPoint2D(p nomath.Vec2) bool {
	u, v, w := t.Barycentric(p)
	return u >= 0 && v >= 0 && w >= 0 && u <= 1 && v <= 1 && w <= 1
}
