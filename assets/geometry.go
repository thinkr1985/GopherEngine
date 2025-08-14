package assets

import (
	"GopherEngine/lookdev"
	"GopherEngine/nomath"
	"GopherEngine/utilities"
)

type Geometry struct {
	Name        string
	ID          string
	Parent      *Assembly
	Transform   *nomath.Transform
	Vertices    []*nomath.Vec3
	Normals     []*nomath.Vec3
	UVs         []*nomath.Vec2
	Triangles   []*Triangle
	BoundingBox *nomath.BoundingBox
	Material    *lookdev.Material
	IsVisible   bool
	SoftNormals bool
}

func (g *Geometry) NewGeometry() *Geometry {
	geo := &Geometry{
		Name:        utilities.GenerateID(),
		ID:          utilities.GenerateUniqueID(),
		Transform:   nomath.NewTransform(),
		BoundingBox: nomath.NewBoundingBox(),
		Material:    lookdev.NewMaterial("DefaultMaterial"),
		IsVisible:   true,
		SoftNormals: true,
	}
	geo.ComputeBoundingBox()
	return geo
}

func (g *Geometry) Update() {
	if g.Transform.Dirty {
		g.Transform.UpdateModelMatrix()
		g.ComputeTransformedBoundingBox()
		g.Transform.Dirty = false
	}
}

func (g *Geometry) ComputeBoundingBox() {
	if len(g.Vertices) == 0 {
		g.BoundingBox = nomath.NewBoundingBox()
		return
	}

	g.BoundingBox = nomath.NewBoundingBox()
	for _, v := range g.Vertices {
		g.BoundingBox.Expand(*v)
	}
}

func (g *Geometry) ComputeTransformedBoundingBox() {
	if len(g.Vertices) == 0 {
		g.BoundingBox = nomath.NewBoundingBox()
		return
	}

	transform := g.Transform.GetWorldMatrix()
	g.BoundingBox = g.BoundingBox.Transform(transform)
}

func (g *Geometry) PrecomputeTextureBuffers() {
	for _, tri := range g.Triangles {
		if tri.BufferCache {
			continue
		}
		tri.PreComputeBuffers()

	}
}

func (g *Geometry) AddTriangle(tri *Triangle) {
	// Set parent-child relationship
	tri.Transform.SetParent(g.Transform)
	tri.Parent = g

	g.Triangles = append(g.Triangles, tri)
}
