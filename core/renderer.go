package core

import (
	"GopherEngine/assets"
	"GopherEngine/lookdev"
	"GopherEngine/nomath"
	"fmt"
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
	fmt.Printf("Resizing Renderer to new size %v %v", width, height)
	r.bufferMutex.Lock()
	defer r.bufferMutex.Unlock()

	// Create new buffers
	newFramebuffer := make([][]lookdev.ColorRGBA, height)
	newDepthBuffer := make([][]float32, height)

	for y := 0; y < height; y++ {
		newFramebuffer[y] = make([]lookdev.ColorRGBA, width)
		newDepthBuffer[y] = make([]float32, width)

		// Optional: Copy existing content if possible
		if y < len(r.Framebuffer) && width == len(r.Framebuffer[y]) {
			copy(newFramebuffer[y], r.Framebuffer[y])
			copy(newDepthBuffer[y], r.DepthBuffer[y])
		} else {
			// Initialize new rows
			for x := 0; x < width; x++ {
				newDepthBuffer[y][x] = math.MaxFloat32
			}
		}
	}

	// Atomic swap (no allocation during rendering)
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
	// Use memset-style optimization for clearing
	for y := 0; y < SCREEN_HEIGHT; y++ {
		rowPixels := r.Framebuffer[y]
		rowDepth := r.DepthBuffer[y]

		// Clear pixels
		for x := 0; x < SCREEN_WIDTH; x++ {
			rowPixels[x] = color
		}

		// Clear depth (using memcpy pattern)
		for x := 0; x < SCREEN_WIDTH; x++ {
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
				A: uint8(c.A * 255), // Convert float alpha to uint8
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
	// Pre-calculate screen dimensions as float64
	width := float64(SCREEN_WIDTH)
	height := float64(SCREEN_HEIGHT)

	// Optimized calculation
	x := int((ndc.X+1)*0.5*width + 0.5) // +0.5 for rounding
	y := int((1-(ndc.Y+1)*0.5)*height + 0.5)
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
	// Early bounds check for both points
	if (x0 < 0 && x1 < 0) || (x0 >= SCREEN_WIDTH && x1 >= SCREEN_WIDTH) ||
		(y0 < 0 && y1 < 0) || (y0 >= SCREEN_HEIGHT && y1 >= SCREEN_HEIGHT) {
		return
	}

	// Early exit if points are the same
	if x0 == x1 && y0 == y1 {
		if x0 >= 0 && x0 < SCREEN_WIDTH && y0 >= 0 && y0 < SCREEN_HEIGHT {
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
	maxIterations := dx + dy + 1 // Reduced by 1 since we check before setting

	for i := 0; i < maxIterations; i++ {
		// Check bounds before setting pixel
		if x0 >= 0 && x0 < SCREEN_WIDTH && y0 >= 0 && y0 < SCREEN_HEIGHT {
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

func (r *Renderer3D) RenderGeometry(g *assets.Geometry) {
	for _, triangle := range g.Triangles {
		fmt.Printf("rendering.. %v", triangle)
	}

}
