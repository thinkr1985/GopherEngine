package main

import (
	"GopherEngine/assets"
	"GopherEngine/core"
	"GopherEngine/gui"
	"GopherEngine/nomath"
	"log"
)

func main() {
	// core.StartCPUProfile()
	scene := core.NewScene()

	// Load the OBJ model
	tree, err := assets.LoadOBJ("objs/spheres.obj")
	if err != nil {
		log.Fatalf("Failed to load OBJ file: %v", err)
	}
	tree.Transform.SetPosition(nomath.Vec3{X: 0, Y: 0, Z: -20})

	// tex, err := lookdev.LoadTexture("textures/DB2X2_L01.png")
	// if err != nil {
	// 	log.Printf("Warning: Failed to load texture: %v", err)
	// } else {
	// 	tree.Material.DiffuseTexture = tex
	// }
	light := core.NewDirectionalLight()
	light.Transform.SetPosition(nomath.Vec3{X: 0, Y: 500, Z: 20})
	light.Transform.UpdateModelMatrix()
	light.Shadows = true
	scene.Lights = append(scene.Lights, light)
	scene.AddObject(tree)
	gui.Window(scene)
	// core.StopCPUProfile()
}
