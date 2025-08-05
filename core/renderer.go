package core

import (
	"GopherEngine/assets"
	"GopherEngine/lookdev"
	"GopherEngine/nomath"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"sync"
)

type Renderer3D struct {
	Framebuffer          [][]lookdev.ColorRGBA // Changed to value type
	DepthBuffer          [][]float32           // Changed to float32 for better cache usage
	BackFaceCulling      bool
	bufferMutex          sync.Mutex // For thread-safe resizing
	precomputedLightDirs []nomath.Vec3
	ambienceFactor       float64

	CachedRGBA   []color.RGBA
	cachedWidth  int
	cachedHeight int

	SSAOEnabled      bool
	SSAOBuffer       [][]float32 // Stores ambient occlusion values
	SSAOKernel       []nomath.Vec3
	SSAORadius       float64
	SSAOBias         float64
	SSAOSamples      int
	SSAONoiseTexture *lookdev.Texture
	SSAONoiseScale   float64

	rowLocks []sync.Mutex // NEW: One mutex per row
}

func NewRenderer3D() *Renderer3D {
	r := &Renderer3D{
		BackFaceCulling: true,
		Framebuffer:     make([][]lookdev.ColorRGBA, SCREEN_HEIGHT),
		DepthBuffer:     make([][]float32, SCREEN_HEIGHT),
		rowLocks:        make([]sync.Mutex, SCREEN_HEIGHT), // INIT ROW LOCKS
		ambienceFactor:  1.0,
	}
	// Init buffers
	for y := 0; y < SCREEN_HEIGHT; y++ {
		r.Framebuffer[y] = make([]lookdev.ColorRGBA, SCREEN_WIDTH)
		r.DepthBuffer[y] = make([]float32, SCREEN_WIDTH)
		for x := 0; x < SCREEN_WIDTH; x++ {
			r.DepthBuffer[y][x] = math.MaxFloat32
		}
	}
	return r
}

func (r *Renderer3D) GetWidth() int {
	if len(r.Framebuffer) == 0 {
		return 0
	}
	return len(r.Framebuffer[0])
}

func (r *Renderer3D) GetHeight() int {
	return len(r.Framebuffer)
}

func (r *Renderer3D) Resize(width, height int) {
	r.bufferMutex.Lock()
	defer r.bufferMutex.Unlock()

	// Ensure minimum size
	width = max(1, width)
	height = max(1, height)

	// Create new framebuffer
	newFramebuffer := make([][]lookdev.ColorRGBA, height)
	for y := 0; y < height; y++ {
		newFramebuffer[y] = make([]lookdev.ColorRGBA, width)
	}

	// Create new depth buffer
	newDepthBuffer := make([][]float32, height)
	for y := 0; y < height; y++ {
		newDepthBuffer[y] = make([]float32, width)
		for x := 0; x < width; x++ {
			newDepthBuffer[y][x] = math.MaxFloat32
		}
	}

	// Create new SSAO buffer if enabled
	if r.SSAOEnabled {
		r.SSAOBuffer = make([][]float32, height)
		for y := 0; y < height; y++ {
			r.SSAOBuffer[y] = make([]float32, width)
		}
	}

	// Update cached RGBA buffer
	r.cachedWidth = width
	r.cachedHeight = height
	r.CachedRGBA = make([]color.RGBA, width*height)

	// Atomic swap of buffers
	r.Framebuffer = newFramebuffer
	r.DepthBuffer = newDepthBuffer

	r.rowLocks = make([]sync.Mutex, height) // When resizing
}

func (r *Renderer3D) Clear(color lookdev.ColorRGBA) {
	// Use the actual renderer dimensions, not SCREEN_WIDTH/HEIGHT
	width := r.GetWidth()
	height := r.GetHeight()

	for y := 0; y < height; y++ {
		rowPixels := r.Framebuffer[y]
		rowDepth := r.DepthBuffer[y]

		for x := 0; x < width; x++ {
			rowPixels[x] = color
			rowDepth[x] = math.MaxFloat32
		}
	}
}

type DirtyRect struct {
	X1, Y1, X2, Y2 int
}

