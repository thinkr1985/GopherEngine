package main

import (
	"GopherEngine/core"
	"GopherEngine/gui"

	"image"
)

func main() {
	gui.Window(func() *image.RGBA {

		img := image.NewRGBA(image.Rect(0, 0, core.SCREEN_WIDTH, core.SCREEN_HEIGHT))
		return img
	})

}
