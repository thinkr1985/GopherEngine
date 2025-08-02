package core

import (
	"GopherEngine/lookdev"
	"GopherEngine/nomath"
	"fmt"
)

type Light struct {
	Name        string
	Transform   *nomath.Transform
	Color       *lookdev.ColorRGBA
	Intensity   float64
	Attenuation float64
}

func NewPointLight() *Light {
	l := &Light{
		Name:        "DefaultPointLight",
		Transform:   nomath.NewTransform(),
		Color:       lookdev.NewColorRGBA(),
		Intensity:   1.0,
		Attenuation: 1.0,
	}
	// making light color a white.
	l.Color.R = 255
	l.Color.G = 255
	l.Color.B = 255

	return l
}

func (s *Light) String() string {
	return fmt.Sprintf("Light(%s)", s.Name)
}
