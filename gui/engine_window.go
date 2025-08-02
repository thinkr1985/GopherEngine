package gui

import (
	"GopherEngine/core"
	"GopherEngine/nomath"
	"image"
	"image/color"
	_ "math"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var engine_icon_path = "sources/go_engine_ico.png"
var debugFont rl.Font
var lastMousePos rl.Vector2
var isFirstFrame = true

func initWindow() {

	// rl.SetConfigFlags(rl.FlagMsaa4xHint) // 4X anti-aliasing .. can reduce FPS.
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(int32(core.SCREEN_WIDTH), int32(core.SCREEN_HEIGHT), "Gopher Engine")

	debugFont = rl.LoadFontEx("fonts/CONSOLA.TTF", 12, nil, 0)

	icon := rl.LoadImage(engine_icon_path)
	rl.SetWindowIcon(*icon)
	rl.UnloadImage(icon)

	rl.SetTargetFPS(0) // uncapped

}

func loadTexture(getImage func() *image.RGBA) rl.Texture2D {
	img := getImage()
	rgbaSlice := convertToColorRGBASlice(img)
	rlImg := rl.Image{
		Data:    unsafe.Pointer(&rgbaSlice[0]),
		Width:   int32(img.Bounds().Dx()),
		Height:  int32(img.Bounds().Dy()),
		Mipmaps: 1,
		Format:  rl.UncompressedR8g8b8a8,
	}
	return rl.LoadTextureFromImage(&rlImg)
}

func handleWindowResize() {
	minWidth := 300
	minHeight := 200
	if rl.IsWindowResized() {
		newWidth := int(rl.GetScreenWidth())
		newHeight := int(rl.GetScreenHeight())

		// Enforce minimum size
		if newWidth < minWidth || newHeight < minHeight {
			rl.SetWindowSize(
				max(newWidth, minWidth),
				max(newHeight, minHeight),
			)
		}

		// Update dimensions
		core.SCREEN_WIDTH = newWidth
		core.SCREEN_HEIGHT = newHeight

	}
}

func Window(getImage func() *image.RGBA, scene *core.Scene) {
	initWindow()
	tex := loadTexture(getImage)
	defer rl.CloseWindow()
	defer rl.UnloadFont(debugFont)
	defer rl.UnloadTexture(tex)

	// window render loop
	for !rl.WindowShouldClose() {
		handleWindowResize()
		handleKeyboardEvents(scene)
		handleMouseEvents(scene)

		rl.DrawFPS(20, 20)
		draw_debug_stats()

		// Draw frame
		rl.BeginDrawing()
		rl.ClearBackground(rl.DarkGray)
		//Draw your triangles here

		rl.EndDrawing()
	}
}

func handleKeyboardEvents(scene *core.Scene) {
	moveSpeed := 0.1
	rotateSpeed := 0.02

	// Forward and backward
	if rl.IsKeyDown(rl.KeyW) {
		forward := scene.Camera.Transform.GetForward().Multiply(moveSpeed)
		scene.Camera.Transform.Translate(forward)
	}
	if rl.IsKeyDown(rl.KeyS) {
		backward := scene.Camera.Transform.GetForward().Multiply(-moveSpeed)
		scene.Camera.Transform.Translate(backward)
	}

	// Left and right
	if rl.IsKeyDown(rl.KeyA) {
		left := scene.Camera.Transform.GetRight().Multiply(-moveSpeed)
		scene.Camera.Transform.Translate(left)
	}
	if rl.IsKeyDown(rl.KeyD) {
		right := scene.Camera.Transform.GetRight().Multiply(moveSpeed)
		scene.Camera.Transform.Translate(right)
	}

	if rl.IsKeyDown(rl.KeyQ) {
		scene.Camera.FocalLength--
	}
	if rl.IsKeyDown(rl.KeyE) {
		scene.Camera.FocalLength++
	}

	if rl.IsKeyDown(rl.KeyRight) {
		scene.Camera.Transform.Rotate(nomath.Vec3{Y: -rotateSpeed})
	}
	if rl.IsKeyDown(rl.KeyLeft) {
		scene.Camera.Transform.Rotate(nomath.Vec3{Y: rotateSpeed})
	}
}

func handleMouseEvents(scene *core.Scene) {
	mousePos := rl.GetMousePosition()

	if isFirstFrame {
		lastMousePos = mousePos
		isFirstFrame = false
		return
	}

	delta := rl.Vector2Subtract(mousePos, lastMousePos)

	// --- Middle mouse pan ---
	if rl.IsMouseButtonDown(rl.MouseMiddleButton) {
		panSpeed := 0.005
		right := scene.Camera.Transform.GetRight().Multiply(float64(-delta.X) * panSpeed)
		up := scene.Camera.Transform.GetUp().Multiply(float64(delta.Y) * panSpeed)
		pan := right.Add(up)
		scene.Camera.Transform.Translate(pan)
	}

	// --- Scroll to zoom ---
	scroll := rl.GetMouseWheelMove()
	if scroll != 0 {
		zoomSpeed := 1.0
		forward := scene.Camera.Transform.GetForward().Multiply(float64(scroll) * zoomSpeed)
		scene.Camera.Transform.Translate(forward)
	}

	// --- Left drag to rotate around Y axis ---
	if rl.IsMouseButtonDown(rl.MouseLeftButton) {
		rotationSpeed := 0.002
		angle := -float64(delta.X) * rotationSpeed
		scene.Camera.Transform.Rotate(nomath.Vec3{Y: angle})
	}

	lastMousePos = mousePos
}

func draw_debug_stats() {
	statsText := core.GetMachineStats()
	textWidth := rl.MeasureText(statsText, 12)
	rl.DrawRectangle(10, 10, textWidth+80, 80, rl.NewColor(0, 0, 0, 60))
	rl.DrawTextEx(debugFont, statsText, rl.NewVector2(20, 40), 12, 2, rl.LightGray)
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
