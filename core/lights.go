package core

import (
	"GopherEngine/lookdev"
	"GopherEngine/nomath"
	"fmt"
)

const (
	LightTypeDirectional = 0
	LightTypePoint       = 1
)

type Light struct {
	Name        string
	Direction   nomath.Vec3
	Transform   *nomath.Transform
	Color       *lookdev.ColorRGBA
	Intensity   float64
	Attenuation float64
	Type        int
}

func NewPointLight() *Light {
	l := &Light{
		Name:        "DefaultPointLight",
		Transform:   nomath.NewTransform(),
		Color:       lookdev.NewColorRGBA(),
		Intensity:   1.0,
		Attenuation: 1.0,
		Type:        LightTypePoint,
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
		Intensity:   1.0,
		Attenuation: 1.0,
		Type:        LightTypeDirectional,
	}
	// making light color a white.
	l.Color.R = 255
	l.Color.G = 255
	l.Color.B = 255
	l.Transform.SetPosition(nomath.Vec3{X: 0, Y: 100, Z: 20})
	l.Transform.Rotation.Z = 45.0
	l.Transform.UpdateModelMatrix()
	return l
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
}
