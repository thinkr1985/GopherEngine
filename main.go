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
	tree, err := assets.LoadOBJ("objs/tree_foliage.obj")
	if err != nil {
		log.Fatalf("Failed to load OBJ file: %v", err)
	}
	tree.Transform.SetPosition(nomath.Vec3{X: 0, Y: 0, Z: -20})

	tex, err := lookdev.LoadTexture("textures/DB2X2_L01.png")
	if err != nil {
		log.Printf("Warning: Failed to load texture: %v", err)
	} else {
		tree.Material.DiffuseTexture = tex
	}
	scene.AddObject(tree)
	gui.Window(scene)
	// core.StopCPUProfile()
}
