package main

import (
	"image"
	"log"
	"software_renderer/core"
	"software_renderer/lookdev"
	"software_renderer/nomath"

	_ "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	// Initialize renderer and scene
	scene := core.NewScene3D()

	// Load Tree Bark
	tree_a, err := core.LoadOBJ("objs/tree_bark.obj")
	if err != nil {
		log.Fatalf("Failed to load OBJ file: %v", err)
	}
	tree_a.Transform.Position = nomath.Vec3{X: 0, Y: 0, Z: -500}
	tree_a.Material.DiffuseMap = "textures/bark_0021.jpg" // Set texture path

	// Load the texture if specified
	if tree_a.Material.DiffuseMap != "" {
		tex, err := lookdev.LoadTexture(tree_a.Material.DiffuseMap)
		if err != nil {
			log.Printf("Warning: Failed to load texture %s: %v", tree_a.Material.DiffuseMap, err)
		} else {
			tree_a.Material.DiffuseTexture = tex // Changed to match the exported field
		}
	}

	// Load Tree Foliage
	tree_foliage, err := core.LoadOBJ("objs/tree_foliage.obj")
	if err != nil {
		log.Fatalf("Failed to load OBJ file: %v", err)
	}
	tree_foliage.Transform.Position = nomath.Vec3{X: 0, Y: 0, Z: -500}
	tree_foliage.Material.DiffuseMap = "textures/DB2X2_L01.png"       // Set texture path
	tree_foliage.Material.SpecularMap = "textures/DB2X2_L01_Spec.png" // Set texture path
	// Load the texture if specified
	if tree_foliage.Material.DiffuseMap != "" {
		tex, err := lookdev.LoadTexture(tree_foliage.Material.DiffuseMap)
		if err != nil {
			log.Printf("Warning: Failed to load texture %s: %v", tree_foliage.Material.DiffuseMap, err)
		} else {
			tree_foliage.Material.DiffuseTexture = tex // Changed to match the exported field
		}
	}
	if tree_a.Material.SpecularMap != "" {
		specTex, err := lookdev.LoadTexture(tree_a.Material.SpecularMap)
		if err != nil {
			log.Printf("Warning: Failed to load specular texture %s: %v", tree_a.Material.SpecularMap, err)
		} else {
			tree_a.Material.SpecularTexture = specTex
		}
	}

	tree_a.CalculateNormals()
	tree_a.PrecomputeAllBuffers()
	tree_a.UpdateModelMatrix()
	scene.AddObject(tree_a)

	tree_foliage.CalculateNormals()
	tree_foliage.PrecomputeAllBuffers()
	tree_foliage.UpdateModelMatrix()
	scene.AddObject(tree_foliage)

	// Main render loop
	// core.StartCPUProfile()

	// grid := elements.NewGrid3D()
	// view_axis := elements.NewViewAxis()
	light := core.NewDirectionalLight()
	scene.Lights = append(scene.Lights, light)

	core.Window(func() *image.RGBA {
		// scene.Renderer.Clear(&lookdev.ColorRGBA{R: 0, G: 102, B: 153, A: 1.0}, true)
		for _, value := range scene.Lights {
			value.UpdateModelMatrix()
		}
		// Reset and render
		// scene.ResetDrawnTriangles()
		scene.Render()
		// grid.Draw(scene.Renderer, scene.Camera)
		// view_axis.Draw(scene.Renderer, scene.Camera)
		return scene.Renderer.ToImage()
	}, scene)
	// core.StopCPUProfile()
}
