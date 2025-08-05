package core

import (
	"GopherEngine/assets"
	"GopherEngine/lookdev"
	"GopherEngine/nomath"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand/v2"
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
	ambienceColor        lookdev.ColorRGBA

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

	currentX, currentY int // Track current pixel being rendered
}

func NewRenderer3D() *Renderer3D {
	r := &Renderer3D{
		BackFaceCulling: true,
		Framebuffer:     make([][]lookdev.ColorRGBA, SCREEN_HEIGHT),
		DepthBuffer:     make([][]float32, SCREEN_HEIGHT),
		ambienceFactor:  1.0,
		SSAOEnabled:     false,
		SSAORadius:      0.5,
		SSAOBias:        0.025,
		SSAOSamples:     16,
		SSAONoiseScale:  4.0,
	}
	r.ambienceColor = lookdev.ColorRGBA{R: 255, G: 202, B: 138, A: 1} // orange tint
	// Initialize each row
	for y := 0; y < SCREEN_HEIGHT; y++ {
		r.Framebuffer[y] = make([]lookdev.ColorRGBA, SCREEN_WIDTH)
		r.DepthBuffer[y] = make([]float32, SCREEN_WIDTH)

		// Initialize depth buffer with maximum depth
		for x := 0; x < SCREEN_WIDTH; x++ {
			r.DepthBuffer[y][x] = math.MaxFloat32
		}
	}

	// Initialize SSAO kernel
	r.generateSSAOKernel()

	// Initialize SSAO noise texture
	r.generateSSAONoise()
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

func (r *Renderer3D) generateSSAOKernel() {
	r.SSAOKernel = make([]nomath.Vec3, r.SSAOSamples)
	for i := 0; i < r.SSAOSamples; i++ {
		// Generate random samples in hemisphere oriented along Z axis
		sample := nomath.Vec3{
			X: rand.Float64()*2 - 1,
			Y: rand.Float64()*2 - 1,
			Z: rand.Float64(), // Only positive Z (hemisphere)
		}.Normalize()

		// Scale for more samples closer to origin
		scale := float64(i) / float64(r.SSAOSamples)
		scale = lerp(0.1, 1.0, scale*scale)
		sample = sample.Multiply(scale)

		r.SSAOKernel[i] = sample
	}
}

func (r *Renderer3D) generateSSAONoise() {
	// Create a small random rotation texture (4x4)
	noisePixels := make([]lookdev.ColorRGBA, 16)
	for i := 0; i < 16; i++ {
		// Store random rotations in RGB channels
		noisePixels[i] = lookdev.ColorRGBA{
			R: uint8(rand.Float64() * 255),
			G: uint8(rand.Float64() * 255),
			B: 0,
			A: 1.0,
		}
	}

	r.SSAONoiseTexture = &lookdev.Texture{
		Width:  4,
		Height: 4,
		Pixels: noisePixels,
	}
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

func (r *Renderer3D) calculateWorldPosition(x, y int, depth float64, camera *PerspectiveCamera) nomath.Vec3 {
	// Convert screen coordinates to NDC
	ndcX := (2.0 * float64(x) / float64(r.GetWidth())) - 1.0
	ndcY := 1.0 - (2.0 * float64(y) / float64(r.GetHeight()))
	ndcZ := depth*2.0 - 1.0 // Convert back to [-1,1] range

	// Create clip coordinates
	clipPos := nomath.Vec4{X: ndcX, Y: ndcY, Z: ndcZ, W: 1.0}

	// Transform back to world space
	invViewProj := camera.GetProjectionMatrix().Multiply(camera.GetViewMatrix()).Inverse()
	worldPos := invViewProj.MultiplyVec4(clipPos)
	worldPos = worldPos.Multiply(1.0 / worldPos.W) // Perspective divide

	return worldPos.ToVec3()
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
	// --- Transform vertices to clip space ---

	v0 := mvpMatrix.MultiplyVec4(tri.V0.ToVec4(1.0))
	v1 := mvpMatrix.MultiplyVec4(tri.V1.ToVec4(1.0))
	v2 := mvpMatrix.MultiplyVec4(tri.V2.ToVec4(1.0))

	// --- Perspective division (now in NDC [-1, 1]) ---
	v0 = v0.Multiply(1.0 / v0.W)
	v1 = v1.Multiply(1.0 / v1.W)
	v2 = v2.Multiply(1.0 / v2.W)

	// --- Backface culling ---
	normal := tri.Normal()
	viewDir := camera.Transform.GetForward()
	if r.BackFaceCulling {
		if normal.Dot(viewDir) > 0 {
			return // Skip back faces
		}
	}
	scene.DrawnTriangles += 1
	// --- Convert to screen coordinates ---
	x0, y0 := r.NDCToScreen(v0.ToVec3())
	x1, y1 := r.NDCToScreen(v1.ToVec3())
	x2, y2 := r.NDCToScreen(v2.ToVec3())

	// --- Bounding box setup ---
	minX := max(0, min(x0, min(x1, x2)))
	maxX := min(r.GetWidth()-1, max(x0, max(x1, x2)))
	minY := max(0, min(y0, min(y1, y2)))
	maxY := min(r.GetHeight()-1, max(y0, max(y1, y2)))

	// Precompute screen-space vertices for barycentric coords
	v0Screen := nomath.Vec2{U: float64(x0), V: float64(y0)}
	v1Screen := nomath.Vec2{U: float64(x1), V: float64(y1)}
	v2Screen := nomath.Vec2{U: float64(x2), V: float64(y2)}

	// --- Rasterize the triangle ---
	for y := minY; y <= maxY; y++ {
		r.currentY = y
		for x := minX; x <= maxX; x++ {
			r.currentX = x
			// Convert screen coords to NDC for barycentric
			// ndcX := (2.0 * float64(x) / float64(r.GetWidth())) - 1.0
			// ndcY := 1.0 - (2.0 * float64(y) / float64(r.GetHeight()))
			p := nomath.Vec2{U: float64(x), V: float64(y)}

			// Calculate barycentric coordinates
			u, v, w := tri.Barycentric(p, v0Screen, v1Screen, v2Screen)

			// Check if pixel is inside the triangle
			if u >= 0 && v >= 0 && w >= 0 {
				// print("Depth Test triggered...")
				// Calculate interpolated depth in NDC space [-1,1]
				depth := u*v0.Z + v*v1.Z + w*v2.Z

				// Convert to [0,1] range for depth buffer storage
				// Near plane maps to 0, far plane maps to 1
				depth = (depth + 1) * 0.5

				// Depth test (note reversed comparison for depth buffer)
				if depth >= 0 && depth <= 1 && depth < float64(r.DepthBuffer[y][x]) {
					color := r.calculateLighting(tri, normal, viewDir)
					r.safeSetPixel(x, y, *color)
					r.DepthBuffer[y][x] = float32(depth)
				}
				if depth >= 0 && depth <= 1 {
					gray := uint8(depth * 255)
					debugColor := lookdev.ColorRGBA{R: gray, G: gray, B: gray, A: 1}
					r.Framebuffer[y][x] = debugColor
					r.DepthBuffer[y][x] = float32(depth)
				}
			}
			// calculate lighting
			// color := r.calculateLighting(tri, normal, viewDir)
			// r.Framebuffer[y][x] = *color
			// r.DepthBuffer[y][x] = float32(1)
		}
	}
	// After rendering all triangles, calculate SSAO
	if r.SSAOEnabled {
		r.calculateSSAO(camera)
	}
}

func (r *Renderer3D) calculateLighting(
	triangle *assets.Triangle, normal nomath.Vec3, viewDir nomath.Vec3) *lookdev.ColorRGBA {
	diffuseColor := triangle.DiffuseBuffer
	// diffuseColor := triangle.Material.DiffuseColor
	// Apply ambient occlusion
	// ambientOcclusion := 1.0
	// if r.SSAOEnabled {
	// 	ambientOcclusion = float64(r.SSAOBuffer[r.currentY][r.currentX])
	// }

	// // Calculate ambient with occlusion
	// ambient := lookdev.ColorRGBA{
	// 	R: uint8(float64(r.ambienceColor.R) * r.ambienceFactor * ambientOcclusion),
	// 	G: uint8(float64(r.ambienceColor.G) * r.ambienceFactor * ambientOcclusion),
	// 	B: uint8(float64(r.ambienceColor.B) * r.ambienceFactor * ambientOcclusion),
	// }

	// Add diffuse and specular from each light
	for _, lightDir := range r.precomputedLightDirs {
		// Diffuse
		diffuseFactor := math.Max(0, normal.Dot(lightDir))
		diffuseColor.R += uint8(float64(diffuseColor.R) * diffuseFactor)
		diffuseColor.G += uint8(float64(diffuseColor.G) * diffuseFactor)
		diffuseColor.B += uint8(float64(diffuseColor.B) * diffuseFactor)

		// Specular (Blinn-Phong)
		halfDir := lightDir.Add(viewDir).Normalize()
		specFactor := math.Pow(math.Max(0, normal.Dot(halfDir)), float64(triangle.Material.Shininess))
		diffuseColor.R += uint8(float64(triangle.Material.SpecularColor.R) * specFactor)
		diffuseColor.G += uint8(float64(triangle.Material.SpecularColor.G) * specFactor)
		diffuseColor.B += uint8(float64(triangle.Material.SpecularColor.B) * specFactor)
	}
	// Calculating overall ambience factor
	// diffuseColor.R += uint8(float64(ambient.R))
	// diffuseColor.G += uint8(float64(ambient.G))
	// diffuseColor.B += uint8(float64(ambient.B))

	// Clamp color values
	diffuseColor.R = min(diffuseColor.R, 255)
	diffuseColor.G = min(diffuseColor.G, 255)
	diffuseColor.B = min(diffuseColor.B, 255)

	return diffuseColor
}
func (r *Renderer3D) calculateSSAO(camera *PerspectiveCamera) {
	width := r.GetWidth()
	height := r.GetHeight()

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if r.DepthBuffer[y][x] >= 1.0 {
				r.SSAOBuffer[y][x] = 1.0 // No occlusion for background
				continue
			}

			// Get world position
			pos := r.calculateWorldPosition(x, y, float64(r.DepthBuffer[y][x]), camera)

			// Reconstruct normal from depth buffer
			normal := r.reconstructNormal(x, y, camera)

			// Get random rotation from noise texture
			noiseVec := r.getSSAONoise(x, y)

			// Create tangent space matrix with noise rotation
			tangent := nomath.Vec3{X: 1, Y: 0, Z: 0}.Cross(normal).Normalize()
			tangent = tangent.Add(noiseVec.Multiply(0.5)).Normalize() // Apply noise
			bitangent := normal.Cross(tangent).Normalize()
			tbn := nomath.Mat3{
				tangent.X, bitangent.X, normal.X,
				tangent.Y, bitangent.Y, normal.Y,
				tangent.Z, bitangent.Z, normal.Z,
			}

			occlusion := 0.0
			for i := 0; i < r.SSAOSamples; i++ {
				// Rotate sample to normal's hemisphere with noise variation
				sample := tbn.MultiplyVec3(r.SSAOKernel[i])
				sample = pos.Add(sample.Multiply(r.SSAORadius))

				// Project sample position to screen space
				sampleScreen := r.worldToScreen(sample, camera)
				sx, sy := int(sampleScreen.X), int(sampleScreen.Y)

				// Check bounds
				if sx < 0 || sx >= width || sy < 0 || sy >= height {
					continue
				}

				// Get sample depth
				sampleDepth := float64(r.DepthBuffer[sy][sx])
				if sampleDepth >= 1.0 {
					continue
				}

				// Get sample position
				samplePos := r.calculateWorldPosition(sx, sy, sampleDepth, camera)

				// Range check and accumulate
				sampleDelta := samplePos.Subtract(pos)
				sampleDist := sampleDelta.Length()
				rangeCheck := smoothstep(0.0, 1.0, r.SSAORadius/sampleDist)
				occlusion += step(samplePos.Z, pos.Z-r.SSAOBias) * rangeCheck
			}

			occlusion = 1.0 - (occlusion / float64(r.SSAOSamples))
			r.SSAOBuffer[y][x] = float32(occlusion)
		}
	}

	// Apply blur to reduce noise
	r.blurSSAOBuffer()
}