func (r *Renderer3D) ToImage() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, r.GetWidth(), r.GetHeight()))
	for y := 0; y < r.GetHeight(); y++ {
		for x := 0; x < r.GetWidth(); x++ {
			c := r.Framebuffer[y][x]
			img.SetRGBA(x, y, color.RGBA{
				R: c.R,
				G: c.G,
				B: c.B,
				A: 255,
			})
		}
	}
	return img
}

func (r *Renderer3D) SaveToPNG(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, r.ToImage())
}

func (r *Renderer3D) DrawText2D(text string, x, y int, color *lookdev.ColorRGBA) {
	// Basic text rendering implementation
	// You'll need to implement proper text rendering
	// For now, just draw a marker
	r.DrawLine2D(x, y, x+5, y+5, color)
	r.DrawLine2D(x+5, y, x, y+5, color)
}

// NDCToScreen optimized version with proper return signature
func (r *Renderer3D) NDCToScreen(ndc nomath.Vec3) (int, int) {
	x := int((ndc.X + 1) * 0.5 * float64(r.GetWidth()))
	y := int((1 - (ndc.Y+1)*0.5) * float64(r.GetHeight()))
	return x, y
}

func (r *Renderer3D) DrawLine3D(p0, p1 nomath.Vec3, camera *PerspectiveCamera, color *lookdev.ColorRGBA) {
	// Precompute matrices once
	viewMatrix := camera.GetViewMatrix()
	projectionMatrix := camera.GetProjectionMatrix()
	nearPlane := camera.NearPlane

	// Transform points
	viewP0 := viewMatrix.MultiplyVec4(p0.ToVec4(1.0))
	viewP1 := viewMatrix.MultiplyVec4(p1.ToVec4(1.0))
	clip0 := projectionMatrix.MultiplyVec4(viewP0)
	clip1 := projectionMatrix.MultiplyVec4(viewP1)

	// Early rejection
	if clip0.W <= 0 && clip1.W <= 0 {
		return
	}

	// Handle perspective division with clipping
	var ndc0, ndc1 nomath.Vec3
	if clip0.W > 0 {
		ndc0 = clip0.ToVec3()
	} else {
		t := (nearPlane - viewP0.Z) / (viewP1.Z - viewP0.Z)
		ndc0 = nomath.Vec3{
			X: (viewP0.X + t*(viewP1.X-viewP0.X)) / nearPlane,
			Y: (viewP0.Y + t*(viewP1.Y-viewP0.Y)) / nearPlane,
			Z: 1.0,
		}
	}

	if clip1.W > 0 {
		ndc1 = clip1.ToVec3()
	} else {
		t := (nearPlane - viewP0.Z) / (viewP1.Z - viewP0.Z)
		ndc1 = nomath.Vec3{
			X: (viewP0.X + t*(viewP1.X-viewP0.X)) / nearPlane,
			Y: (viewP0.Y + t*(viewP1.Y-viewP0.Y)) / nearPlane,
			Z: 1.0,
		}
	}

	// Convert to screen coordinates
	x0, y0 := r.NDCToScreen(ndc0)
	x1, y1 := r.NDCToScreen(ndc1)

	// Early exit if line is completely off-screen
	if (x0 < 0 && x1 < 0) || (x0 >= SCREEN_WIDTH && x1 >= SCREEN_WIDTH) ||
		(y0 < 0 && y1 < 0) || (y0 >= SCREEN_HEIGHT && y1 >= SCREEN_HEIGHT) {
		return
	}

	// Draw the line
	r.DrawLine2D(x0, y0, x1, y1, color)
}

