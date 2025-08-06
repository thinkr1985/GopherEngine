package main

import (
	"GopherEngine/assets"
	"GopherEngine/core"
	"GopherEngine/gui"
	"GopherEngine/lookdev"
	"GopherEngine/nomath"
	"log"
)

func main() {
	// core.StartCPUProfile()
	scene := core.NewScene()

	// Load the OBJ model
	tree, err := assets.LoadOBJ("objs/tree_bark.obj")
	if err != nil {
		log.Fatalf("Failed to load OBJ file: %v", err)
	}
	tree.Transform.SetPosition(nomath.Vec3{X: 0, Y: 0, Z: -20})

	foliage, err := assets.LoadOBJ("objs/tree_foliage.obj")
	if err != nil {
		log.Fatalf("Failed to load OBJ file: %v", err)
	}
	foliage.Transform.SetPosition(nomath.Vec3{X: 0, Y: 0, Z: -20})

	tex, err := lookdev.LoadTexture("textures/DB2X2_L01.png")
	if err != nil {
		log.Printf("Warning: Failed to load texture: %v", err)
	} else {
		foliage.Material.DiffuseTexture = tex
	}

	light := core.NewDirectionalLight(scene)
	light.Name = "Shadow Light"
	light.Intensity = 1.0
	light.Transform.SetPosition(nomath.Vec3{X: 0, Y: 20, Z: 10})
	light.Transform.LookAt(nomath.Vec3{X: 0, Y: 0, Z: -20}, nomath.Vec3{X: 0, Y: 1, Z: 0})
	light.Shadows = true
	light.InitShadowMap(1024, 1024)
	scene.Lights = append(scene.Lights, light)

	scene.AddObject(tree)
	scene.AddObject(foliage)
	gui.Window(scene)
	// core.StopCPUProfile()
}
