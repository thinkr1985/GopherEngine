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
	Lights         []*Light
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

func (s *Scene) UpdateScene() {
	// Update all geometries in the scene.
	if len(s.Objects) > 1 {
		for _, geometry := range s.Objects {
			geometry.Update()
		}
	}

	// Update the Lights
	if len(s.Lights) > 1 {
		for _, light := range s.Objects {
			light.Transform.UpdateModelMatrix()

		}
	}

	//Updatet the Camera
	s.Camera.Transform.UpdateModelMatrix()

}
