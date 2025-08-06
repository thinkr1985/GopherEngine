package main

import (
	"GopherEngine/assets"
	"GopherEngine/core"
	"GopherEngine/gui"
	"GopherEngine/nomath"
	"log"
	"math"
)

func main() {
	// core.StartCPUProfile()
	scene := core.NewScene()

	// Load the OBJ model
	tree, err := assets.LoadOBJ("objs/watch_tower.obj")
	if err != nil {
		log.Fatalf("Failed to load OBJ file: %v", err)
	}
	tree.Transform.SetPosition(nomath.Vec3{X: 0, Y: 0, Z: -20})

	// tex, err := lookdev.LoadTexture("textures/Wood_Tower_Col.jpg")
	// if err != nil {
	// 	log.Printf("Warning: Failed to load texture: %v", err)
	// } else {
	// 	tree.Material.DiffuseTexture = tex
	// }
	light := core.NewDirectionalLight(scene)
	light.Intensity = 1.0
	light.Transform.SetPosition(nomath.Vec3{X: 0, Y: 20, Z: 10})
	light.Transform.SetRotation(nomath.Vec3{
		X: math.Pi / 1, // 120 degrees down
		Y: 0,
		Z: 0,
	})
	light.Transform.UpdateModelMatrix()
	light.Shadows = true
	scene.Lights = append(scene.Lights, light)
	scene.AddObject(tree)
	gui.Window(scene)
	// core.StopCPUProfile()
}
