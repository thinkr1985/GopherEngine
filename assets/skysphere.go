package assets

import (
	"GopherEngine/lookdev"
	"GopherEngine/nomath"
	"math"
)

func CreateSkySphere(texturePath string) *Geometry {
	const segments = 64
	const rings = 32

	vertices := make([]*nomath.Vec3, 0)
	triangles := make([]*Triangle, 0)

	// Generate vertices
	for r := 0; r <= rings; r++ {
		theta := float64(r) * math.Pi / float64(rings)
		sinTheta := math.Sin(theta)
		cosTheta := math.Cos(theta)

		for s := 0; s <= segments; s++ {
			phi := float64(s) * 2 * math.Pi / float64(segments)
			sinPhi := math.Sin(phi)
			cosPhi := math.Cos(phi)

			x := cosPhi * sinTheta
			y := cosTheta
			z := sinPhi * sinTheta

			vertices = append(vertices, &nomath.Vec3{X: x, Y: y, Z: z})
		}
	}

	// Generate triangles
	for r := 0; r < rings; r++ {
		for s := 0; s < segments; s++ {
			first := (r * (segments + 1)) + s
			second := first + segments + 1

			// Create two triangles per quad
			triangles = append(triangles,
				NewTriangle(nil, nil,
					vertices[first],
					vertices[second],
					vertices[first+1],
					nil, nil, nil,
					nil, nil, nil),

				NewTriangle(nil, nil,
					vertices[second],
					vertices[second+1],
					vertices[first+1],
					nil, nil, nil,
					nil, nil, nil),
			)
		}
	}

	// Create material with HDRI texture
	material := lookdev.NewMaterial("SkySphereMaterial")
	if texturePath != "" {
		tex, err := lookdev.LoadTexture(texturePath)
		if err == nil {
			material.DiffuseTexture = tex
		}
	}

	geom := &Geometry{
		Name:        "SkySphere",
		Transform:   nomath.NewTransform(),
		Vertices:    vertices,
		Triangles:   triangles,
		BoundingBox: nomath.NewBoundingBox(),
		Material:    material,
		IsVisible:   true,
	}

	geom.Transform.Scale = nomath.Vec3{X: 1000, Y: 1000, Z: 1000}
	geom.Transform.Dirty = true
	geom.ComputeBoundingBox()
	geom.Transform.Dirty = false

	return geom
}
