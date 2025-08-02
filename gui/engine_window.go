package gui

import (
	"fmt"
	"image"
	"image/color"
	_ "math"
	"unsafe"

	"GopherEngine/core"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var engine_icon_path = "sources/go_engine_ico.png"

func Window(getImage func() *image.RGBA) {
	minWidth := 300
	minHeight := 200
	rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagMsaa4xHint)
	rl.InitWindow(int32(core.SCREEN_WIDTH), int32(core.SCREEN_HEIGHT), "Gopher Engine")
	defer rl.CloseWindow()
	icon := rl.LoadImage(engine_icon_path)
	rl.SetWindowIcon(*icon)
	rl.UnloadImage(icon)

	img := getImage()
	rgbaSlice := convertToColorRGBASlice(img)
	rlImg := rl.Image{
		Data:    unsafe.Pointer(&rgbaSlice[0]),
		Width:   int32(img.Bounds().Dx()),
		Height:  int32(img.Bounds().Dy()),
		Mipmaps: 1,
		Format:  rl.UncompressedR8g8b8a8,
	}
	tex := rl.LoadTextureFromImage(&rlImg)
	defer rl.UnloadTexture(tex)

	for !rl.WindowShouldClose() {

		// Handle window resize
		if rl.IsWindowResized() {
			newWidth := int(rl.GetScreenWidth())
			newHeight := int(rl.GetScreenHeight())

			// Enforce minimum size
			if newWidth < minWidth || newHeight < minHeight {
				rl.SetWindowSize(
					max(newWidth, minWidth),
					max(newHeight, minHeight),
				)
				continue
			}

			// Update dimensions
			core.SCREEN_WIDTH = newWidth
			core.SCREEN_HEIGHT = newHeight

		}

		// Draw frame
		rl.BeginDrawing()
		rl.ClearBackground(rl.DarkGray)

		fps := rl.GetFPS()
		rl.SetWindowTitle(fmt.Sprintf("Gopher Engine - FPS: %d", fps))

		rl.EndDrawing()
	}
}

func convertToColorRGBASlice(img *image.RGBA) []color.RGBA {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	pixels := make([]color.RGBA, w*h)

	i := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			pixels[i] = color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			}
			i++
		}
	}
	return pixels
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
