package core

import (
	"GopherEngine/assets"
	"GopherEngine/lookdev"
	"GopherEngine/nomath"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
)

var SCREEN_WIDTH int = 854
var SCREEN_HEIGHT int = 480

type RenderTask struct {
	Triangle *assets.Triangle
	MVP      nomath.Mat4
	// Add any other frequently used precomputed data here
	NormalMatrix nomath.Mat4 // Optional: for normal transformations
	LightDots    []float64   // Optional: precomputed light factors
	ModelMatrix  nomath.Mat4
}

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

	// caching matrices
	cachedViewMatrix       nomath.Mat4
	cachedProjectionMatrix nomath.Mat4
	cachedViewProjMatrix   nomath.Mat4

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
	matrixMutex           sync.RWMutex
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
	s.Camera.Scene = &s
	return &s
}

func (s *Scene) UpdateScene() {
	// Update camera first
	s.Camera.Update()

	// Update other objects
	for _, light := range s.Lights {
		light.Update()
	}
	for _, obj := range s.Objects {
		obj.Update()
	}
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

	viewDir := s.Camera.Transform.GetForward()
	s.matrixMutex.RLock()
	viewProjMatrix := s.cachedViewProjMatrix
	s.matrixMutex.RUnlock()

	mainBuffers := &ThreadLocalBuffers{
		Framebuffer: s.Renderer.Framebuffer,
		DepthBuffer: s.Renderer.DepthBuffer,
	}

	for _, triangle := range s.Triangles {
		if !s.Camera.IsVisible(triangle.Parent.BoundingBox) ||
			triangle.Normal().Dot(viewDir) > 0 || triangle.WorldNormal.Dot(viewDir) > 0 {
			continue
		}

		modelMatrix := triangle.Parent.Transform.GetMatrix()
		mvpMatrix := viewProjMatrix.Multiply(modelMatrix)
		normalMatrix := modelMatrix.Inverse().Transpose()
		worldNormal := normalMatrix.TransformVec3(triangle.Normal()).Normalize()
		triangle.WorldNormal = worldNormal

		triangle.LightDotNormals = make([]float64, len(s.Lights))
		for i, light := range s.Lights {
			lightDir := light.GetDirection()
			triangle.LightDotNormals[i] = max(0, worldNormal.Dot(lightDir))
		}

		s.Renderer.RenderTriangleToBuffers(&mvpMatrix, s.Camera, triangle, s.Lights, s, mainBuffers)
		s.DrawnTriangles++
	}
}

func (s *Scene) RenderOnThread() {
	s.UpdateScene()
	atomic.StoreInt32(&s.DrawnTriangles, 0)
	s.Renderer.PreComputeLightDirs(s)

	width := s.Renderer.GetWidth()
	height := s.Renderer.GetHeight()

	s.matrixMutex.RLock()
	viewProjMatrix := s.cachedViewProjMatrix
	s.matrixMutex.RUnlock()
	viewDir := s.Camera.Transform.GetForward()

	var tasks []RenderTask

	for _, triangle := range s.Triangles {
		if !s.Camera.IsVisible(triangle.Parent.BoundingBox) {
			continue
		}
		if triangle.Normal().Dot(viewDir) > 0 {
			continue
		}

		modelMatrix := triangle.Parent.Transform.GetMatrix()
		mvpMatrix := viewProjMatrix.Multiply(modelMatrix)

		tasks = append(tasks, RenderTask{
			Triangle: triangle,
			MVP:      mvpMatrix,
		})
	}

	numWorkers := runtime.NumCPU()
	var wg sync.WaitGroup
	workChan := make(chan RenderTask, len(tasks))
	threadBuffers := make([]*ThreadLocalBuffers, numWorkers)

	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		threadBuffers[i] = initThreadBuffers(width, height)
		go func(threadID int) {
			defer wg.Done()
			localBuf := threadBuffers[threadID]
			var localCount int32

			for task := range workChan {
				s.Renderer.RenderTriangleToBuffers(&task.MVP, s.Camera, task.Triangle, s.Lights, s, localBuf)
				localCount++
			}
			atomic.AddInt32(&s.DrawnTriangles, localCount)
		}(i)
	}

	for _, task := range tasks {
		workChan <- task
	}
	close(workChan)
	wg.Wait()

	// Merge all thread buffers into main buffer
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			bestDepth := float32(math.MaxFloat32)
			var bestColor lookdev.ColorRGBA

			for _, buf := range threadBuffers {
				if buf.DepthBuffer[y][x] < bestDepth {
					bestDepth = buf.DepthBuffer[y][x]
					bestColor = buf.Framebuffer[y][x]
				}
			}

			s.Renderer.Framebuffer[y][x] = bestColor
			s.Renderer.DepthBuffer[y][x] = bestDepth
		}
	}
}

func (s *Scene) RenderWithTiling(tileSize int) {
	s.UpdateScene()
	atomic.StoreInt32(&s.DrawnTriangles, 0)
	s.Renderer.PreComputeLightDirs(s)

	width := s.Renderer.GetWidth()
	height := s.Renderer.GetHeight()

	s.matrixMutex.RLock()
	viewProjMatrix := s.cachedViewProjMatrix
	s.matrixMutex.RUnlock()
	viewDir := s.Camera.Transform.GetForward()

	// Group triangles into tasks
	var tasks []RenderTask
	for _, tri := range s.Triangles {
		if !s.Camera.IsVisible(tri.Parent.BoundingBox) ||
			tri.Normal().Dot(viewDir) > 0 {
			continue
		}

		model := tri.Parent.Transform.GetMatrix()
		mvp := viewProjMatrix.Multiply(model)
		normalMatrix := model.Inverse().Transpose()
		worldNormal := normalMatrix.TransformVec3(tri.Normal()).Normalize()
		tri.WorldNormal = worldNormal

		tri.LightDotNormals = make([]float64, len(s.Lights))
		for i, light := range s.Lights {
			tri.LightDotNormals[i] = max(0, worldNormal.Dot(light.GetDirection()))
		}

		tasks = append(tasks, RenderTask{
			Triangle: tri,
			MVP:      mvp,
		})
	}

	numWorkers := runtime.NumCPU()
	var wg sync.WaitGroup
	tileChan := make(chan [2]int, (width/tileSize)*(height/tileSize))

	// Worker: each processes tiles independently
	worker := func() {
		defer wg.Done()

		for tile := range tileChan {
			tx, ty := tile[0], tile[1]
			x0 := tx * tileSize
			y0 := ty * tileSize
			x1 := min(x0+tileSize, width)
			y1 := min(y0+tileSize, height)

			for _, task := range tasks {
				s.Renderer.RenderTriangleClipped(&task.MVP, s.Camera, task.Triangle, s.Lights, s, x0, y0, x1, y1)
			}
		}
	}

	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go worker()
	}

	for y := 0; y < height; y += tileSize {
		for x := 0; x < width; x += tileSize {
			tileChan <- [2]int{x / tileSize, y / tileSize}
		}
	}
	close(tileChan)
	wg.Wait()
}
