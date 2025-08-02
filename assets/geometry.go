package assets

import (
	"GopherEngine/nomath"
)

type Geometry struct {
	Name      string
	Transform nomath.Transform
}

func (g *Geometry) NewGeometry() *Geometry {
	return &Geometry{
		Name:      "Object001",
		Transform: *nomath.NewTransform(),
	}
}
