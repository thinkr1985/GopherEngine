package assets

import (
	"GopherEngine/lookdev"
	"GopherEngine/nomath"
	"log"
)

type Skybox struct {
	Cube      *Geometry
	Textures  [6]*lookdev.Texture // One for each face
	IsVisible bool
}

func NewSkybox() (*Skybox, error) {
	s := &Skybox{
		IsVisible: true,
	}
	skyboxTextures := [6]string{
		"textures/skybox/3766.jpg",
		"textures/skybox/3766.jpg",
		"textures/skybox/3766.jpg",
		"textures/skybox/3766.jpg",
		"textures/skybox/3766.jpg",
		"textures/skybox/3766.jpg",
	}
	// Create cube geometry (inverted normals since we're inside)
	s.Cube = createSkyboxCube(skyboxTextures)

	// Load textures for each face
	/*
		var err error
		for i := 0; i < 6; i++ {
			if skyboxTextures[i] != "" {
				s.Textures[i], err = lookdev.LoadTexture(skyboxTextures[i])
				if err != nil {
					return nil, err
				}
			}
		}
	*/
	return s, nil
}

func createSkyboxCube(textures [6]string) *Geometry {
	// Create a cube with inverted normals (since we're inside it)
	// Vertices for a cube centered at origin, size 2 (from -1 to 1)
	vertices := []*nomath.Vec3{
		// Front face
		{X: -1, Y: -1, Z: -1}, // 0
		{X: 1, Y: -1, Z: -1},  // 1
		{X: 1, Y: 1, Z: -1},   // 2
		{X: -1, Y: 1, Z: -1},  // 3

		// Back face
		{X: -1, Y: -1, Z: 1}, // 4
		{X: 1, Y: -1, Z: 1},  // 5
		{X: 1, Y: 1, Z: 1},   // 6
		{X: -1, Y: 1, Z: 1},  // 7
	}

	// Triangles (2 per face, CW winding since we're inside the cube)
	triangles := []*Triangle{
		// Front (inverted)
		NewTriangle(nil, nil, vertices[2], vertices[1], vertices[0], nil, nil, nil, nil, nil, nil),
		NewTriangle(nil, nil, vertices[3], vertices[2], vertices[0], nil, nil, nil, nil, nil, nil),

		// Back (inverted)
		NewTriangle(nil, nil, vertices[6], vertices[5], vertices[4], nil, nil, nil, nil, nil, nil),
		NewTriangle(nil, nil, vertices[7], vertices[6], vertices[4], nil, nil, nil, nil, nil, nil),

		// Left (inverted)
		NewTriangle(nil, nil, vertices[3], vertices[0], vertices[4], nil, nil, nil, nil, nil, nil),
		NewTriangle(nil, nil, vertices[7], vertices[3], vertices[4], nil, nil, nil, nil, nil, nil),

		// Right (inverted)
		NewTriangle(nil, nil, vertices[1], vertices[2], vertices[6], nil, nil, nil, nil, nil, nil),
		NewTriangle(nil, nil, vertices[1], vertices[6], vertices[5], nil, nil, nil, nil, nil, nil),

		// Top (inverted)
		NewTriangle(nil, nil, vertices[2], vertices[3], vertices[7], nil, nil, nil, nil, nil, nil),
		NewTriangle(nil, nil, vertices[2], vertices[7], vertices[6], nil, nil, nil, nil, nil, nil),

		// Bottom (inverted)
		NewTriangle(nil, nil, vertices[0], vertices[1], vertices[5], nil, nil, nil, nil, nil, nil),
		NewTriangle(nil, nil, vertices[0], vertices[5], vertices[4], nil, nil, nil, nil, nil, nil),
	}

	// Assign UVs based on face
	for i, tri := range triangles {
		faceIdx := i / 2 // 0-5 for each face
		uv0 := &nomath.Vec2{U: 0, V: 0}
		uv1 := &nomath.Vec2{U: 1, V: 0}
		uv2 := &nomath.Vec2{U: 1, V: 1}
		if i%2 == 1 { // Second triangle of the face
			uv0 = &nomath.Vec2{U: 0, V: 0}
			uv1 = &nomath.Vec2{U: 1, V: 1}
			uv2 = &nomath.Vec2{U: 0, V: 1}
		}

		tri.UV0 = uv0
		tri.UV1 = uv1
		tri.UV2 = uv2
		tri.Material = lookdev.NewMaterial("SkyBoxMaterial")
		tri.Material.DiffuseColor = lookdev.ColorRGBA{R: 255, G: 255, B: 255, A: 1.0}

		if faceIdx < len(textures) && textures[faceIdx] != "" {
			skyTex, err := lookdev.LoadTexture(textures[faceIdx])
			if err != nil {
				log.Printf("Warning: Failed to load SkyBox texture: %v", err)
			} else {
				tri.Material.DiffuseTexture = skyTex
				tri.HasTexture = true
			}
		}
	}

	geom := &Geometry{
		Name:        "DefaultSkyBox",
		Transform:   nomath.NewTransform(),
		Vertices:    vertices,
		Triangles:   triangles,
		BoundingBox: nomath.NewBoundingBox(),
		IsVisible:   true,
	}
	geom.Transform.Scale = nomath.Vec3{X: 1000, Y: 1000, Z: 1000} // Large scale to encompass scene
	geom.Transform.Dirty = true
	geom.ComputeBoundingBox()

	return geom
}