// reconstructNormal estimates normal from depth buffer using a cross pattern
func (r *Renderer3D) reconstructNormal(x, y int, camera *PerspectiveCamera) nomath.Vec3 {
	width := r.GetWidth()
	height := r.GetHeight()

	// Get neighboring pixels with boundary checks
	x0 := max(0, x-1)
	x1 := min(width-1, x+1)
	y0 := max(0, y-1)
	y1 := min(height-1, y+1)

	// Get world positions for center and 4 neighbors
	center := r.calculateWorldPosition(x, y, float64(r.DepthBuffer[y][x]), camera)

	// Only proceed if we have valid depth values around us
	if r.DepthBuffer[y][x0] >= 1.0 || r.DepthBuffer[y][x1] >= 1.0 ||
		r.DepthBuffer[y0][x] >= 1.0 || r.DepthBuffer[y1][x] >= 1.0 {
		return nomath.Vec3{Z: 1} // Fallback to facing forward
	}

	left := r.calculateWorldPosition(x0, y, float64(r.DepthBuffer[y][x0]), camera)
	right := r.calculateWorldPosition(x1, y, float64(r.DepthBuffer[y][x1]), camera)
	top := r.calculateWorldPosition(x, y0, float64(r.DepthBuffer[y0][x]), camera)
	bottom := r.calculateWorldPosition(x, y1, float64(r.DepthBuffer[y1][x]), camera)

	// Calculate horizontal and vertical differences
	dx := right.Subtract(left)
	dy := bottom.Subtract(top)

	// Cross product to get normal (flipped for correct facing)
	normal := dy.Cross(dx).Normalize()

	// Ensure normal faces the camera (consistent winding)
	viewDir := center.Subtract(camera.Transform.Position).Normalize()
	if normal.Dot(viewDir) < 0 {
		normal = normal.Negate()
	}

	return normal
}

