package main

import (
	"GopherEngine/core"
	"GopherEngine/gui"
)

func main() {
	scene := core.NewScene()
	gui.Window(scene) // Remove the getImage parameter entirely
}
