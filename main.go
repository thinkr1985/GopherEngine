package main

import (
	"GopherEngine/assets"
	"GopherEngine/core"
	"GopherEngine/gui"
	"GopherEngine/lookdev"
	"GopherEngine/nomath"
	"log"
	"math/rand"
)

func main() {
	// core.StartCPUProfile()
	scene := core.NewScene()

	spawn := 10
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
	/*
		// Load the OBJ model
		tree_foliage, err := assets.LoadOBJ("objs/tree_foliage.obj")
		if err != nil {
			log.Fatalf("Failed to load OBJ file: %v", err)
		}
		// tree_foliage.Transform.SetScale(nomath.Vec3{X: 5, Y: 5, Z: 5})
		tree_foliage.Transform.UpdateModelMatrix()
		tex_foliage, err := lookdev.LoadTexture("textures/DB2X2_L01.png")
		if err != nil {
			log.Printf("Warning: Failed to load texture: %v", err)
		} else {
			tree_foliage.Material.DiffuseTexture = tex_foliage
		}
		scene.AddObject(tree_foliage)

		tree_bark, err := assets.LoadOBJ("objs/tree_bark.obj")
		if err != nil {
			log.Fatalf("Failed to load OBJ file: %v", err)
		}

		tex_bark, err := lookdev.LoadTexture("textures/bark_0021.jpg")
		if err != nil {
			log.Printf("Warning: Failed to load texture: %v", err)
		} else {
			tree_bark.Material.DiffuseTexture = tex_bark
		}
		scene.AddObject(tree_bark)
	*/

	gui.Window(scene)
	// core.StopCPUProfile()
}
