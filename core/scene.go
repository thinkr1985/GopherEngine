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
	ViewAxes       *ViewAxes
	Grid           *Grid
	Lights         []*Light
	Triangles      []*assets.Triangle
	DrawnTriangles int32
	HalfResolution bool
}

func NewScene() *Scene {
	return &Scene{
		Renderer:     NewRenderer3D(),
		Camera:       NewPerspectiveCamera(),
		DefaultLight: NewPointLight(),
		ViewAxes:     NewViewAxes(),
		Grid:         NewGrid(),
	}
}

func (s *Scene) UpdateScene() {

	// Update the Lights
	for _, light := range s.Lights {
		light.Transform.UpdateModelMatrix()
	}

	// Update the Camera
	s.Camera.Transform.UpdateModelMatrix()
	s.Camera.UpdateFrustumPlanes()
}

func (s *Scene) RenderScene() {
	s.UpdateScene()
	for _, obj := range s.Objects {
		if s.Camera.IsVisible(obj.BoundingBox) {
			obj.Update()
			s.Renderer.RenderGeometry(obj)
		}
	}

}
