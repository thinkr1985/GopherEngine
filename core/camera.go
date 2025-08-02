package core

import (
	"GopherEngine/nomath"
)

type PerspectiveCamera struct {
	Name        string
	Transform   *nomath.Transform
	FocalLength int
	NearPlane   float32
	FarPlane    float32
}

func NewPerspectiveCamera() *PerspectiveCamera {
	return &PerspectiveCamera{
		Name:        "NewPerspectiveCamera",
		Transform:   nomath.NewTransform(),
		FocalLength: 75,
		NearPlane:   0.1,
		FarPlane:    10000.0,
	}
}
