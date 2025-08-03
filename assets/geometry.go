package assets

import (
	"GopherEngine/lookdev"
	"GopherEngine/nomath"
)

type Geometry struct {
	Name        string
	Transform   *nomath.Transform
	Vertices    []*nomath.Vec3
	Normals     []*nomath.Vec3
	UVs         []*nomath.Vec2
	Triangles   []*Triangle
	BoundingBox *nomath.BoundingBox
	Material    *lookdev.Material
}

func (g *Geometry) NewGeometry() *Geometry {
	geo := &Geometry{
		Name:        "Object001",
		Transform:   nomath.NewTransform(),
		BoundingBox: nomath.NewBoundingBox(),
		Material:    lookdev.NewMaterial("DefaultMaterial"),
	}
	geo.ComputeBoundingBox()
	return geo
}

func (g *Geometry) Update() {
	g.Transform.UpdateModelMatrix()
	g.ComputeTransformedBoundingBox()
}

func (g *Geometry) ComputeBoundingBox() {
	if len(g.Vertices) == 0 {
		return
	}

	min := *g.Vertices[0]
	max := *g.Vertices[0]

	for _, v := range g.Vertices[1:] {
		if v.X < min.X {
			min.X = v.X
		}
		if v.Y < min.Y {
			min.Y = v.Y
		}
		if v.Z < min.Z {
			min.Z = v.Z
		}
		if v.X > max.X {
			max.X = v.X
		}
		if v.Y > max.Y {
			max.Y = v.Y
		}
		if v.Z > max.Z {
			max.Z = v.Z
		}
	}

	g.BoundingBox = &nomath.BoundingBox{Min: min, Max: max}
}

func (g *Geometry) ComputeTransformedBoundingBox() {
	if len(g.Vertices) == 0 {
		return
	}

	transform := g.Transform.GetMatrix()

	first := transform.MultiplyVec4(g.Vertices[0].ToVec4(1)).ToVec3()
	min := first
	max := first

	for _, v := range g.Vertices[1:] {
		tv := transform.MultiplyVec4(v.ToVec4(1)).ToVec3()
		min = nomath.Min(min, tv)
		max = nomath.Max(max, tv)
	}

	g.BoundingBox = &nomath.BoundingBox{Min: min, Max: max}
}

func (g *Geometry) PrecomputeTextureBuffers() {
	for _, tri := range g.Triangles {
		// Initialize buffers
		tri.DiffuseBuffer = lookdev.NewWarningColorRGBA()
		tri.SpecularBuffer = lookdev.NewWarningColorRGBA()
		tri.AlphaBuffer = 1.0 // Default to fully opaque
		// Compute barycentric center for sampling
		u, v, w := 1.0/3.0, 1.0/3.0, 1.0/3.0
		uv := tri.InterpolatedUV(u, v, w)

		// Handle diffuse texture with alpha
		if tri.Material.DiffuseTexture != nil {
			diffuseSample := tri.Material.DiffuseTexture.Sample(uv.U, uv.V)
			*tri.DiffuseBuffer = diffuseSample
			tri.AlphaBuffer = diffuseSample.A // Store alpha from texture

			// If material has alpha, combine it with texture alpha
			if tri.Material.DiffuseColor.A < 1.0 {
				tri.AlphaBuffer *= tri.Material.DiffuseColor.A
			}
		} else {
			*tri.DiffuseBuffer = tri.Material.DiffuseColor
			tri.AlphaBuffer = tri.Material.DiffuseColor.A
		}

		// Handle specular texture
		if tri.Material.SpecularTexture != nil {
			*tri.SpecularBuffer = tri.Material.SpecularTexture.Sample(uv.U, uv.V)
		} else {
			*tri.SpecularBuffer = tri.Material.SpecularColor
		}

		// Apply transparency to diffuse color
		if tri.AlphaBuffer < 1.0 {
			tri.DiffuseBuffer.A = tri.AlphaBuffer
		}
	}
}
