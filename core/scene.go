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
	// Resolution scaling settings
	ResolutionScale       float64 // Current scale (1.0 = full, 0.5 = half, etc.)
	AutoResolution        bool    // Whether auto-scaling is enabled
	LastFPS               int     // Track last FPS reading
	MinResolutionScale    float64 // Minimum allowed resolution (e.g., 0.1 for 10%)
	LastScaleChange       float64 // Time since last resolution change (now float64)
	FPSHistory            []int   // Store last few FPS readings for smoothing
	FPSSum                int     // Sum of FPS history for averaging
	TargetResolutionScale float64 // The scale we're gradually moving toward
	ResolutionChangeSpeed float64 // How fast we adjust resolution (0.1 = 10% per second)
}

func NewScene() *Scene {
	return &Scene{
		Renderer:     NewRenderer3D(),
		Camera:       NewPerspectiveCamera(),
		DefaultLight: NewPointLight(),
		ViewAxes:     NewViewAxes(),
		Grid:         NewGrid(),

		// Resolution scaling defaults
		ResolutionScale:       1.0,
		AutoResolution:        false,
		LastScaleChange:       0.0, // Initialize as float64
		MinResolutionScale:    0.1, // Never go below 10%
		FPSHistory:            make([]int, 0, 10),
		TargetResolutionScale: 1.0,
		ResolutionChangeSpeed: 0.25, // Adjust scale by up to 50% per second
	}
}

func (s *Scene) UpdateScene() {

	// Update the Lights
	for _, light := range s.Lights {
		light.Transform.UpdateModelMatrix()
	}

	// Update the objects
	for _, obj := range s.Objects {
		obj.Update()
		// obj.PrecomputeTextureBuffers()
	}

	// Update the Camera
	s.Camera.Transform.UpdateModelMatrix()
	s.Camera.UpdateFrustumPlanes()
}

func (s *Scene) AddObject(geom *assets.Geometry) {
	geom.PrecomputeTextureBuffers()
	s.Objects = append(s.Objects, geom)
	s.Triangles = append(s.Triangles, geom.Triangles...)
}

func (s *Scene) RenderScene() {
	s.DrawnTriangles = 0
	s.UpdateScene()
	for _, triangle := range s.Triangles {
		if s.Camera.IsVisible(triangle.Parent.BoundingBox) {
			s.Renderer.RenderTriangle(s.Camera, triangle, s.Lights, s)
		}
		// fmt.Println(triangle.V0, triangle.V1, triangle.V2)
	}

}
