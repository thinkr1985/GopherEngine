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
var currentKeyboardImage string = "default"

func generateKeybaordTextureMap() map[string]rl.Texture2D {
	keyboardTextures := map[string]rl.Texture2D{
		"default":     rl.LoadTexture("sources/keyboard.png"),
		"W":           rl.LoadTexture("sources/W_pressed.png"),
		"A":           rl.LoadTexture("sources/A_pressed.png"),
		"S":           rl.LoadTexture("sources/S_pressed.png"),
		"D":           rl.LoadTexture("sources/D_pressed.png"),
		"E":           rl.LoadTexture("sources/E_pressed.png"),
		"Q":           rl.LoadTexture("sources/Q_pressed.png"),
		"arrowRight":  rl.LoadTexture("sources/Right_Arrow_pressed.png"),
		"arrowLeft":   rl.LoadTexture("sources/Left_Arrow_pressed.png"),
		"middleMouse": rl.LoadTexture("sources/scroll.png"),
		"leftMouse":   rl.LoadTexture("sources/left_mouse_clicked.png"),
		"rightMouse":  rl.LoadTexture("sources/right_mouse_clicked.png"),
		"scroll":      rl.LoadTexture("sources/scroll.png"),
	}

	return keyboardTextures
}

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

	keyboardTextures := generateKeybaordTextureMap()
	defer func() {
		for _, tex := range keyboardTextures {
			rl.UnloadTexture(tex)
		}
	}()

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
		// Draw keyboard overlay
		drawKeyboardOverlay(keyboardTextures[currentKeyboardImage])
		rl.EndDrawing()
	}
}

func handleKeyboardEvents(scene *core.Scene) {
	moveSpeed := 0.1
	rotateSpeed := 0.02
	currentKeyboardImage = "default" // reset every frame

	// Forward and backward
	if rl.IsKeyDown(rl.KeyW) {
		forward := scene.Camera.Transform.GetForward().Multiply(moveSpeed)
		scene.Camera.Transform.Translate(forward)
		currentKeyboardImage = "W"
	}
	if rl.IsKeyDown(rl.KeyS) {
		backward := scene.Camera.Transform.GetForward().Multiply(-moveSpeed)
		scene.Camera.Transform.Translate(backward)
		currentKeyboardImage = "S"
	}

	// Left and right
	if rl.IsKeyDown(rl.KeyA) {
		left := scene.Camera.Transform.GetRight().Multiply(-moveSpeed)
		scene.Camera.Transform.Translate(left)
		currentKeyboardImage = "A"
	}
	if rl.IsKeyDown(rl.KeyD) {
		right := scene.Camera.Transform.GetRight().Multiply(moveSpeed)
		scene.Camera.Transform.Translate(right)
		currentKeyboardImage = "D"
	}

	if rl.IsKeyDown(rl.KeyQ) {
		scene.Camera.FocalLength--
		currentKeyboardImage = "Q"
	}
	if rl.IsKeyDown(rl.KeyE) {
		scene.Camera.FocalLength++
		currentKeyboardImage = "E"
	}

	if rl.IsKeyDown(rl.KeyRight) {
		scene.Camera.Transform.Rotate(nomath.Vec3{Y: -rotateSpeed})
		currentKeyboardImage = "arrowRight"
	}
	if rl.IsKeyDown(rl.KeyLeft) {
		scene.Camera.Transform.Rotate(nomath.Vec3{Y: rotateSpeed})
		currentKeyboardImage = "arrowLeft"
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
		currentKeyboardImage = "scroll"
	}

	// --- Scroll to zoom ---
	scroll := rl.GetMouseWheelMove()
	if scroll != 0 {
		zoomSpeed := 1.0
		forward := scene.Camera.Transform.GetForward().Multiply(float64(scroll) * zoomSpeed)
		scene.Camera.Transform.Translate(forward)
		currentKeyboardImage = "scroll"
	}

	// --- Left drag to rotate around Y axis ---
	if rl.IsMouseButtonDown(rl.MouseLeftButton) {
		rotationSpeed := 0.002
		angle := -float64(delta.X) * rotationSpeed
		scene.Camera.Transform.Rotate(nomath.Vec3{Y: angle})
		currentKeyboardImage = "leftMouse"
	}
	if rl.IsMouseButtonDown(rl.MouseRightButton) {

		currentKeyboardImage = "rightMouse"
	}

	lastMousePos = mousePos
}

func draw_debug_stats() {
	statsText := core.GetMachineStats()
	textWidth := rl.MeasureText(statsText, 12)
	rl.DrawRectangle(10, 10, textWidth+80, 80, rl.NewColor(0, 0, 0, 60))
	rl.DrawTextEx(debugFont, statsText, rl.NewVector2(20, 40), 12, 2, rl.LightGray)
}

func drawKeyboardOverlay(tex rl.Texture2D) {
	x := 20
	y := rl.GetScreenHeight() - int(tex.Height) - 20
	rl.DrawTexture(tex, int32(x), int32(y), rl.White)
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
