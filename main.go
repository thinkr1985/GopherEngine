package main

import (
	"GopherEngine/core"
	"GopherEngine/gui"
	"GopherEngine/nomath"
	"math"
	v2 "math/rand/v2"

	// "GopherEngine/nomath"
	_ "log"
	// "math/rand"
)

func main() {
	core.StartCPUProfile()
	scene := core.NewScene()
	// scene.LoadAsset("E:/GitHub/GopherEngine/tests/test_tree_scene.asset")
	// scene.LoadAsset("E:/GitHub/GopherEngine/tests/skySphere.asset")
	// scene.LoadAsset("E:/GitHub/GopherEngine/tests/NestedTest.asset")

	// scene.LoadAsset("E:/GitHub/GopherEngine/tests/CartoonyTreeA.asset")
	// scene.LoadAsset("E:/GitHub/GopherEngine/tests/CartoonyTreeB.asset")
	// scene.LoadAsset("E:/GitHub/GopherEngine/tests/Stones.asset")

	i := 0
	for i < 25 {
		i += 1
		asmb := scene.LoadAsset("E:/GitHub/GopherEngine/tests/CartoonyTreeA.asset")
		asmb.Transform.SetPosition(nomath.Vec3{X: float64(v2.IntN(200)), Y: 0, Z: float64(v2.IntN(200))})
		asmb.Transform.SetRotation(nomath.Vec3{X: 0, Y: float64(v2.IntN(1 + 3)), Z: 0})
	}

	// scene.LoadAsset("E:/GitHub/GopherEngine/tests/Stones.asset")
	/*
		assemby := assets.NewAssembly()
		assemby.Name = "Stones"

		tree_bark, err := assets.LoadOBJ("objs/stone_a.obj")
		if err != nil {
			log.Fatalf("Failed to load OBJ file: %v", err)
		}
		tree_bark.Material.Shininess = 1.0
		tree_bark.Material.DiffuseColor = lookdev.ColorRGBA{R: 112, G: 112, B: 112, A: 1}
		tree_bark.Material.Reflectivity = 0.0
		tree_bark.Name = "stone_a"

		tree_canopy01, err := assets.LoadOBJ("objs/stone_b.obj")
		if err != nil {
			log.Fatalf("Failed to load OBJ file: %v", err)
		}
		tree_canopy01.Material.Shininess = 1.0
		tree_canopy01.Material.DiffuseColor = lookdev.ColorRGBA{R: 120, G: 120, B: 120, A: 1}
		tree_canopy01.Material.Reflectivity = 0.0
		tree_canopy01.Name = "stone_b"

		tree_canopy02, err := assets.LoadOBJ("objs/stone_c.obj")
		if err != nil {
			log.Fatalf("Failed to load OBJ file: %v", err)
		}
		tree_canopy02.Material.Shininess = 1.0
		tree_canopy02.Material.DiffuseColor = lookdev.ColorRGBA{R: 80, G: 80, B: 80, A: 1}
		tree_canopy02.Material.Reflectivity = 0.0
		tree_canopy02.Name = "stone_c"

		tree_canopy03, err := assets.LoadOBJ("objs/stone_d.obj")
		if err != nil {
			log.Fatalf("Failed to load OBJ file: %v", err)
		}
		tree_canopy03.Material.Shininess = 1.0
		tree_canopy03.Material.DiffuseColor = lookdev.ColorRGBA{R: 112, G: 112, B: 112, A: 1}
		tree_canopy03.Material.Reflectivity = 0.0
		tree_canopy03.Name = "stone_c"

		// ground_tex, err := lookdev.LoadTexture("textures/CarBTexture.png")
		// if err != nil {
		// 	log.Printf("Warning: Failed to load texture: %v", err)
		// } else {
		// 	plane.Material.DiffuseTexture = ground_tex
		// }

		assemby.AddGeometry(tree_bark)
		assemby.AddGeometry(tree_canopy01)
		assemby.AddGeometry(tree_canopy02)
		assemby.AddGeometry(tree_canopy03)

		scene.AddAssembly(assemby)
		assets.AssetExport(assemby, "E:/GitHub/GopherEngine/tests/Stones.asset")
	*/

	scene.DefaultLight.Transform.Rotation.X += 3.0 + math.Sin(10)*1.0
	scene.DefaultLight.Transform.Dirty = true
	gui.Window(scene)
	core.StopCPUProfile()
}
