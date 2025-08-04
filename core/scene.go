package core

import (
	"GopherEngine/assets"
	"GopherEngine/nomath"
	"runtime"
	"sync"
	"sync/atomic"
)

var SCREEN_WIDTH int = 854
var SCREEN_HEIGHT int = 480

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
	default_light := NewDirectionalLight()

	s := Scene{
		Renderer:     NewRenderer3D(),
		Camera:       NewPerspectiveCamera(),
		DefaultLight: default_light,
		Lights:       []*Light{default_light},
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
	s.Renderer.PreComputeLightDirs(&s)
	return &s
}

func (s *Scene) UpdateScene() {

	// Update the Lights
	for _, light := range s.Lights {
		light.Update()
	}

	// Update the objects
	for _, obj := range s.Objects {
		obj.Update()
		obj.PrecomputeTextureBuffers()
	}

	// Update the Camera
	s.Camera.Update()

	// Precompute Lightdirs
	s.Renderer.PreComputeLightDirs(s)
}

func (s *Scene) AddObject(geom *assets.Geometry) {
	geom.PrecomputeTextureBuffers()
	s.Objects = append(s.Objects, geom)
	s.Triangles = append(s.Triangles, geom.Triangles...)
}

func (s *Scene) RenderScene() {
	s.DrawnTriangles = 0
	s.UpdateScene()

	viewMatrix := s.Camera.GetViewMatrix()
	projectionMatrix := s.Camera.GetProjectionMatrix()
	viewProjMatrix := projectionMatrix.Multiply(viewMatrix)
	viewDir := s.Camera.Transform.GetForward()

	for _, obj := range s.Objects {
		if !s.Camera.IsVisible(obj.BoundingBox) {
			continue
		}

		obj.Update()
		obj.PrecomputeTextureBuffers()

		modelMatrix := obj.Transform.GetMatrix()
		mvpMatrix := viewProjMatrix.Multiply(modelMatrix)

		// Precompute NormalMatrix = inverse transpose of modelMatrix
		normalMatrix := modelMatrix.Inverse().Transpose()

		// Precompute light dot normal per triangle
		for _, triangle := range obj.Triangles {
			// Transform triangle normal using normalMatrix
			worldNormal := normalMatrix.TransformVec3(triangle.Normal()).Normalize()
			triangle.WorldNormal = worldNormal

			// Precompute light dot normal for each light
			triangle.LightDotNormals = make([]float64, len(s.Lights))
			for i, light := range s.Lights {
				lightDir := light.GetDirection() // assuming normalized direction
				triangle.LightDotNormals[i] = max(0, worldNormal.Dot(lightDir))
			}
		}

		for _, triangle := range obj.Triangles {
			if triangle.WorldNormal.Dot(viewDir) > 0 {
				continue // backface
			}

			s.Renderer.RenderTriangle(&mvpMatrix, s.Camera, triangle, s.Lights, s)
			s.DrawnTriangles++
		}
	}
}

/*
func (s *Scene) RenderOnThread() {
	atomic.StoreInt32(&s.DrawnTriangles, 0)
	s.Renderer.PreComputeLightDirs(s)

	viewMatrix := s.Camera.GetViewMatrix()
	projectionMatrix := s.Camera.GetProjectionMatrix()
	viewProjMatrix := projectionMatrix.Multiply(viewMatrix)
	viewDir := s.Camera.Transform.GetForward()

	// Precompute per-object matrices and collect triangles
	type RenderTask struct {
		Triangle *assets.Triangle
		MVP      nomath.Mat4
	}
	var tasks []RenderTask

	for _, obj := range s.Objects {
		obj.Update()
		obj.PrecomputeTextureBuffers()

		modelMatrix := obj.Transform.GetMatrix()
		mvpMatrix := viewProjMatrix.Multiply(modelMatrix)

		if !s.Camera.IsVisible(obj.BoundingBox) {
			continue
		}

		normalMatrix := modelMatrix.Inverse().Transpose()

		for _, triangle := range obj.Triangles {
			if triangle.Normal().Dot(viewDir) > 0 {
				continue
			}

			// Precompute per-vertex lighting
			for i, vert := range [3]*nomath.Vec3{triangle.V0, triangle.V1, triangle.V2} {
				normal := normalMatrix.TransformVec3(vert.Normalize()).Normalize()
				lightDir := s.DefaultLight.Direction.Normalize()
				dot := normal.Dot(lightDir)
				if dot < 0 {
					dot = 0
				}
				triangle.LightDotNormals[i] = dot
			}

			tasks = append(tasks, RenderTask{
				Triangle: triangle,
				MVP:      mvpMatrix,
			})
		}
	}

	numWorkers := runtime.NumCPU()
	var wg sync.WaitGroup
	workChan := make(chan RenderTask, len(tasks))

	worker := func() {
		defer wg.Done()
		var localCount int32

		for task := range workChan {
			s.Renderer.RenderTriangle(&task.MVP, s.Camera, task.Triangle, s.Lights, s)
			localCount++
		}
		atomic.AddInt32(&s.DrawnTriangles, localCount)
	}

	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go worker()
	}

	for _, task := range tasks {
		workChan <- task
	}
	close(workChan)
	wg.Wait()
}
*/

func (s *Scene) RenderOnThread() {
	atomic.StoreInt32(&s.DrawnTriangles, 0)
	s.Renderer.PreComputeLightDirs(s)

	viewMatrix := s.Camera.GetViewMatrix()
	projectionMatrix := s.Camera.GetProjectionMatrix()
	viewProjMatrix := projectionMatrix.Multiply(viewMatrix)
	viewDir := s.Camera.Transform.GetForward()

	// Precompute per-object matrices and collect triangles
	type RenderTask struct {
		Triangle *assets.Triangle
		MVP      nomath.Mat4
	}
	var tasks []RenderTask

	for _, obj := range s.Objects {
		obj.Update()
		obj.PrecomputeTextureBuffers()

		modelMatrix := obj.Transform.GetMatrix()
		mvpMatrix := viewProjMatrix.Multiply(modelMatrix)

		// Skip entire object if not in view
		if !s.Camera.IsVisible(obj.BoundingBox) {
			continue
		}

		for _, triangle := range obj.Triangles {
			// Optional: finer culling per triangle
			if triangle.Normal().Dot(viewDir) > 0 {
				continue
			}
			tasks = append(tasks, RenderTask{
				Triangle: triangle,
				MVP:      mvpMatrix,
			})
		}
	}

	numWorkers := runtime.NumCPU()
	var wg sync.WaitGroup
	workChan := make(chan RenderTask, len(tasks))

	worker := func() {
		defer wg.Done()
		var localCount int32

		for task := range workChan {
			s.Renderer.RenderTriangle(&task.MVP, s.Camera, task.Triangle, s.Lights, s)
			localCount++
		}
		atomic.AddInt32(&s.DrawnTriangles, localCount)
	}

	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go worker()
	}

	for _, task := range tasks {
		workChan <- task
	}
	close(workChan)
	wg.Wait()
}
