package assets

import (
	"GopherEngine/lookdev"
	"GopherEngine/nomath"
	"GopherEngine/utilities"
)

type Geometry struct {
	Name        string
	ID          string
	Transform   *nomath.Transform
	Vertices    []*nomath.Vec3
	Normals     []*nomath.Vec3
	UVs         []*nomath.Vec2
	Triangles   []*Triangle
	BoundingBox *nomath.BoundingBox
	Material    *lookdev.Material
	IsVisible   bool
}

func (g *Geometry) NewGeometry() *Geometry {
	geo := &Geometry{
		Name:        utilities.GenerateID(),
		ID:          utilities.GenerateUniqueID(),
		Transform:   nomath.NewTransform(),
		BoundingBox: nomath.NewBoundingBox(),
		Material:    lookdev.NewMaterial("DefaultMaterial"),
		IsVisible:   true,
	}
	geo.ComputeBoundingBox()
	return geo
}

func (g *Geometry) Update() {
	if g.Transform.Dirty {
		g.Transform.Mutex.Lock()
		defer g.Transform.Mutex.Unlock()
		g.Transform.UpdateModelMatrix()
		g.ComputeTransformedBoundingBox()
		g.Transform.Dirty = false
	}
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
		if tri.BufferCache {
			continue
		}
		tri.PreComputeBuffers()

	}
}