func (r *Renderer3D) DrawLine2D(x0, y0, x1, y1 int, color *lookdev.ColorRGBA) {
	// Get current renderer dimensions
	width := r.GetWidth()
	height := r.GetHeight()

	// Early bounds check using actual renderer dimensions
	if (x0 < 0 && x1 < 0) || (x0 >= width && x1 >= width) ||
		(y0 < 0 && y1 < 0) || (y0 >= height && y1 >= height) {
		return
	}

	// Early exit if points are the same
	if x0 == x1 && y0 == y1 {
		if x0 >= 0 && x0 < width && y0 >= 0 && y0 < height {
			r.Framebuffer[y0][x0] = *color
		}
		return
	}

	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	sy := -1
	if y0 < y1 {
		sy = 1
	}

	err := dx - dy
	maxIterations := dx + dy + 1

	for i := 0; i < maxIterations; i++ {
		// Check bounds using actual dimensions
		if x0 >= 0 && x0 < width && y0 >= 0 && y0 < height {
			r.safeSetPixel(x0, y0, *color)
		}

		if x0 == x1 && y0 == y1 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// Helper function for absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (r *Renderer3D) PreComputeLightDirs(s *Scene) {
	r.precomputedLightDirs = make([]nomath.Vec3, len(s.Lights))
	for i, light := range s.Lights {
		if light.Transform.Dirty {
			if light.Type == LightTypeDirectional {
				r.precomputedLightDirs[i] = light.Transform.Position.Normalize().Negate()
			} else {
				r.precomputedLightDirs[i] = light.Transform.Position
			}
		}
	}
}
func (r *Renderer3D) RenderTriangle(mvpMatrix *nomath.Mat4, camera *PerspectiveCamera, tri *assets.Triangle, lights []*Light, scene *Scene) {
	nearPlane := camera.NearPlane

	// Transform vertices to clip space
	v0 := mvpMatrix.MultiplyVec4(tri.V0.ToVec4(1.0))
	v1 := mvpMatrix.MultiplyVec4(tri.V1.ToVec4(1.0))
	v2 := mvpMatrix.MultiplyVec4(tri.V2.ToVec4(1.0))

	// Store in array for easier indexing
	clipVerts := [3]nomath.Vec4{v0, v1, v2}
	screenVerts := [3]nomath.Vec3{}

	// Count how many vertices are in front of the near plane
	inFront := [3]bool{}
	numInFront := 0
	for i := 0; i < 3; i++ {
		if clipVerts[i].Z > -nearPlane {
			inFront[i] = true
			numInFront++
		}
	}

	// If all behind, skip
	if numInFront == 0 {
		return
	}

	// If all in front, proceed with regular rasterization
	if numInFront == 3 {
		for i := 0; i < 3; i++ {
			screenVerts[i] = clipVerts[i].ToVec3()
		}
		r.rasterizeTriangle(screenVerts, tri, lights, camera)
		return
	}

	// Otherwise, clip against near plane and reconstruct 1 or 2 triangles
	var newVerts []nomath.Vec3

	getIntersect := func(a, b nomath.Vec4) nomath.Vec3 {
		t := (-nearPlane - a.Z) / (b.Z - a.Z)
		interp := a.Add(b.Sub(a).Multiply(t))
		return interp.Divide(interp.W).ToVec3()
	}

	for i := 0; i < 3; i++ {
		curr := clipVerts[i]
		next := clipVerts[(i+1)%3]

		currIn := inFront[i]
		nextIn := inFront[(i+1)%3]

		if currIn {
			// Keep current vertex
			newVerts = append(newVerts, curr.ToVec3())
		}
		if currIn != nextIn {
			// Edge crosses near plane — compute intersection
			newVerts = append(newVerts, getIntersect(curr, next))
		}
	}

	// Rasterize new triangle(s)
	if len(newVerts) < 3 {
		return // degenerate
	}
	if len(newVerts) == 3 {
		r.rasterizeTriangle([3]nomath.Vec3{newVerts[0], newVerts[1], newVerts[2]}, tri, lights, camera)
	} else if len(newVerts) == 4 {
		// Split quad into 2 triangles
		r.rasterizeTriangle([3]nomath.Vec3{newVerts[0], newVerts[1], newVerts[2]}, tri, lights, camera)
		r.rasterizeTriangle([3]nomath.Vec3{newVerts[0], newVerts[2], newVerts[3]}, tri, lights, camera)
	}
}

func (r *Renderer3D) rasterizeTriangle(verts [3]nomath.Vec3, tri *assets.Triangle, lights []*Light, camera *PerspectiveCamera) {
	x0, y0 := r.NDCToScreen(verts[0])
	x1, y1 := r.NDCToScreen(verts[1])
	x2, y2 := r.NDCToScreen(verts[2])

	minX := max(0, min(x0, min(x1, x2)))
	maxX := min(r.GetWidth()-1, max(x0, max(x1, x2)))
	minY := max(0, min(y0, min(y1, y2)))
	maxY := min(r.GetHeight()-1, max(y0, max(y1, y2)))

	if minX > maxX || minY > maxY {
		return
	}

	v0Screen := nomath.Vec2{U: float64(x0), V: float64(y0)}
	v1Screen := nomath.Vec2{U: float64(x1), V: float64(y1)}
	v2Screen := nomath.Vec2{U: float64(x2), V: float64(y2)}

	depth0 := (verts[0].Z + 1) * 0.5
	depth1 := (verts[1].Z + 1) * 0.5
	depth2 := (verts[2].Z + 1) * 0.5

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			p := nomath.Vec2{U: float64(x), V: float64(y)}
			u, v, w := tri.Barycentric(p, v0Screen, v1Screen, v2Screen)

			if u >= 0 && v >= 0 && w >= 0 {
				depth := u*depth0 + v*depth1 + w*depth2
				if depth >= 0 && depth <= 1 && depth < float64(r.DepthBuffer[y][x]) {
					var color *lookdev.ColorRGBA
					if len(tri.LightDotNormals) == len(lights) {
						color = r.calculateLightingWithPrecomputed(tri, lights)
					} else {
						color = r.calculateLighting(tri, tri.WorldNormal, camera.Transform.GetForward(), lights)
					}
					r.safeSetPixel(x, y, *color)
					r.DepthBuffer[y][x] = float32(depth)
				}
			}
		}
	}
}

