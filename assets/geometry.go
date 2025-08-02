package assets

import (
	"GopherEngine/nomath"
)

type Geometry struct {
	Name      string
	Transform *nomath.Transform
	Vertices  []*nomath.Vec3
	Normals   []*nomath.Vec3
	UVs       []*nomath.Vec2
	Triangles []*Triangle
}

func (g *Geometry) NewGeometry() *Geometry {
	return &Geometry{
		Name:      "Object001",
		Transform: nomath.NewTransform(),
	}
}
