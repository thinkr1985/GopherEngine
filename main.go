package main

import (
	"GopherEngine/assets"
	"GopherEngine/core"
	"GopherEngine/gui"
	"GopherEngine/lookdev"

	// "GopherEngine/nomath"
	"log"
	// "math/rand"
)

func main() {
	// core.StartCPUProfile()
	scene := core.NewScene()
	assemby := assets.NewAssembly()
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

	assets.AssetExport(assemby, "E:/GitHub/GopherEngine/tests/tree.asset")
	// assemby.SaveAssembly("StandardTree", "E:/GitHub/GopherEngine/tests")
	// assembly := assets.NewAssembly()
	// assembly.LoadAssembly("E:/GitHub/GopherEngine/tests")

	/*
		spawn := 3
		for i := 0; i < spawn; i++ {
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
			obj.Transform.SetPosition(nomath.Vec3{X: rand.Float64()*30 - 10, Y: 1, Z: rand.Float64()*20 - 5})
			obj.Transform.UpdateModelMatrix()

			scene.AddObject(obj)

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
			obj2.Transform.SetPosition(nomath.Vec3{X: rand.Float64()*30 - 10, Y: 1, Z: rand.Float64()*20 - 5})
			obj2.Transform.UpdateModelMatrix()

			scene.AddObject(obj2)
		}

		ground_plane, err := assets.LoadOBJ("objs/ground_plane.obj")
		if err != nil {
			log.Fatalf("Failed to load OBJ file: %v", err)
		}
		scene.AddObject(ground_plane)
	*/

	gui.Window(scene)
	// core.StopCPUProfile()
}
