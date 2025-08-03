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
	Framebuffer     [][]lookdev.ColorRGBA // Changed to value type
	DepthBuffer     [][]float32           // Changed to float32 for better cache usage
	BackFaceCulling bool
	bufferMutex     sync.Mutex // For thread-safe resizing
}

func NewRenderer3D() *Renderer3D {
	r := &Renderer3D{
		BackFaceCulling: true,
		Framebuffer:     make([][]lookdev.ColorRGBA, SCREEN_HEIGHT),
		DepthBuffer:     make([][]float32, SCREEN_HEIGHT),
	}

	// Initialize each row
	for y := 0; y < SCREEN_HEIGHT; y++ {
		r.Framebuffer[y] = make([]lookdev.ColorRGBA, SCREEN_WIDTH)
		r.DepthBuffer[y] = make([]float32, SCREEN_WIDTH)

		// Initialize depth buffer with maximum depth
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

	// Create new buffers
	newFramebuffer := make([][]lookdev.ColorRGBA, height)
	newDepthBuffer := make([][]float32, height)

	for y := 0; y < height; y++ {
		newFramebuffer[y] = make([]lookdev.ColorRGBA, width)
		newDepthBuffer[y] = make([]float32, width)

		// Initialize depth buffer
		for x := 0; x < width; x++ {
			newDepthBuffer[y][x] = math.MaxFloat32
		}
	}

	r.Framebuffer = newFramebuffer
	r.DepthBuffer = newDepthBuffer
}
func (r *Renderer3D) PartialResize(newWidth, newHeight int) {
	r.bufferMutex.Lock()
	defer r.bufferMutex.Unlock()

	// Grow existing slices instead of reallocating
	for y := 0; y < newHeight; y++ {
		if y >= len(r.Framebuffer) {
			r.Framebuffer = append(r.Framebuffer, make([]lookdev.ColorRGBA, newWidth))
			r.DepthBuffer = append(r.DepthBuffer, make([]float32, newWidth))
		} else if newWidth > len(r.Framebuffer[y]) {
			newRow := make([]lookdev.ColorRGBA, newWidth)
			copy(newRow, r.Framebuffer[y])
			r.Framebuffer[y] = newRow

			newDepth := make([]float32, newWidth)
			copy(newDepth, r.DepthBuffer[y])
			r.DepthBuffer[y] = newDepth
		}
	}
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

func (r *Renderer3D) ClearParellel(color lookdev.ColorRGBA) {
	r.bufferMutex.Lock()
	defer r.bufferMutex.Unlock()

	height := len(r.Framebuffer)
	if height == 0 {
		return
	}
	width := len(r.Framebuffer[0])

	// Parallel clear with row caching
	var wg sync.WaitGroup
	wg.Add(height)

	for y := 0; y < height; y++ {
		go func(y int) {
			defer wg.Done()
			rowPixels := r.Framebuffer[y]
			rowDepth := r.DepthBuffer[y]

			// Batch processing (16 pixels at a time)
			for x := 0; x < width; x += 16 {
				end := min(x+16, width)
				for i := x; i < end; i++ {
					rowPixels[i] = color
					rowDepth[i] = math.MaxFloat32
				}
			}
		}(y)
	}
	wg.Wait()
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
			r.Framebuffer[y0][x0] = *color
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

func (r *Renderer3D) RenderTriangle(camera *PerspectiveCamera, tri *assets.Triangle, lights []*Light, scene *Scene) {
	// --- Transform vertices to clip space ---
	viewMatrix := camera.GetViewMatrix()
	projectionMatrix := camera.GetProjectionMatrix()
	modelMatrix := tri.Parent.Transform.GetMatrix()
	mvpMatrix := projectionMatrix.Multiply(viewMatrix).Multiply(modelMatrix)

	v0 := mvpMatrix.MultiplyVec4(tri.V0.ToVec4(1.0))
	v1 := mvpMatrix.MultiplyVec4(tri.V1.ToVec4(1.0))
	v2 := mvpMatrix.MultiplyVec4(tri.V2.ToVec4(1.0))

	// --- Perspective division (now in NDC [-1, 1]) ---
	v0 = v0.Multiply(1.0 / v0.W)
	v1 = v1.Multiply(1.0 / v1.W)
	v2 = v2.Multiply(1.0 / v2.W)

	// --- Backface culling ---
	if r.BackFaceCulling {
		edge1 := tri.V1.Subtract(*tri.V0)
		edge2 := tri.V2.Subtract(*tri.V0)
		normal := edge1.Cross(edge2).Normalize()
		viewDir := camera.Transform.GetForward()
		if normal.Dot(viewDir) > 0 {
			return // Skip back faces
		}
	}
	scene.DrawnTriangles += 1
	// --- Convert to screen coordinates ---
	x0, y0 := r.NDCToScreen(v0.ToVec3())
	x1, y1 := r.NDCToScreen(v1.ToVec3())
	x2, y2 := r.NDCToScreen(v2.ToVec3())

	// Debug: Print screen coordinates
	// fmt.Printf("Triangle screen coords: (%d,%d), (%d,%d), (%d,%d)\n", x0, y0, x1, y1, x2, y2)

	// --- Bounding box setup ---
	minX := max(0, min(x0, min(x1, x2)))
	maxX := min(r.GetWidth()-1, max(x0, max(x1, x2)))
	minY := max(0, min(y0, min(y1, y2)))
	maxY := min(r.GetHeight()-1, max(y0, max(y1, y2)))

	// Debug: Print bounding box
	// fmt.Printf("Rasterizing bounding box: [%d,%d] to [%d,%d]\n", minX, minY, maxX, maxY)

	// Precompute screen-space vertices for barycentric coords
	v0Screen := nomath.Vec2{U: float64(x0), V: float64(y0)}
	v1Screen := nomath.Vec2{U: float64(x1), V: float64(y1)}
	v2Screen := nomath.Vec2{U: float64(x2), V: float64(y2)}

	// --- Rasterize the triangle ---
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			// Convert screen coords to NDC for barycentric
			ndcX := (2.0 * float64(x) / float64(r.GetWidth())) - 1.0
			ndcY := 1.0 - (2.0 * float64(y) / float64(r.GetHeight()))
			p := nomath.Vec2{U: ndcX, V: ndcY}

			// Calculate barycentric coordinates
			u, v, w := tri.Barycentric(p, v0Screen, v1Screen, v2Screen)

			// Debug: Log first few pixels
			if x < minX+2 && y < minY+2 {
				// fmt.Printf("Pixel (%d,%d): u=%.2f v=%.2f w=%.2f\n", x, y, u, v, w)
			}

			// Check if pixel is inside the triangle
			if u >= 0 && v >= 0 && w >= 0 {
				// Calculate depth (remap from NDC [-1,1] to [0,1])
				depth := (u*v0.Z+v*v1.Z+w*v2.Z)*0.5 + 0.5

				// Depth test
				if depth >= 0 && depth <= 1 && depth < float64(r.DepthBuffer[y][x]) {
					r.Framebuffer[y][x] = *tri.DiffuseBuffer
					r.DepthBuffer[y][x] = float32(depth)
				}
			}
			r.Framebuffer[y][x] = *tri.DiffuseBuffer
			r.DepthBuffer[y][x] = float32(1)
		}
	}
}