func (r *Renderer3D) getSSAONoise(x, y int) nomath.Vec3 {
	// Scale coordinates by noise scale and wrap
	nx := int(float64(x)/r.SSAONoiseScale) % r.SSAONoiseTexture.Width
	ny := int(float64(y)/r.SSAONoiseScale) % r.SSAONoiseTexture.Height
	if nx < 0 {
		nx += r.SSAONoiseTexture.Width
	}
	if ny < 0 {
		ny += r.SSAONoiseTexture.Height
	}

	// Get noise value and remap to [-1,1]
	noise := r.SSAONoiseTexture.Pixels[ny*r.SSAONoiseTexture.Width+nx]
	return nomath.Vec3{
		X: float64(noise.R)/127.5 - 1.0,
		Y: float64(noise.G)/127.5 - 1.0,
		Z: 0,
	}.Normalize()
}

func (r *Renderer3D) blurSSAOBuffer() {
	width, height := r.GetWidth(), r.GetHeight()
	temp := make([][]float32, height)

	// Horizontal pass
	for y := 0; y < height; y++ {
		temp[y] = make([]float32, width)
		for x := 1; x < width-1; x++ {
			temp[y][x] = (r.SSAOBuffer[y][x-1] +
				r.SSAOBuffer[y][x]*2 +
				r.SSAOBuffer[y][x+1]) / 4
		}
	}

	// Vertical pass
	for y := 1; y < height-1; y++ {
		for x := 0; x < width; x++ {
			r.SSAOBuffer[y][x] = (temp[y-1][x] +
				temp[y][x]*2 +
				temp[y+1][x]) / 4
		}
	}
}

