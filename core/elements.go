package core

import (
	"GopherEngine/lookdev"
	"GopherEngine/nomath"
)

// ViewAxes represents the 3D axis indicator displayed in screen space
type ViewAxes struct {
	Enabled    bool
	Size       float64     // Size in screen pixels
	ScreenPos  nomath.Vec2 // Position in screen space (0-1 normalized)
	Colors     [3]lookdev.ColorRGBA
	cameraAxes [6]nomath.Vec3 // Stores current camera orientation axes
}

func NewViewAxes() *ViewAxes {
	return &ViewAxes{
		Enabled:   true,
		Size:      200.0,                       // Pixel size of the widget
		ScreenPos: nomath.Vec2{U: 0.9, V: 0.9}, // Top-right corner
		Colors: [3]lookdev.ColorRGBA{
			{R: 255, G: 0, B: 0, A: 1.0}, // X
			{R: 0, G: 255, B: 0, A: 1.0}, // Y
			{R: 0, G: 0, B: 255, A: 1.0}, // Z
		},
	}
}

func (va *ViewAxes) Update(camera *PerspectiveCamera) {
	if camera == nil {
		return
	}

	// Get camera orientation vectors
	right := camera.Transform.GetRight()
	up := camera.Transform.GetUp()
	forward := camera.Transform.GetForward()

	// Scale vectors to make them visible
	scale := 0.2 * va.Size
	va.cameraAxes = [6]nomath.Vec3{
		nomath.Vec3{}, right.Multiply(scale), // X axis (right)
		nomath.Vec3{}, up.Multiply(scale), // Y axis (up)
		nomath.Vec3{}, forward.Multiply(-scale), // Z axis (forward)
	}
}
func (va *ViewAxes) Draw(renderer *Renderer3D, camera *PerspectiveCamera) {
	if !va.Enabled || renderer == nil || camera == nil {
		return
	}

	va.Update(camera)

	// Convert screen position to pixel coordinates using renderer dimensions
	screenX := int(va.ScreenPos.U * float64(renderer.GetWidth()))
	screenY := int(va.ScreenPos.V * float64(renderer.GetHeight()))

	// Draw axes in screen space
	for i := 0; i < 3; i++ {
		start := va.cameraAxes[i*2]
		end := va.cameraAxes[i*2+1]

		// Convert to screen space using renderer dimensions
		startScreen := nomath.Vec3{
			X: float64(screenX) + start.X,
			Y: float64(screenY) - start.Y,
			Z: 0,
		}
		endScreen := nomath.Vec3{
			X: float64(screenX) + end.X,
			Y: float64(screenY) - end.Y,
			Z: 0,
		}

		renderer.DrawLine2D(
			int(startScreen.X), int(startScreen.Y),
			int(endScreen.X), int(endScreen.Y),
			&va.Colors[i],
		)

		// Draw labels
		labelPos := nomath.Vec3{
			X: float64(screenX) + end.X*1.2,
			Y: float64(screenY) - end.Y*1.2,
			Z: 0,
		}
		renderer.DrawText2D(
			string([]byte{'X' + byte(i)}),
			int(labelPos.X), int(labelPos.Y),
			&va.Colors[i],
		)
	}
}

// Grid represents a fixed-size 3D grid
type Grid struct {
	Enabled     bool
	Lines       []LineSegment // Pre-computed line segments
	Color       lookdev.ColorRGBA
	CenterColor lookdev.ColorRGBA
}

type LineSegment struct {
	Start, End nomath.Vec3
	Color      *lookdev.ColorRGBA
}

func NewGrid() *Grid {
	g := &Grid{
		Enabled:     true,
		Color:       lookdev.ColorRGBA{R: 191, G: 196, B: 197, A: 1.0}, // Brighter color
		CenterColor: lookdev.ColorRGBA{R: 255, G: 0, B: 0, A: 1.0},     // Red center lines
	}
	g.BuildGrid(21, 5.0) // 21x21 grid with spacing 1.0
	return g
}

func (g *Grid) BuildGrid(size int, spacing float64) {
	halfSize := float64(size-1) * spacing / 2
	halfCount := size / 2

	// Pre-allocate exact number of lines (size*2 for X + size*2 for Z)
	g.Lines = make([]LineSegment, 0, size*2)

	// Build X-axis lines
	for i := -halfCount; i <= halfCount; i++ {
		x := float64(i) * spacing
		color := &g.Color
		if i == 0 && g.Enabled {
			color = &g.CenterColor
		}

		g.Lines = append(g.Lines, LineSegment{
			Start: nomath.Vec3{X: x, Y: 0, Z: -halfSize},
			End:   nomath.Vec3{X: x, Y: 0, Z: halfSize},
			Color: color,
		})
	}

	// Build Z-axis lines
	for i := -halfCount; i <= halfCount; i++ {
		z := float64(i) * spacing
		color := &g.Color
		if i == 0 && g.Enabled {
			color = &g.CenterColor
		}

		g.Lines = append(g.Lines, LineSegment{
			Start: nomath.Vec3{X: -halfSize, Y: 0, Z: z},
			End:   nomath.Vec3{X: halfSize, Y: 0, Z: z},
			Color: color,
		})
	}
}

func (g *Grid) Draw(renderer *Renderer3D, camera *PerspectiveCamera) {
	if !g.Enabled || renderer == nil || camera == nil {
		return
	}

	// Draw pre-computed lines
	for _, line := range g.Lines {
		// Create a small bounding box around the line
		min := nomath.Min(line.Start, line.End) // Use the standalone Min function
		max := nomath.Max(line.Start, line.End) // Use the standalone Max function
		bbox := &nomath.BoundingBox{Min: min, Max: max}

		if camera.IsVisible(bbox) {
			renderer.DrawLine3D(line.Start, line.End, camera, line.Color)
		}
	}
}
