package core

import (
	"GopherEngine/assets"
	"GopherEngine/nomath"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
)

var UNIQUE_NAMES []string
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
	Renderer             *Renderer3D
	Assemblies           []*assets.Assembly
	Objects              []*assets.Geometry
	Camera               *PerspectiveCamera
	DefaultLight         *Light
	ViewAxes             *ViewAxes
	Grid                 *Grid
	Lights               []*Light
	Triangles            []*assets.Triangle
	DrawnTriangles       int32
	TotalTriangleCounter int32

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
	// Important to update camera first!
	s.Camera.Update()

	// Update lights
	for _, light := range s.Lights {
		light.Update()
	}

	// Update assemblies
	for _, assembly := range s.Assemblies {
		assembly.Update()
	}

	// Update renderer light directions
	s.Renderer.PreComputeLightDirs(s)
}

func (s *Scene) AddAssembly(assembly *assets.Assembly) {
	s.Assemblies = append(s.Assemblies, assembly)
	s.Triangles = append(s.Triangles, assembly.Triangles...)
}

func (s *Scene) LoadAsset(asset_path string) {
	assembly, err := assets.AssetImport(asset_path)
	if err != nil {
		return
	}
	s.AddAssembly(assembly)

}

func (s *Scene) LoadAssembly(assembly_path string) {
	assembly := assets.NewAssembly()
	assembly.LoadAssembly(assembly_path)
	s.AddAssembly(assembly)

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
		}
	}

	viewDir := s.Camera.Transform.GetForward()
	viewProjMatrix := s.cachedViewProjMatrix

	// Precompute light dot normal per triangle
	for _, assembly := range s.Assemblies {

		if !assembly.IsVisible {
			continue
		}

		for _, geom := range assembly.Geometries {
			if !s.Camera.IsVisible(geom.BoundingBox) || !geom.IsVisible {
				continue
			}

			for _, triangle := range geom.Triangles {
				if triangle.Normal().Dot(viewDir) > 0 || triangle.WorldNormal.Dot(viewDir) > 0 {
					continue
				}
				modelMatrix := geom.Transform.GetMatrix()
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

				s.Renderer.RenderTriangle(&mvpMatrix, &modelMatrix, s.Camera, triangle, s.Lights, s)
				s.DrawnTriangles++
			}
		}
	}
}

func (s *Scene) RenderOnThreads() {
	// First update all scene elements and matrices
	s.UpdateScene()
	atomic.StoreInt32(&s.DrawnTriangles, 0)
	s.Renderer.PreComputeLightDirs(s)

	// Update and cache matrices with write lock
	s.matrixMutex.Lock()
	s.cachedViewMatrix = s.Camera.GetViewMatrix()
	s.cachedProjectionMatrix = s.Camera.GetProjectionMatrix()
	s.cachedViewProjMatrix = s.cachedProjectionMatrix.Multiply(s.cachedViewMatrix)
	s.matrixMutex.Unlock()

	// Clear shadow maps
	var shadowWG sync.WaitGroup
	for _, light := range s.Lights {
		if light.Shadows && light.ShadowMap != nil {
			shadowWG.Add(1)
			go func(sm *ShadowMap) {
				defer shadowWG.Done()
				for y := range sm.Depth {
					row := sm.Depth[y]
					for x := range row {
						row[x] = math.MaxFloat64
					}
				}
			}(light.ShadowMap)
		}
	}
	shadowWG.Wait()

	// Draw overlays
	s.Grid.Draw(s.Renderer, s.Camera)
	s.ViewAxes.Draw(s.Renderer, s.Camera)
	for _, light := range s.Lights {
		light.DrawLight()
	}

	// Get view direction (after matrix updates)
	viewDir := s.Camera.Transform.GetForward()

	// Worker pool setup
	numWorkers := runtime.NumCPU()
	workChan := make(chan RenderTask, numWorkers*4)
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	// Worker function
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			var localCount int32

			for task := range workChan {
				tri := task.Triangle

				// Do shadow map rendering inside the worker
				for li, light := range s.Lights {
					if light.Shadows && light.ShadowMap != nil {
						lightVP := light.ShadowMap.ProjMatrix.Multiply(light.ShadowMap.ViewMatrix)
						s.renderTriangleToShadowMap(tri, lightVP.Multiply(task.ModelMatrix), light)
					}
					task.LightDots[li] = max(0, tri.WorldNormal.Dot(light.GetDirection()))
				}

				// Render the triangle using the correct matrices
				s.Renderer.RenderTriangle(&task.MVP, &task.ModelMatrix, s.Camera, tri, s.Lights, s)
				localCount++
			}
			atomic.AddInt32(&s.DrawnTriangles, localCount)
		}()
	}

	// Traverse scene and stream triangles to workers
	for _, assembly := range s.Assemblies {
		if !assembly.IsVisible {
			continue
		}

		for _, geom := range assembly.Geometries {
			if !geom.IsVisible || !s.Camera.IsVisible(geom.BoundingBox) {
				continue
			}

			modelMatrix := geom.Transform.GetMatrix()
			mvpMatrix := s.cachedViewProjMatrix.Multiply(modelMatrix)
			normalMatrix := modelMatrix.Inverse().Transpose()

			for _, tri := range geom.Triangles {
				// Backface culling
				if tri.Normal().Dot(viewDir) > 0 || tri.WorldNormal.Dot(viewDir) > 0 {
					continue
				}

				// Update world normal
				tri.WorldNormal = normalMatrix.TransformVec3(tri.Normal()).Normalize()

				// Reuse light dots slice
				if cap(tri.LightDotNormals) < len(s.Lights) {
					tri.LightDotNormals = make([]float64, len(s.Lights))
				} else {
					tri.LightDotNormals = tri.LightDotNormals[:len(s.Lights)]
				}

				// Push to worker
				workChan <- RenderTask{
					Triangle:     tri,
					MVP:          mvpMatrix,
					NormalMatrix: normalMatrix,
					ModelMatrix:  modelMatrix,
					LightDots:    tri.LightDotNormals,
				}
			}
		}
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
