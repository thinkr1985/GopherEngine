package core

import (
	"GopherEngine/lookdev"
	"GopherEngine/nomath"
	"fmt"
	"math"
)

const (
	LightTypeDirectional = 0
	LightTypePoint       = 1
	LightTypeSun         = 2
)

type ShadowMap struct {
	Width      int
	Height     int
	Depth      [][]float64
	ViewMatrix nomath.Mat4
	ProjMatrix nomath.Mat4
}

type Light struct {
	Name        string
	scene       *Scene
	Direction   nomath.Vec3
	Transform   *nomath.Transform
	Color       *lookdev.ColorRGBA
	Intensity   float64
	Attenuation float64
	Type        int
	Shadows     bool
	ShadowMap   *ShadowMap
}

func NewPointLight(s *Scene) *Light {
	l := &Light{
		Name:        "DefaultPointLight",
		scene:       s,
		Transform:   nomath.NewTransform(),
		Color:       lookdev.NewColorRGBA(),
		Intensity:   1.0,
		Attenuation: 1.0,
		Type:        LightTypePoint,
		Shadows:     false,
	}
	// making light color a white.
	l.Color.R = 255
	l.Color.G = 255
	l.Color.B = 255
	l.Transform.SetPosition(nomath.Vec3{X: 0, Y: 100, Z: 0})
	l.Transform.UpdateModelMatrix()

	return l
}

func NewDirectionalLight(s *Scene) *Light {
	l := &Light{
		Name:        "DefaultPointLight",
		scene:       s,
		Transform:   nomath.NewTransform(),
		Color:       lookdev.NewColorRGBA(),
		Intensity:   10.0,
		Attenuation: 1.0,
		Type:        LightTypeDirectional,
		Shadows:     false,
	}
	// making light color a white.
	l.Color.R = 255
	l.Color.G = 255
	l.Color.B = 255
	l.Transform.SetPosition(nomath.Vec3{X: 0, Y: 50, Z: 0})
	l.Transform.UpdateModelMatrix()

	// Initialize shadow map for directional light
	l.InitShadowMap(1024, 1024)

	return l
}

func NewSunLight(s *Scene) *Light {
	l := &Light{
		Name:        "DefaultPointLight",
		scene:       s,
		Transform:   nomath.NewTransform(),
		Color:       lookdev.NewColorRGBA(),
		Intensity:   10.0,
		Attenuation: 1.0,
		Type:        LightTypeSun,
		Shadows:     false,
	}
	// making light color a white.
	l.Color.R = 255
	l.Color.G = 255
	l.Color.B = 255
	l.Transform.SetPosition(nomath.Vec3{X: 0, Y: 60, Z: -30})
	l.Transform.SetRotation(nomath.Vec3{
		X: math.Pi / 2, // 90 degrees down
		Y: 0,
		Z: 0,
	})
	l.Transform.UpdateModelMatrix()

	// Initialize shadow map for directional light
	l.InitShadowMap(1024, 1024)

	return l
}

func (l *Light) InitShadowMap(width, height int) {
	l.ShadowMap = &ShadowMap{
		Width:  width,
		Height: height,
		Depth:  make([][]float64, height),
	}
	for i := range l.ShadowMap.Depth {
		l.ShadowMap.Depth[i] = make([]float64, width)
		for j := range l.ShadowMap.Depth[i] {
			l.ShadowMap.Depth[i][j] = math.MaxFloat64
		}
	}
}

func (l *Light) GetDirection() nomath.Vec3 {
	if l.Type == LightTypeDirectional {
		return l.Transform.GetForward().Normalize().Negate()
	}
	return nomath.Vec3{X: 0, Y: -1, Z: 0}
}

func (l *Light) String() string {
	return fmt.Sprintf("Light(%s, %s)", l.Name, l.Type)
}

func (l *Light) Update() {

	l.Transform.UpdateModelMatrix()

	if l.Shadows && l.ShadowMap != nil {
		// Update light view and projection matrices
		lightPos := l.Transform.Position
		target := lightPos.Add(l.GetDirection().Multiply(-100)) // Look in light direction
		up := nomath.Vec3{X: 0, Y: 1, Z: 0}                     // World up vector

		// Create light view matrix
		l.ShadowMap.ViewMatrix = nomath.LookAtMatrix(lightPos, target, up)

		// Create orthographic projection for directional light
		// Adjust these values based on your scene size
		left, right := -50.0, 50.0
		bottom, top := -50.0, 50.0
		near, far := 0.1, 200.0

		l.ShadowMap.ProjMatrix = nomath.Ortho(left, right, bottom, top, near, far)
	}
}

