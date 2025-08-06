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

	rowLocks []sync.Mutex // NEW: One mutex per row
}

func NewRenderer3D() *Renderer3D {
	r := &Renderer3D{
		BackFaceCulling: true,
		Framebuffer:     make([][]lookdev.ColorRGBA, SCREEN_HEIGHT),
		DepthBuffer:     make([][]float32, SCREEN_HEIGHT),
		rowLocks:        make([]sync.Mutex, SCREEN_HEIGHT), // INIT ROW LOCKS
		ambienceFactor:  0.01,
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
			u, v, w := assets.Barycentric(p, v0Screen, v1Screen, v2Screen)

			if u >= 0 && v >= 0 && w >= 0 {
				depth := u*depth0 + v*depth1 + w*depth2
				if depth >= 0 && depth <= 1 && depth < float64(r.DepthBuffer[y][x]) {
					// var color *lookdev.ColorRGBA
					color := r.calculateLighting(tri, camera.Transform.GetForward(), lights, u, v, w)

					r.safeSetPixel(x, y, *color)
					r.DepthBuffer[y][x] = float32(depth)
				}
			}
		}
	}
}

func (r *Renderer3D) safeSetPixel(x, y int, color lookdev.ColorRGBA) {
	// fmt.Printf("Pixel(%d,%d) = %v\n", x, y, color)
	if x < 0 || x >= r.GetWidth() || y < 0 || y >= r.GetHeight() {
		return
	}

	r.rowLocks[y].Lock()
	r.Framebuffer[y][x] = color
	r.rowLocks[y].Unlock()
}

func (r *Renderer3D) isInShadow(pos nomath.Vec3, light *Light) bool {
	if !light.Shadows || light.ShadowMap == nil {
		return false
	}

	// Early transform to light space
	lightPos := light.ShadowMap.ViewMatrix.MultiplyVec4(pos.ToVec4(1.0))
	lightClip := light.ShadowMap.ProjMatrix.MultiplyVec4(lightPos)

	// Early rejection if behind light
	if lightClip.W <= 0 {
		return false
	}

	// Perspective divide
	divided := lightClip.Divide(lightClip.W)
	ndc := divided.ToVec3()

	// Convert to shadow map coordinates [0,1] range
	sx := (ndc.X + 1) * 0.5
	sy := (1 - (ndc.Y+1)*0.5)

	// Early rejection if outside shadow map
	if sx < 0 || sx > 1 || sy < 0 || sy > 1 {
		return false
	}

	// Convert to texture coordinates
	x := sx * float64(light.ShadowMap.Width-1)
	y := sy * float64(light.ShadowMap.Height-1)

	// Current fragment's depth in light space [0,1] range
	fragmentDepth := (ndc.Z + 1) * 0.5
	const bias = 0.001

	// Rotated grid sampling (8 samples - good quality/performance balance)
	offsets := [8]struct{ x, y float64 }{
		{-1, -1}, {1, -1}, {-1, 1}, {1, 1},
		{-1.5, -0.5}, {0.5, -1.5}, {-0.5, 1.5}, {1.5, 0.5},
	}
	texelSize := 1.0 / float64(light.ShadowMap.Width)
	shadow := 0.0

	for _, offset := range offsets {
		sampleX := x + offset.x*texelSize
		sampleY := y + offset.y*texelSize

		// Skip samples outside the shadow map
		ix := int(sampleX)
		iy := int(sampleY)
		if ix < 0 || ix >= light.ShadowMap.Width || iy < 0 || iy >= light.ShadowMap.Height {
			continue
		}

		shadowMapDepth := light.ShadowMap.Depth[iy][ix]
		if fragmentDepth > shadowMapDepth+bias {
			shadow += 1.0

			// Early exit if definitely in shadow (4+ samples agree)
			if shadow >= 4.0 {
				return true
			}
		} else {
			// Early exit if definitely not in shadow (5+ samples agree)
			if (8.0 - shadow) >= 5.0 {
				return false
			}
		}
	}

	// Return shadow factor (0.0 = fully lit, 1.0 = fully shadowed)
	return shadow/8.0 > 0.5
}

