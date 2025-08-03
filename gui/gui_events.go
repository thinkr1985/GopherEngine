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
	currentKeyboardImage = "default" // reset every frame
	HandleKeyboardEvents(scene)
	HandleMouseEvents(scene)
}

func HandleKeyboardEvents(scene *core.Scene) {
	moveSpeed := 0.1
	rotateSpeed := 0.02

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
