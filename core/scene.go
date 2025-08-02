package core

import (
	"GopherEngine/assets"
)

var SCREEN_WIDTH int = 800
var SCREEN_HEIGHT int = 600

type Scene struct {
	Renderer       *Renderer3D
	Objects        []*assets.Geometry
	Camera         *PerspectiveCamera
	DefaultLight   *Light
	Triangles      []*assets.Triangle
	DrawnTriangles int32
}

func NewScene() *Scene {
	return &Scene{
		Renderer:     NewRenderer3D(),
		Camera:       NewPerspectiveCamera(),
		DefaultLight: NewPointLight(),
	}
}
