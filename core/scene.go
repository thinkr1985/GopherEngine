package core

import (
	"GopherEngine/assets"
	"GopherEngine/nomath"
	"fmt"
	"math"
	"runtime"
	"sync"
	// "sync/atomic"
)

var UNIQUE_NAMES []string
var SCREEN_WIDTH int = 854
var SCREEN_HEIGHT int = 480

type RenderTask struct {
	Triangle       *assets.Triangle
	MVP            nomath.Mat4
	NormalMatrix   nomath.Mat4   // For normal transformations
	ModelMatrix    nomath.Mat4   // For world position calculations
	LightDots      []float64     // Precomputed light factors
	ShadowMatrices []nomath.Mat4 // One per shadow-casting light
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
	// s.DefaultLight.Transform.Rotation.X += 0. + math.Sin(0.2)*0.1
	// s.DefaultLight.Transform.Dirty = true
	// s.Assemblies[0].Geometries[0].Transform.Rotation.Y += 0. + math.Sin(0.2)*0.1
	// s.Assemblies[0].Geometries[0].Transform.Dirty = true
	// Important to update camera first!
	s.Camera.DirtyFrustum = true
	s.Camera.Transform.Dirty = true
	s.matrixMutex.Lock()
	s.Camera.Update()
	s.matrixMutex.Unlock()

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

	if len(assembly.References) > 0 {
		for _, reference := range assembly.References {
			reference.LoadReference()
			s.AddAssembly(reference.Parent)

		}
	}
}

func (s *Scene) LoadAsset(asset_path string) *assets.Assembly {
	assembly, err := assets.AssetImport(asset_path)
	if err != nil {
		fmt.Println("********************* ERROR **********************")
		return assembly
	}
	s.AddAssembly(assembly)
	return assembly

}

func (s *Scene) LoadAssembly(assembly_path string) {
	assembly := assets.NewAssembly()
	assembly.LoadAssembly(assembly_path)
	s.AddAssembly(assembly)

}

func (s *Scene) Render() {
	s.TotalTriangleCounter = 0
	s.DrawnTriangles = 0
	s.UpdateScene()

	// Clear shadow maps (now done in parallel with rendering)
	for _, light := range s.Lights {
		light.DrawLight()
		if light.Shadows && light.ShadowMap != nil {
			// Clear in parallel chunks
			var clearWg sync.WaitGroup
			rowsPerWorker := light.ShadowMap.Height / runtime.NumCPU()

			for i := 0; i < runtime.NumCPU(); i++ {
				clearWg.Add(1)
				start := i * rowsPerWorker
				end := start + rowsPerWorker
				if i == runtime.NumCPU()-1 {
					end = light.ShadowMap.Height
				}

				go func(start, end int) {
					defer clearWg.Done()
					for y := start; y < end; y++ {
						for x := 0; x < light.ShadowMap.Width; x++ {
							light.ShadowMap.Depth[y][x] = math.MaxFloat64
						}
					}
				}(start, end)
			}
			clearWg.Wait()
		}
	}

	// Drawing scene elements
	s.ViewAxes.Draw(s.Renderer, s.Camera)
	viewDir := s.Camera.Transform.GetForward()

	if s.Renderer.MultiThreading {
		s.RenderOnThreads(viewDir)
	} else {
		s.RenderScene(viewDir)
	}
}

func (s *Scene) RenderScene(viewDir nomath.Vec3) {

	viewProjMatrix := s.Camera.cachedViewProjMatrix

	// Precompute light dot normal per triangle
	for _, assembly := range s.Assemblies {
		s.TotalTriangleCounter += int32(len(assembly.Triangles))
		if !assembly.IsVisible {
			continue
		}

		for _, geom := range assembly.Geometries {
			if !s.Camera.IsVisible(geom.BoundingBox) || !geom.IsVisible {
				continue
			}

			modelMatrix := geom.Transform.GetMatrix()
			mvpMatrix := viewProjMatrix.Multiply(modelMatrix)
			normalMatrix := modelMatrix.Inverse().Transpose()

			for _, triangle := range geom.Triangles {
				// Backface culling
				if triangle.Normal().Dot(viewDir) > 0 || triangle.WorldNormal.Dot(viewDir) > 0 {
					continue
				}

				// Transform triangle normal using normalMatrix
				worldNormal := normalMatrix.TransformVec3(triangle.Normal()).Normalize()
				triangle.WorldNormal = worldNormal

				// Precompute light dot normal for each light
				triangle.LightDotNormals = make([]float64, len(s.Lights))
				for i, light := range s.Lights {
					triangle.LightDotNormals[i] = max(0, worldNormal.Dot(s.Renderer.precomputedLightDirs[i]))

					// Render to shadow maps
					if light.Shadows && light.ShadowMap != nil {
						shadowMVP := light.LightVp.Multiply(modelMatrix)
						s.Renderer.renderShadowTriangle(&shadowMVP, triangle, light)
					}
				}

				s.Renderer.RenderTriangle(&mvpMatrix, &modelMatrix, s.Camera, triangle, s.Lights, s)

			}
		}
	}
}

func (s *Scene) RenderOnThreads(viewDir nomath.Vec3) {
	// Worker pool setup
	numWorkers := runtime.NumCPU()
	workChan := make(chan RenderTask, numWorkers*4)
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	// Worker function that handles both main rendering and shadows
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for task := range workChan {
				tri := task.Triangle

				// Render to shadow maps first
				for i, light := range s.Lights {
					if light.Shadows && light.ShadowMap != nil {
						s.Renderer.renderShadowTriangle(&task.ShadowMatrices[i], tri, light)
					}
				}

				// Then render to main framebuffer
				s.Renderer.RenderTriangle(&task.MVP, &task.ModelMatrix, s.Camera, tri, s.Lights, s)
			}
		}()
	}

	// Traverse scene and create tasks
	for _, assembly := range s.Assemblies {
		s.TotalTriangleCounter += int32(len(assembly.Triangles))
		if !assembly.IsVisible {
			continue
		}

		for _, geom := range assembly.Geometries {
			if !geom.IsVisible || !s.Camera.IsVisible(geom.BoundingBox) {
				continue
			}

			modelMatrix := geom.Transform.GetMatrix()
			mvpMatrix := s.Camera.cachedViewProjMatrix.Multiply(modelMatrix)

			// Precompute shadow matrices for all lights
			shadowMatrices := make([]nomath.Mat4, len(s.Lights))
			for i, light := range s.Lights {
				if light.Shadows && light.ShadowMap != nil {
					shadowMatrices[i] = light.LightVp.Multiply(modelMatrix)
				}
			}

			for _, tri := range geom.Triangles {
				// Backface culling
				if tri.Normal().Dot(viewDir) > 0 || tri.WorldNormal.Dot(viewDir) > 0 {
					continue
				}

				workChan <- RenderTask{
					Triangle:       tri,
					MVP:            mvpMatrix,
					ModelMatrix:    modelMatrix,
					ShadowMatrices: shadowMatrices,
				}
			}
		}
	}

	close(workChan)
	wg.Wait()
}
