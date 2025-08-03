package main

import (
	"GopherEngine/assets"
	"GopherEngine/core"
	"GopherEngine/gui"
	"GopherEngine/lookdev"
	"log"
)

func main() {
	scene := core.NewScene()

	// Load the OBJ model
	tree, err := assets.LoadOBJ("objs/tree_bark.obj")
	if err != nil {
		log.Fatalf("Failed to load OBJ file: %v", err)
	}
	tree.Material.DiffuseColor = lookdev.ColorRGBA{R: 34, G: 139, B: 34, A: 1.0} // Forest Green

	tex, err := lookdev.LoadTexture("textures/bark_0021.jpg")
	if err != nil {
		log.Printf("Warning: Failed to load texture: %v", err)
	} else {
		tree.Material.DiffuseTexture = tex
	}
	scene.AddObject(tree)
	gui.Window(scene)
}
