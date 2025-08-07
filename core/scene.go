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
	Triangle     *assets.Triangle
	MVP          nomath.Mat4
	NormalMatrix nomath.Mat4 // For normal transformations
	ModelMatrix  nomath.Mat4 // For world position calculations
	LightDots    []float64   // Precomputed light factors
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

	s := Scene{
		Renderer: NewRenderer3D(),
		Camera:   NewPerspectiveCamera(),
		ViewAxes: NewViewAxes(),
		Grid:     NewGrid(),

		// Resolution scaling defaults
		ResolutionScale:       1.0,
		AutoResolution:        false,
		LastScaleChange:       0.0, // Initialize as float64
		MinResolutionScale:    0.1, // Never go below 10%
		FPSHistory:            make([]int, 0, 10),
		TargetResolutionScale: 1.0,
		ResolutionChangeSpeed: 0.25, // Adjust scale by up to 50% per second
	}

	// Setting a Sun Light
	default_light := NewSunLight(&s)
	default_light.scene = &s
	s.DefaultLight = default_light
	s.Renderer.PreComputeLightDirs(&s)

	s.Camera.Scene = &s
	return &s
}

func (s *Scene) UpdateScene() {

	s.DefaultLight.Transform.Rotation.X += 0.5 + math.Sin(10)*1.0
	s.DefaultLight.Transform.Dirty = true
	s.Renderer.PreComputeLightDirs(s)
	// Update camera first
	s.Camera.Update()

	// Update other objects
	for _, light := range s.Lights {
		light.Update()
	}
	for _, obj := range s.Objects {
		obj.Update()
	}

}

func (s *Scene) AddObject(geom *assets.Geometry) {
	geom.PrecomputeTextureBuffers()
	s.Objects = append(s.Objects, geom)
	s.Triangles = append(s.Triangles, geom.Triangles...)
}

func (s *Scene) RenderScene() {
	s.DrawnTriangles = 0
	s.UpdateScene()

	// Drawing scene elements firsst!
	s.Grid.Draw(s.Renderer, s.Camera)
	s.ViewAxes.Draw(s.Renderer, s.Camera)
	for _, light := range s.Lights {
		light.DrawLight()
	}

	// Rendering a scene
	s.matrixMutex.Lock()
	s.cachedViewMatrix = s.Camera.GetViewMatrix()
	s.cachedProjectionMatrix = s.Camera.GetProjectionMatrix()
	s.cachedViewProjMatrix = s.cachedProjectionMatrix.Multiply(s.cachedViewMatrix)
	s.matrixMutex.Unlock()

	// Render shadow maps first
	for _, light := range s.Lights {
		if light.Shadows {
			s.Renderer.RenderShadowMap(light, s)
			// SaveShadowMapAsImage(light, "shadowmap_debug.png")
		}
	}

	viewDir := s.Camera.Transform.GetForward()
	viewProjMatrix := s.cachedViewProjMatrix

	// Precompute light dot normal per triangle
	for _, triangle := range s.Triangles {
		if !s.Camera.IsVisible(triangle.Parent.BoundingBox) ||
			triangle.Normal().Dot(viewDir) > 0 || triangle.WorldNormal.Dot(viewDir) > 0 {
			continue
		}

		modelMatrix := triangle.Parent.Transform.GetMatrix()
		mvpMatrix := viewProjMatrix.Multiply(modelMatrix)
		normalMatrix := modelMatrix.Inverse().Transpose()
		// Transform triangle normal using normalMatrix
		worldNormal := normalMatrix.TransformVec3(triangle.Normal()).Normalize()
		triangle.WorldNormal = worldNormal

		// Precompute light dot normal for each light
		triangle.LightDotNormals = make([]float64, len(s.Lights))
		for i, light := range s.Lights {
			lightDir := light.GetDirection() // assuming normalized direction
			triangle.LightDotNormals[i] = max(0, worldNormal.Dot(lightDir))
		}

		s.Renderer.RenderTriangle(&mvpMatrix, s.Camera, triangle, s.Lights, s)
		s.DrawnTriangles++
	}
}

