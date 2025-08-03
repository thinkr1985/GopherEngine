package gui

import (
	"GopherEngine/core"
	"GopherEngine/nomath"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var lastMousePos rl.Vector2
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
		"upArrow":     rl.LoadTexture("sources/Up_Arrow_pressed.png"),
		"downArrow":   rl.LoadTexture("sources/Down_Arrow_pressed.png"),
		"middleMouse": rl.LoadTexture("sources/scroll.png"),
		"leftMouse":   rl.LoadTexture("sources/left_mouse_clicked.png"),
		"rightMouse":  rl.LoadTexture("sources/right_mouse_clicked.png"),
		"scroll":      rl.LoadTexture("sources/scroll.png"),
		"space":       rl.LoadTexture("sources/Space_pressed.png"),
	}

	return keyboardTextures
}

func HandleInputEvents(scene *core.Scene) {
	currentKeyboardImage = "default"

	if !rl.IsWindowFocused() {
		return
	}

	if rl.IsKeyPressed(rl.KeyF1) {
		scene.HalfResolution = !scene.HalfResolution
		// Force immediate resize
		handleWindowResize(scene)
	}

	if rl.IsWindowReady() {
		HandleKeyboardEvents(scene)
		HandleMouseEvents(scene)
	}
}

func HandleKeyboardEvents(scene *core.Scene) {
	moveSpeed := 1.5 // Never use value below 1.0
	rotateSpeed := 0.02

	// Get camera vectors (note these are in world space)
	forward := scene.Camera.Transform.GetForward() // Points toward camera's view direction
	right := scene.Camera.Transform.GetRight()     // Points to camera's right

	// Movement controls - FINAL CORRECT VERSION
	if rl.IsKeyDown(rl.KeyW) {
		// Move in camera's forward direction (use positive forward vector)
		scene.Camera.Transform.Translate(forward.Multiply(moveSpeed))
		currentKeyboardImage = "W"
	}
	if rl.IsKeyDown(rl.KeyS) {
		// Move in camera's backward direction
		scene.Camera.Transform.Translate(forward.Multiply(-moveSpeed))
		currentKeyboardImage = "S"
	}
	if rl.IsKeyDown(rl.KeyA) {
		// Move left (negative right vector)
		scene.Camera.Transform.Translate(right.Multiply(-moveSpeed))
		currentKeyboardImage = "A"
	}
	if rl.IsKeyDown(rl.KeyD) {
		// Move right (positive right vector)
		scene.Camera.Transform.Translate(right.Multiply(moveSpeed))
		currentKeyboardImage = "D"
	}

	// Rotation controls (unchanged)
	if rl.IsKeyDown(rl.KeyRight) {
		scene.Camera.Transform.Rotate(nomath.Vec3{Y: -rotateSpeed})
		currentKeyboardImage = "arrowRight"
	}
	if rl.IsKeyDown(rl.KeyLeft) {
		scene.Camera.Transform.Rotate(nomath.Vec3{Y: rotateSpeed})
		currentKeyboardImage = "arrowLeft"
	}
	if rl.IsKeyDown(rl.KeyUp) {
		currentKeyboardImage = "upArrow"
	}
	if rl.IsKeyDown(rl.KeyDown) {
		currentKeyboardImage = "downArrow"
	}
	if rl.IsKeyDown(rl.KeySpace) {
		currentKeyboardImage = "space"
	}

}

func HandleMouseEvents(scene *core.Scene) {
	mousePos := rl.GetMousePosition()

	if isFirstFrame {
		lastMousePos = mousePos
		isFirstFrame = false
		return
	}

	delta := rl.Vector2Subtract(mousePos, lastMousePos)

	// --- Middle mouse pan ---
	if rl.IsMouseButtonDown(rl.MouseMiddleButton) {
		panSpeed := 0.05
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

func handleWindowResize(scene *core.Scene) {
	if !rl.IsWindowReady() {
		return
	}

	newWidth := max(300, int(rl.GetScreenWidth()))
	newHeight := max(200, int(rl.GetScreenHeight()))

	// Update global dimensions
	core.SCREEN_WIDTH = newWidth
	core.SCREEN_HEIGHT = newHeight

	// Calculate render dimensions based on half-resolution setting
	renderWidth := newWidth
	renderHeight := newHeight
	if scene.HalfResolution {
		renderWidth = newWidth / 2
		renderHeight = newHeight / 2
	}

	// Ensure minimum size
	renderWidth = max(1, renderWidth)
	renderHeight = max(1, renderHeight)

	// Resize render buffers
	scene.Renderer.Resize(renderWidth, renderHeight)

	// Update camera projection
	scene.Camera.FocalLength = int(float64(scene.Camera.FocalLength) *
		float64(newWidth) / float64(core.SCREEN_WIDTH))
	scene.Camera.Transform.UpdateModelMatrix()
	scene.Camera.UpdateFrustumPlanes()
}
