package assets

import (
	"GopherEngine/lookdev"
	"GopherEngine/nomath"
	"math"
)

type Triangle struct {
	Parent   *Geometry // Reference to parent geometry
	Material *lookdev.Material

	V0              *nomath.Vec3 // Vertex positions
	V1              *nomath.Vec3 // Vertex positions
	V2              *nomath.Vec3 // Vertex positions
	N0              *nomath.Vec3 // Vertex normals
	N1              *nomath.Vec3 // Vertex normals
	N2              *nomath.Vec3 // Vertex normals
	UV0             *nomath.Vec2 // Texture coordinates
	UV1             *nomath.Vec2 // Texture coordinates
	UV2             *nomath.Vec2 // Texture coordinates
	DiffuseBuffer   *lookdev.ColorRGBA
	SpecularBuffer  *lookdev.ColorRGBA
	AlphaBuffer     float64 // Separate alpha buffer for transparency
	BufferCache     bool
	LightDotNormals []float64   // one per light
	WorldNormal     nomath.Vec3 // transformed normal after applying NormalMatrix
	HasTexture      bool        // New field to track texture status
}

func NewTriangle(
	geometry *Geometry, material *lookdev.Material,
	v0, v1, v2, n0, n1, n2 *nomath.Vec3,
	uv0, uv1, uv2 *nomath.Vec2) *Triangle {

	// Ensure we have a material
	if material == nil {
		material = lookdev.NewMaterial("default")
	}

	tri := &Triangle{
		Parent:         geometry,
		Material:       material,
		V0:             v0,
		V1:             v1,
		V2:             v2,
		N0:             n0,
		N1:             n1,
		N2:             n2,
		UV0:            uv0,
		UV1:            uv1,
		UV2:            uv2,
		DiffuseBuffer:  &material.DiffuseColor, // Initialize with material color
		SpecularBuffer: &material.SpecularColor,
		BufferCache:    false,
	}

	return tri
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
	if t.UV0 == nil || t.UV1 == nil || t.UV2 == nil {
		return nomath.Vec2{U: 0, V: 0}
	}

	// Normalize barycentric weights
	sum := u + v + w
	if sum != 0 {
		inv := 1.0 / sum
		u *= inv
		v *= inv
		w *= inv
	}

	return nomath.Vec2{
		U: t.UV0.U*u + t.UV1.U*v + t.UV2.U*w,
		V: t.UV0.V*u + t.UV1.V*v + t.UV2.V*w,
	}
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func Barycentric(p nomath.Vec2, v0Screen, v1Screen, v2Screen nomath.Vec2) (u, v, w float64) {
	denom := (v1Screen.V-v2Screen.V)*(v0Screen.U-v2Screen.U) +
		(v2Screen.U-v1Screen.U)*(v0Screen.V-v2Screen.V)

	if math.Abs(denom) < 1e-6 {
		return -1, -1, -1 // Degenerate triangle
	}

	u = ((v1Screen.V-v2Screen.V)*(p.U-v2Screen.U) +
		(v2Screen.U-v1Screen.U)*(p.V-v2Screen.V)) / denom
	v = ((v2Screen.V-v0Screen.V)*(p.U-v2Screen.U) +
		(v0Screen.U-v2Screen.U)*(p.V-v2Screen.V)) / denom
	w = 1 - u - v
	return
}
func (t *Triangle) PreComputeBuffers() {
	if t.BufferCache || t.Material == nil {
		return
	}

	// Initialize buffers if nil
	if t.DiffuseBuffer == nil {
		t.DiffuseBuffer = &t.Material.DiffuseColor
	}
	if t.SpecularBuffer == nil {
		t.SpecularBuffer = &t.Material.SpecularColor
	}

	t.HasTexture = (t.Material != nil && t.Material.DiffuseTexture != nil)

	if t.HasTexture {
		// Sample at all three vertices
		uv0 := t.InterpolatedUV(1, 0, 0) // Vertex 0
		uv1 := t.InterpolatedUV(0, 1, 0) // Vertex 1
		uv2 := t.InterpolatedUV(0, 0, 1) // Vertex 2

		color0 := t.Material.DiffuseTexture.Sample(uv0.U, uv0.V)
		color1 := t.Material.DiffuseTexture.Sample(uv1.U, uv1.V)
		color2 := t.Material.DiffuseTexture.Sample(uv2.U, uv2.V)

		// Store as vertex colors
		t.DiffuseBuffer = &lookdev.ColorRGBA{
			R: uint8((float64(color0.R) + float64(color1.R) + float64(color2.R)) / 3),
			G: uint8((float64(color0.G) + float64(color1.G) + float64(color2.G)) / 3),
			B: uint8((float64(color0.B) + float64(color1.B) + float64(color2.B)) / 3),
			A: (color0.A + color1.A + color2.A) / 3,
		}
	} else {
		// Fallback to material color
		t.DiffuseBuffer = &t.Material.DiffuseColor
	}

	t.BufferCache = true
}