func (s *Scene) RenderOnThread() {
	s.UpdateScene()
	atomic.StoreInt32(&s.DrawnTriangles, 0)
	s.Renderer.PreComputeLightDirs(s)

	// Clear all shadow maps first
	for _, light := range s.Lights {
		if light.Shadows && light.ShadowMap != nil {
			for y := 0; y < light.ShadowMap.Height; y++ {
				for x := 0; x < light.ShadowMap.Width; x++ {
					light.ShadowMap.Depth[y][x] = math.MaxFloat64
				}
			}
		}
	}

	// Drawing scene elements first!
	s.Grid.Draw(s.Renderer, s.Camera)
	s.ViewAxes.Draw(s.Renderer, s.Camera)
	for _, light := range s.Lights {
		light.DrawLight()
	}

	// Safely get the view-projection matrix
	s.matrixMutex.RLock()
	viewProjMatrix := s.cachedViewProjMatrix
	s.matrixMutex.RUnlock()
	viewDir := s.Camera.Transform.GetForward()

	var tasks []RenderTask

	// First pass: generate shadow maps and collect render tasks
	for _, triangle := range s.Triangles {
		// Skip entire object if not in view
		if !s.Camera.IsVisible(triangle.Parent.BoundingBox) {
			continue
		}

		// Optional: finer culling per triangle
		if triangle.Normal().Dot(viewDir) > 0 || triangle.WorldNormal.Dot(viewDir) > 0 {
			continue
		}

		modelMatrix := triangle.Parent.Transform.GetMatrix()
		mvpMatrix := viewProjMatrix.Multiply(modelMatrix)
		normalMatrix := modelMatrix.Inverse().Transpose()

		// Transform triangle normal using normalMatrix
		worldNormal := normalMatrix.TransformVec3(triangle.Normal()).Normalize()
		triangle.WorldNormal = worldNormal

		// Precompute light dot normal for each light
		triangle.LightDotNormals = make([]float64, len(s.Lights))
		for i, light := range s.Lights {
			if light.Shadows && light.ShadowMap != nil {
				// Render to shadow map if needed
				lightVP := light.ShadowMap.ProjMatrix.Multiply(light.ShadowMap.ViewMatrix)
				shadowMVP := lightVP.Multiply(modelMatrix)
				s.renderTriangleToShadowMap(triangle, shadowMVP, light)
			}

			lightDir := light.GetDirection()
			triangle.LightDotNormals[i] = max(0, worldNormal.Dot(lightDir))
		}

		tasks = append(tasks, RenderTask{
			Triangle:     triangle,
			MVP:          mvpMatrix,
			NormalMatrix: normalMatrix,
			ModelMatrix:  modelMatrix,
			LightDots:    triangle.LightDotNormals,
		})
	}

	// Clear the framebuffer and depth buffer
	s.Renderer.Clear(lookdev.ColorRGBA{R: 0, G: 0, B: 0, A: 255})

	// Second pass: render scene with parallel workers
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

func (s *Scene) renderTriangleToShadowMap(tri *assets.Triangle, mvpMatrix nomath.Mat4, light *Light) {
	// Transform vertices to clip space
	v0 := mvpMatrix.MultiplyVec4(tri.V0.ToVec4(1.0))
	v1 := mvpMatrix.MultiplyVec4(tri.V1.ToVec4(1.0))
	v2 := mvpMatrix.MultiplyVec4(tri.V2.ToVec4(1.0))

	// Skip triangles that are completely behind the light
	if v0.W <= 0 && v1.W <= 0 && v2.W <= 0 {
		return
	}

	// Perform perspective divide
	ndc0 := v0.Divide(v0.W).ToVec3()
	ndc1 := v1.Divide(v1.W).ToVec3()
	ndc2 := v2.Divide(v2.W).ToVec3()

	// Convert to shadow map coordinates [0,1] range
	v0Screen := nomath.Vec2{
		U: (ndc0.X + 1) * 0.5 * float64(light.ShadowMap.Width),
		V: (1 - (ndc0.Y+1)*0.5) * float64(light.ShadowMap.Height),
	}
	v1Screen := nomath.Vec2{
		U: (ndc1.X + 1) * 0.5 * float64(light.ShadowMap.Width),
		V: (1 - (ndc1.Y+1)*0.5) * float64(light.ShadowMap.Height),
	}
	v2Screen := nomath.Vec2{
		U: (ndc2.X + 1) * 0.5 * float64(light.ShadowMap.Width),
		V: (1 - (ndc2.Y+1)*0.5) * float64(light.ShadowMap.Height),
	}

	// Convert depth from [-1,1] to [0,1] range
	depth0 := (ndc0.Z + 1) * 0.5
	depth1 := (ndc1.Z + 1) * 0.5
	depth2 := (ndc2.Z + 1) * 0.5

	// Find bounding box in shadow map
	minX := max(0, min(int(v0Screen.U), min(int(v1Screen.U), int(v2Screen.U))))
	maxX := min(light.ShadowMap.Width-1, max(int(v0Screen.U), max(int(v1Screen.U), int(v2Screen.U))))
	minY := max(0, min(int(v0Screen.V), min(int(v1Screen.V), int(v2Screen.V))))
	maxY := min(light.ShadowMap.Height-1, max(int(v0Screen.V), max(int(v1Screen.V), int(v2Screen.V))))

	// Rasterize triangle to shadow map
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			p := nomath.Vec2{U: float64(x), V: float64(y)}
			u, v, w := assets.Barycentric(p, v0Screen, v1Screen, v2Screen)

			if u >= 0 && v >= 0 && w >= 0 {
				// Interpolate depth
				depth := u*depth0 + v*depth1 + w*depth2

				// Update shadow map depth if this is closer
				if depth < light.ShadowMap.Depth[y][x] {
					light.ShadowMap.Depth[y][x] = depth
				}
			}
		}
	}
}