func (l *Light) DrawLight() {
	if l.scene == nil || l.scene.Renderer == nil {
		return
	}

	renderer := l.scene.Renderer
	camera := l.scene.Camera

	arrowLength := 10.0
	headLength := 5.0
	headWidth := 1.0 // width of the arrowhead triangle base

	start := l.Transform.Position
	var dir nomath.Vec3
	color := lookdev.NewColorRGBA()

	// Determine color
	if l.Type == LightTypeSun {
		color.R = 255
		color.G = 165
		color.B = 0
	} else {
		color.R = 255
		color.G = 255
		color.B = 0
	}

	// Determine direction
	if l.Type == LightTypeDirectional || l.Type == LightTypeSun {
		dir = l.GetDirection().Normalize()
	} else {
		dir = nomath.Vec3{X: 0, Y: 1, Z: 0}
	}

	end := start.Add(dir.Multiply(arrowLength))

	// Draw main shaft
	renderer.DrawLine3D(start, end, camera, color)

	// Choose an "up" vector that is not parallel to dir
	var up nomath.Vec3
	if math.Abs(dir.Y) > 0.99 {
		up = nomath.Vec3{X: 0, Y: 0, Z: 1}
	} else {
		up = nomath.Vec3{X: 0, Y: 1, Z: 0}
	}
	right := dir.Cross(up).Normalize()
	up = right.Cross(dir).Normalize()

	// Draw 2 triangle arrowheads (in up and right planes)
	baseCenter := end.Subtract(dir.Multiply(headLength))

	// Triangle 1: in UP plane
	tip := end
	base1a := baseCenter.Add(up.Multiply(headWidth))
	base1b := baseCenter.Subtract(up.Multiply(headWidth))
	renderer.DrawTriangle3D(tip, base1a, base1b, camera, color)

	// Triangle 2: in RIGHT plane
	base2a := baseCenter.Add(right.Multiply(headWidth))
	base2b := baseCenter.Subtract(right.Multiply(headWidth))
	renderer.DrawTriangle3D(tip, base2a, base2b, camera, color)

	// Draw Sun representation if type is SUN
	if l.Type == LightTypeSun {
		radius := 5.0
		segments := 32

		// Circle in XZ plane
		for i := 0; i < segments; i++ {
			theta1 := float64(i) * 2 * math.Pi / float64(segments)
			theta2 := float64(i+1) * 2 * math.Pi / float64(segments)
			p1 := start.Add(nomath.Vec3{
				X: radius * math.Cos(theta1),
				Y: 0,
				Z: radius * math.Sin(theta1),
			})
			p2 := start.Add(nomath.Vec3{
				X: radius * math.Cos(theta2),
				Y: 0,
				Z: radius * math.Sin(theta2),
			})
			renderer.DrawLine3D(p1, p2, camera, color)
		}

		// Circle in YZ plane
		for i := 0; i < segments; i++ {
			theta1 := float64(i) * 2 * math.Pi / float64(segments)
			theta2 := float64(i+1) * 2 * math.Pi / float64(segments)
			p1 := start.Add(nomath.Vec3{
				X: 0,
				Y: radius * math.Cos(theta1),
				Z: radius * math.Sin(theta1),
			})
			p2 := start.Add(nomath.Vec3{
				X: 0,
				Y: radius * math.Cos(theta2),
				Z: radius * math.Sin(theta2),
			})
			renderer.DrawLine3D(p1, p2, camera, color)
		}

		// Draw label "Sun"
		labelPos := start.Add(nomath.Vec3{X: 0, Y: -5, Z: 0}) // just below light position
		renderer.DrawText3D("Sun", labelPos, camera, color)
	}
	if l.Type == LightTypePoint {
		starLength := 5.0
		origin := l.Transform.Position

		// X-axis line (red)
		color := lookdev.NewColorRGBA()
		p1 := origin.Add(nomath.Vec3{X: -starLength, Y: 0, Z: 0})
		p2 := origin.Add(nomath.Vec3{X: starLength, Y: 0, Z: 0})
		color.R = 255
		color.G = 0
		color.B = 0
		renderer.DrawLine3D(p1, p2, camera, color)

		// Y-axis line (green)
		p3 := origin.Add(nomath.Vec3{X: 0, Y: -starLength, Z: 0})
		p4 := origin.Add(nomath.Vec3{X: 0, Y: starLength, Z: 0})
		color.R = 0
		color.G = 255
		color.B = 0
		renderer.DrawLine3D(p3, p4, camera, color)

		// Z-axis line (blue)
		p5 := origin.Add(nomath.Vec3{X: 0, Y: 0, Z: -starLength})
		p6 := origin.Add(nomath.Vec3{X: 0, Y: 0, Z: starLength})
		color.R = 0
		color.G = 0
		color.B = 255
		renderer.DrawLine3D(p5, p6, camera, color)

		// Optional: Label it
		labelPos := origin.Add(nomath.Vec3{X: 0, Y: -starLength - 2, Z: 0})
		renderer.DrawText3D("Point", labelPos, camera, color)
	}

}
