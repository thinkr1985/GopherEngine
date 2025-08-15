package core

import (
	"GopherEngine/lookdev"
	"GopherEngine/nomath"
	"math"
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
	forward := camera.Transform.GetForward().Negate()

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

type Grid struct {
	Enabled     bool
	Color       lookdev.ColorRGBA
	CenterColor lookdev.ColorRGBA
	Spacing     float64
	Size        int
	MaxDistance float64 // Distance beyond which grid isn't drawn
}

func NewGrid() *Grid {
	return &Grid{
		Enabled:     true,
		Color:       lookdev.ColorRGBA{R: 191, G: 196, B: 197, A: 0.5}, // Semi-transparent
		CenterColor: lookdev.ColorRGBA{R: 255, G: 0, B: 0, A: 1.0},     // Red center lines
		Spacing:     5.0,
		Size:        21, // Should be odd number to have center line
		MaxDistance: 200.0,
	}
}

func (g *Grid) Draw(renderer *Renderer3D, camera *PerspectiveCamera) {
	if !g.Enabled || renderer == nil || camera == nil {
		return
	}

	// Get camera position and direction
	camPos := camera.Transform.Position
	viewDir := camera.Transform.GetForward()

	// Skip if camera is too far or looking away
	if camPos.Y > g.MaxDistance || viewDir.Dot(nomath.Vec3{Y: -1}) < 0.3 {
		return
	}

	// Calculate grid center at camera's XZ position but fixed Y=0
	center := nomath.Vec3{X: camPos.X, Y: 0, Z: camPos.Z}

	// Determine LOD based on distance
	distance := camPos.DistanceTo(center)
	lod := 1
	if distance > 50 {
		lod = 2
	}
	if distance > 100 {
		lod = 4
	}

	// Calculate visible range based on camera frustum
	halfSize := float64(g.Size-1) * g.Spacing / 2
	startX := center.X - halfSize
	endX := center.X + halfSize
	startZ := center.Z - halfSize
	endZ := center.Z + halfSize

	// Draw X-axis lines
	for z := startZ; z <= endZ; z += g.Spacing * float64(lod) {
		// Skip if not aligned with LOD
		if math.Mod(z, g.Spacing*float64(lod)) != 0 {
			continue
		}

		color := &g.Color
		if math.Abs(z-center.Z) < g.Spacing/2 {
			color = &g.CenterColor
		}

		renderer.DrawLine3D(
			nomath.Vec3{X: startX, Y: 0, Z: z},
			nomath.Vec3{X: endX, Y: 0, Z: z},
			camera, color,
		)
	}

	// Draw Z-axis lines
	for x := startX; x <= endX; x += g.Spacing * float64(lod) {
		// Skip if not aligned with LOD
		if math.Mod(x, g.Spacing*float64(lod)) != 0 {
			continue
		}

		color := &g.Color
		if math.Abs(x-center.X) < g.Spacing/2 {
			color = &g.CenterColor
		}

		renderer.DrawLine3D(
			nomath.Vec3{X: x, Y: 0, Z: startZ},
			nomath.Vec3{X: x, Y: 0, Z: endZ},
			camera, color,
		)
	}
}
