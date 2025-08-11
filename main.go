package main

import (
	"GopherEngine/core"
	"GopherEngine/gui"
	"math"

	// "GopherEngine/nomath"
	_ "log"
	// "math/rand"
)

func main() {
	core.StartCPUProfile()
	scene := core.NewScene()
	scene.LoadAsset("E:/GitHub/GopherEngine/tests/test_tree_scene.asset")
	// scene.LoadAsset("E:/GitHub/GopherEngine/tests/skySphere.asset")
	// scene.LoadAsset("E:/GitHub/GopherEngine/tests/watchTower.asset")
	// scene.LoadAsset("E:/GitHub/GopherEngine/tests/ManySpheres.asset")

	/*
		assemby := assets.NewAssembly()
		assemby.Name = "PoliceCar"

		plane, err := assets.LoadOBJ("objs/carB.obj")
		if err != nil {
			log.Fatalf("Failed to load OBJ file: %v", err)
		}
		plane.Name = "PoliceCarMesh"
		ground_tex, err := lookdev.LoadTexture("textures/CarBTexture.png")
		if err != nil {
			log.Printf("Warning: Failed to load texture: %v", err)
		} else {
			plane.Material.DiffuseTexture = ground_tex
		}
		spec_tex, err := lookdev.LoadTexture("textures/Wood_Tower_Nor.jpg")
		if err != nil {
			log.Printf("Warning: Failed to load texture: %v", err)
		} else {
			plane.Material.DiffuseTexture = spec_tex
		}

		assemby.AddGeometry(plane)

		scene.AddAssembly(assemby)
		assets.AssetExport(assemby, "E:/GitHub/GopherEngine/tests/PoliceCar.asset")
	*/
	scene.DefaultLight.Transform.Rotation.X += 3.0 + math.Sin(10)*1.0
	scene.DefaultLight.Transform.Dirty = true
	gui.Window(scene)
	core.StopCPUProfile()
}