func (r *Renderer3D) RenderShadowMap(light *Light, scene *Scene) {
	if !light.Shadows || light.ShadowMap == nil {
		return
	}

	// Clear shadow map
	for y := 0; y < light.ShadowMap.Height; y++ {
		for x := 0; x < light.ShadowMap.Width; x++ {
			light.ShadowMap.Depth[y][x] = math.MaxFloat64
		}
	}

	// Get light view-projection matrix
	lightVP := light.ShadowMap.ProjMatrix.Multiply(light.ShadowMap.ViewMatrix)

	// Render all triangles from light's perspective
	for _, triangle := range scene.Triangles {
		modelMatrix := triangle.Parent.Transform.GetMatrix()
		mvpMatrix := lightVP.Multiply(modelMatrix)

		// Transform vertices to clip space
		v0 := mvpMatrix.MultiplyVec4(triangle.V0.ToVec4(1.0))
		v1 := mvpMatrix.MultiplyVec4(triangle.V1.ToVec4(1.0))
		v2 := mvpMatrix.MultiplyVec4(triangle.V2.ToVec4(1.0))

		// Skip triangles that are completely behind the light
		if v0.W <= 0 && v1.W <= 0 && v2.W <= 0 {
			continue
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
}
func (r *Renderer3D) calculateLighting(
	tri *assets.Triangle,
	viewDir nomath.Vec3,
	lights []*Light,
	u, v, w float64,
) *lookdev.ColorRGBA {
	baseColor := tri.DiffuseBuffer
	if baseColor == nil {
		return lookdev.NewColorRGBA() // fallback gray
	}

	// Precompute world position and normal once
	local := tri.V0.Multiply(u).Add(tri.V1.Multiply(v)).Add(tri.V2.Multiply(w))
	modelMatrix := tri.Parent.Transform.GetMatrix()
	fragmentPos := modelMatrix.MultiplyVec4(local.ToVec4(1.0)).ToVec3()
	worldNormal := modelMatrix.Inverse().Transpose().TransformVec3(
		tri.InterpolatedNormal(u, v, w),
	).Normalize()
	viewDir = viewDir.Normalize()

	var accumR, accumG, accumB float64

	// Ambient term (compute once)
	ambientR := float64(baseColor.R) * r.ambienceFactor
	ambientG := float64(baseColor.G) * r.ambienceFactor
	ambientB := float64(baseColor.B) * r.ambienceFactor

	for _, light := range lights {
		// Skip light if intensity is negligible
		if light.Intensity <= 0.001 {
			continue
		}

		// Compute light direction and distance
		var lightDir nomath.Vec3
		var dist float64
		var attenuation float64 = 1.0

		if light.Type == LightTypeDirectional {
			lightDir = light.GetDirection().Normalize()
		} else {
			lightDir = light.Transform.Position.Subtract(fragmentPos)
			dist = lightDir.Length()
			lightDir = lightDir.Normalize()

			// Skip light if too far away
			if light.Type == LightTypePoint && dist > 50.0 { // Adjust distance as needed
				continue
			}
			attenuation = 1.0 / (1.0 + light.Attenuation*dist*dist)
		}

		// Flip normal if backface
		normal := worldNormal
		if normal.Dot(lightDir) < 0 {
			normal = normal.Negate()
		}

		// Shadow calculation (only if light casts shadows)
		shadowFactor := 0.0
		if light.Shadows && light.Intensity > 0.1 { // Skip shadow calc for weak lights
			if r.isInShadow(fragmentPos, light) {
				shadowFactor = 1.0
			}
		}

		// Diffuse term
		diff := max(0.0, normal.Dot(lightDir))

		// Specular term (skip if not facing light)
		var specularR, specularG, specularB float64
		if diff > 0 {
			halfVec := lightDir.Add(viewDir).Normalize()
			specAngle := max(0.0, normal.Dot(halfVec))
			specular := math.Pow(specAngle, tri.Material.Shininess)

			specularR = float64(tri.SpecularBuffer.R) * specular
			specularG = float64(tri.SpecularBuffer.G) * specular
			specularB = float64(tri.SpecularBuffer.B) * specular
		}

		// Apply all factors
		lightFactor := light.Intensity * attenuation * (1.0 - shadowFactor)
		diff *= lightFactor

		accumR += float64(baseColor.R)*diff + specularR
		accumG += float64(baseColor.G)*diff + specularG
		accumB += float64(baseColor.B)*diff + specularB
	}

	// Add ambient and clamp
	return &lookdev.ColorRGBA{
		R: uint8(min(255, accumR+ambientR)),
		G: uint8(min(255, accumG+ambientG)),
		B: uint8(min(255, accumB+ambientB)),
		A: baseColor.A,
	}
}

func (r *Renderer3D) DrawTriangle3D(p1, p2, p3 nomath.Vec3, camera *PerspectiveCamera, color *lookdev.ColorRGBA) {
	// Transform and rasterize triangle with projection
	// You can draw as 3 lines or fill it if you support that
	r.DrawLine3D(p1, p2, camera, color)
	r.DrawLine3D(p2, p3, camera, color)
	r.DrawLine3D(p3, p1, camera, color)
}

func (r *Renderer3D) DrawText3D(text string, position nomath.Vec3, camera *PerspectiveCamera, color *lookdev.ColorRGBA) {
	x, y, ok := r.ProjectToScreen(position, camera)
	if !ok {
		return
	}
	r.DrawText2D(text, x, y, &lookdev.ColorRGBA{R: color.R, G: color.G, B: color.B, A: 255})
}

func (r *Renderer3D) ProjectToScreen(pos nomath.Vec3, camera *PerspectiveCamera) (int, int, bool) {
	viewMatrix := camera.GetViewMatrix()
	projMatrix := camera.GetProjectionMatrix()

	viewPos := viewMatrix.MultiplyVec4(pos.ToVec4(1.0))
	clipPos := projMatrix.MultiplyVec4(viewPos)

	// Avoid projecting behind the camera
	if clipPos.W <= 0 {
		return 0, 0, false
	}

	ndc := clipPos.Divide(clipPos.W).ToVec3()
	x, y := r.NDCToScreen(ndc)

	// Optionally reject screen points outside the frame
	if x < 0 || x >= r.GetWidth() || y < 0 || y >= r.GetHeight() {
		return x, y, false
	}

	return x, y, true
}

func SaveShadowMapAsImage(light *Light, filename string) {
	img := image.NewGray(image.Rect(0, 0, light.ShadowMap.Width, light.ShadowMap.Height))
	for y := 0; y < light.ShadowMap.Height; y++ {
		for x := 0; x < light.ShadowMap.Width; x++ {
			depth := light.ShadowMap.Depth[y][x]
			normalized := uint8((1 - min(depth, 1.0)) * 255) // invert for visual clarity
			img.SetGray(x, y, color.Gray{Y: normalized})
		}
	}
	f, _ := os.Create(filename)
	defer f.Close()
	png.Encode(f, img)
}