func (r *Renderer3D) worldToScreen(pos nomath.Vec3, camera *PerspectiveCamera) nomath.Vec3 {
	viewProj := camera.GetProjectionMatrix().Multiply(camera.GetViewMatrix())
	clipPos := viewProj.MultiplyVec4(pos.ToVec4(1.0))
	clipPos = clipPos.Multiply(1.0 / clipPos.W) // Perspective divide

	// Convert to screen coordinates
	return nomath.Vec3{
		X: (clipPos.X + 1.0) * 0.5 * float64(r.GetWidth()),
		Y: (1.0 - (clipPos.Y+1.0)*0.5) * float64(r.GetHeight()),
		Z: clipPos.Z,
	}
}

func lerp(a, b, t float64) float64 {
	return a + t*(b-a)
}

func smoothstep(edge0, edge1, x float64) float64 {
	t := clamp((x-edge0)/(edge1-edge0), 0.0, 1.0)
	return t * t * (3.0 - 2.0*t)
}

func clamp(x, min, max float64) float64 {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}

func step(edge, x float64) float64 {
	if x < edge {
		return 0.0
	}
	return 1.0
}

func (r *Renderer3D) safeSetPixel(x, y int, color lookdev.ColorRGBA) {
	r.bufferMutex.Lock()
	defer r.bufferMutex.Unlock()

	if x >= 0 && x < r.GetWidth() && y >= 0 && y < r.GetHeight() {
		r.Framebuffer[y][x] = color
	}
}