func (r *Renderer3D) calculateLightingWithPrecomputed(tri *assets.Triangle, lights []*Light) *lookdev.ColorRGBA {
	result := *tri.DiffuseBuffer

	// Apply precomputed lighting factors
	for i, dot := range tri.LightDotNormals {
		if i >= len(lights) {
			break
		}
		intensity := float64(lights[i].Intensity) / 255.0
		result.R = min(255, result.R+uint8(float64(tri.DiffuseBuffer.R)*dot*intensity))
		result.G = min(255, result.G+uint8(float64(tri.DiffuseBuffer.G)*dot*intensity))
		result.B = min(255, result.B+uint8(float64(tri.DiffuseBuffer.B)*dot*intensity))
	}

	// Apply specular if available
	if tri.SpecularBuffer != nil {
		result.R = min(255, result.R+tri.SpecularBuffer.R)
		result.G = min(255, result.G+tri.SpecularBuffer.G)
		result.B = min(255, result.B+tri.SpecularBuffer.B)
	}

	return &result
}

func (r *Renderer3D) calculateLighting(tri *assets.Triangle, normal nomath.Vec3, viewDir nomath.Vec3, lights []*Light) *lookdev.ColorRGBA {
	result := *tri.DiffuseBuffer

	for _, light := range lights {
		lightDir := light.GetDirection()
		diffuseFactor := math.Max(0, normal.Dot(lightDir))
		intensity := float64(light.Intensity) / 255.0

		result.R = min(255, result.R+uint8(float64(tri.DiffuseBuffer.R)*diffuseFactor*intensity))
		result.G = min(255, result.G+uint8(float64(tri.DiffuseBuffer.G)*diffuseFactor*intensity))
		result.B = min(255, result.B+uint8(float64(tri.DiffuseBuffer.B)*diffuseFactor*intensity))

		// Specular (Blinn-Phong)
		halfDir := lightDir.Add(viewDir).Normalize()
		specFactor := math.Pow(math.Max(0, normal.Dot(halfDir)), float64(tri.Material.Shininess))
		if tri.SpecularBuffer != nil {
			result.R = min(255, result.R+uint8(float64(tri.SpecularBuffer.R)*specFactor*intensity))
			result.G = min(255, result.G+uint8(float64(tri.SpecularBuffer.G)*specFactor*intensity))
			result.B = min(255, result.B+uint8(float64(tri.SpecularBuffer.B)*specFactor*intensity))
		}
	}

	return &result
}

func (r *Renderer3D) safeSetPixel(x, y int, color lookdev.ColorRGBA) {
	if x < 0 || x >= r.GetWidth() || y < 0 || y >= r.GetHeight() {
		return
	}

	r.rowLocks[y].Lock()
	r.Framebuffer[y][x] = color
	r.rowLocks[y].Unlock()
}
