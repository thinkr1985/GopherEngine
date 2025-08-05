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
	Direction   nomath.Vec3
	Transform   *nomath.Transform
	Color       *lookdev.ColorRGBA
	Intensity   float64
	Attenuation float64
	Type        int
	Shadows     bool
	ShadowMap   *ShadowMap
}

func NewPointLight() *Light {
	l := &Light{
		Name:        "DefaultPointLight",
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

func NewDirectionalLight() *Light {
	l := &Light{
		Name:        "DefaultPointLight",
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
	l.Transform.SetPosition(nomath.Vec3{X: 0, Y: 50, Z: 20})
	l.Transform.Rotation.Z = 90.0
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
		return l.Direction.Normalize()
	}
	// Optional: default to zero direction
	return nomath.Vec3{X: 0, Y: 0, Z: 0}
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

		l.ShadowMap.ProjMatrix = nomath.Mat4{
			2 / (right - left), 0, 0, 0,
			0, 2 / (top - bottom), 0, 0,
			0, 0, -2 / (far - near), 0,
			-(right + left) / (right - left), -(top + bottom) / (top - bottom), -(far + near) / (far - near), 1,
		}
	}
}
