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
	// core.StartCPUProfile()
	scene := core.NewScene()
	scene.LoadAsset("E:/GitHub/GopherEngine/tests/tree.asset")
	// scene.LoadAssembly("E:/GitHub/GopherEngine/tests/tree.ably")
	/*
		assemby := assets.NewAssembly()
		assemby.Name = "plane"

		plane, err := assets.LoadOBJ("objs/ground_plane_small.obj")
		if err != nil {
			log.Fatalf("Failed to load OBJ file: %v", err)
		}
		plane.Name = "Plane"
		ground_tex, err := lookdev.LoadTexture("textures/ground_grid.jpg")
		if err != nil {
			log.Printf("Warning: Failed to load texture: %v", err)
		} else {
			plane.Material.DiffuseTexture = ground_tex
		}
		assemby.AddGeometry(plane)

		obj, err := assets.LoadOBJ("objs/tree_foliage.obj")
		if err != nil {
			log.Fatalf("Failed to load OBJ file: %v", err)
		}
		tex_foliage, err := lookdev.LoadTexture("textures/DB2X2_L01.png")
		if err != nil {
			log.Printf("Warning: Failed to load texture: %v", err)
		} else {
			obj.Material.DiffuseTexture = tex_foliage
		}
		spec_foliage, err := lookdev.LoadTexture("textures/DB2X2_L01_Spec.png")
		if err != nil {
			log.Printf("Warning: Failed to load texture: %v", err)
		} else {
			obj.Material.SpecularTexture = spec_foliage
		}

		obj.Name = "tree_foliage"

		obj2, err := assets.LoadOBJ("objs/tree_bark.obj")
		if err != nil {
			log.Fatalf("Failed to load OBJ file: %v", err)
		}
		tex_bark, err := lookdev.LoadTexture("textures/bark_0021.jpg")
		if err != nil {
			log.Printf("Warning: Failed to load texture: %v", err)
		} else {
			obj2.Material.DiffuseTexture = tex_bark
		}
		obj2.Name = "tree_bark"
		assemby.AddGeometry(obj)
		assemby.AddGeometry(obj2)
		scene.AddAssembly(assemby)
		// assemby.SaveAssembly("tree", "tests/")
		// assets.AssetExport(assemby, "E:/GitHub/GopherEngine/tests/tree.asset")
	*/
	scene.DefaultLight.Transform.Rotation.X += 3.0 + math.Sin(25)*1.0
	scene.DefaultLight.Transform.Dirty = true
	gui.Window(scene)
	// core.StopCPUProfile()
}
