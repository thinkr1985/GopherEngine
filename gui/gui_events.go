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
		scene.AutoResolution = !scene.AutoResolution
		if !scene.AutoResolution {
			// Reset to full resolution when turning off auto-scaling
			scene.ResolutionScale = 1.0
		}
		handleWindowResize(scene)
	}

	if rl.IsKeyPressed(rl.KeyF2) {
		if display_debug_screen {
			display_debug_screen = false
		} else {
			display_debug_screen = true
		}

	}

	if rl.IsWindowReady() {
		HandleKeyboardEvents(scene)
		HandleMouseEvents(scene)
	}
}

func HandleKeyboardEvents(scene *core.Scene) {
	moveSpeed := 1.5
	rotateSpeed := 0.02

	// Get camera vectors (note these are in world space)
	forward := scene.Camera.Transform.GetForward()
	right := scene.Camera.Transform.GetRight()
	up := scene.Camera.Transform.GetUp()

	// Movement controls
	if rl.IsKeyDown(rl.KeyW) {
		scene.Camera.Transform.Translate(forward.Multiply(moveSpeed))
		currentKeyboardImage = "W"
		scene.Camera.DirtyFrustum = true
		scene.Camera.Transform.Dirty = true
	}
	if rl.IsKeyDown(rl.KeyS) {
		scene.Camera.Transform.Translate(forward.Multiply(-moveSpeed))
		currentKeyboardImage = "S"
		scene.Camera.DirtyFrustum = true
		scene.Camera.Transform.Dirty = true
	}
	if rl.IsKeyDown(rl.KeyA) {
		scene.Camera.Transform.Translate(right.Multiply(-moveSpeed))
		currentKeyboardImage = "A"
		scene.Camera.DirtyFrustum = true
		scene.Camera.Transform.Dirty = true
	}
	if rl.IsKeyDown(rl.KeyD) {
		scene.Camera.Transform.Translate(right.Multiply(moveSpeed))
		currentKeyboardImage = "D"
		scene.Camera.DirtyFrustum = true
		scene.Camera.Transform.Dirty = true
	}
	if rl.IsKeyDown(rl.KeyQ) {
		scene.Camera.Transform.Translate(up.Multiply(-moveSpeed))
		currentKeyboardImage = "Q"
		scene.Camera.DirtyFrustum = true
		scene.Camera.Transform.Dirty = true
	}
	if rl.IsKeyDown(rl.KeyE) {
		scene.Camera.Transform.Translate(up.Multiply(moveSpeed))
		currentKeyboardImage = "E"
		scene.Camera.DirtyFrustum = true
		scene.Camera.Transform.Dirty = true
	}

	// Rotation controls
	if rl.IsKeyDown(rl.KeyRight) {
		scene.Camera.Transform.Rotate(nomath.Vec3{Y: -rotateSpeed})
		currentKeyboardImage = "arrowRight"
		scene.Camera.DirtyFrustum = true
		scene.Camera.Transform.Dirty = true
	}
	if rl.IsKeyDown(rl.KeyLeft) {
		scene.Camera.Transform.Rotate(nomath.Vec3{Y: rotateSpeed})
		currentKeyboardImage = "arrowLeft"
		scene.Camera.DirtyFrustum = true
		scene.Camera.Transform.Dirty = true
	}
	if rl.IsKeyDown(rl.KeyUp) {
		scene.Camera.Transform.Rotate(nomath.Vec3{X: -rotateSpeed})
		currentKeyboardImage = "upArrow"
		scene.Camera.DirtyFrustum = true
		scene.Camera.Transform.Dirty = true
	}
	if rl.IsKeyDown(rl.KeyDown) {
		scene.Camera.Transform.Rotate(nomath.Vec3{X: rotateSpeed})
		currentKeyboardImage = "downArrow"
		scene.Camera.DirtyFrustum = true
		scene.Camera.Transform.Dirty = true
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
		scene.Camera.DirtyFrustum = true
		scene.Camera.Transform.Dirty = true
	}

	// --- Scroll to zoom ---
	scroll := rl.GetMouseWheelMove()
	if scroll != 0 {
		zoomSpeed := 1.0
		forward := scene.Camera.Transform.GetForward().Multiply(float64(scroll) * zoomSpeed)
		scene.Camera.Transform.Translate(forward)
		currentKeyboardImage = "scroll"
		scene.Camera.DirtyFrustum = true
		scene.Camera.Transform.Dirty = true
	}

	// --- Left drag to rotate around Y axis ---
	if rl.IsMouseButtonDown(rl.MouseLeftButton) {
		rotationSpeed := 0.002
		scene.Camera.Transform.Rotate(nomath.Vec3{
			Y: -float64(delta.X) * rotationSpeed,
			X: -float64(delta.Y) * rotationSpeed,
		})
		scene.Camera.Transform.Dirty = true
		scene.Camera.Transform.Dirty = true
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

	// Only resize if dimensions actually changed
	// Update global dimensions
	core.SCREEN_WIDTH = newWidth
	core.SCREEN_HEIGHT = newHeight

	// Calculate render dimensions based on resolution scale
	renderWidth := int(float64(newWidth) * scene.ResolutionScale)
	renderHeight := int(float64(newHeight) * scene.ResolutionScale)

	// Ensure minimum size
	renderWidth = max(1, renderWidth)
	renderHeight = max(1, renderHeight)

	// Resize render buffers
	scene.Renderer.Resize(renderWidth, renderHeight)

	// Update camera projection
	scene.Camera.Update() // This will update the projection matrix
}
