package gui

import (
	"GopherEngine/core"
	"fmt"
	"image"
	"image/color"
	"math"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var engine_icon_path = "sources/go_engine_ico.png"
var debugFont rl.Font
var isFirstFrame = true

func initWindow() {
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(int32(core.SCREEN_WIDTH), int32(core.SCREEN_HEIGHT), "Gopher Engine")

	debugFont = rl.LoadFontEx("fonts/CONSOLA.TTF", 12, nil, 0)

	icon := rl.LoadImage(engine_icon_path)
	rl.SetWindowIcon(*icon)
	rl.UnloadImage(icon)

	rl.SetTargetFPS(120)

}

func Window(scene *core.Scene) {
	initWindow()
	defer rl.CloseWindow()
	defer rl.UnloadFont(debugFont)

	// Initialize resolution scaling
	scene.ResolutionScale = 1.0
	scene.AutoResolution = false
	scene.LastFPS = 60
	scene.MinResolutionScale = 0.1
	scene.LastScaleChange = rl.GetTime()
	scene.FPSHistory = make([]int, 0, 10)

	keyboardTextures := generateKeybaordTextureMap()
	defer func() {
		for _, tex := range keyboardTextures {
			rl.UnloadTexture(tex)
		}
	}()

	// Create initial texture
	var fullResTex rl.Texture2D
	defer rl.UnloadTexture(fullResTex)

	// Start with a 1x1 black texture
	initialPixels := []color.RGBA{{R: 0, G: 0, B: 0, A: 255}}
	fullResTex = rl.LoadTextureFromImage(&rl.Image{
		Data:    unsafe.Pointer(&initialPixels[0]),
		Width:   1,
		Height:  1,
		Mipmaps: 1,
		Format:  rl.PixelFormat(7),
	})

	for !rl.WindowShouldClose() {
		frameTime := rl.GetFrameTime()
		currentTime := rl.GetTime()
		currentFPS := rl.GetFPS()

		// Update FPS history
		if len(scene.FPSHistory) >= 120 {
			scene.FPSSum -= scene.FPSHistory[0]
			scene.FPSHistory = scene.FPSHistory[1:]
		}
		scene.FPSHistory = append(scene.FPSHistory, int(currentFPS))
		scene.FPSSum += int(currentFPS)

		// Handle auto-resolution
		if scene.AutoResolution {
			if currentTime-scene.LastScaleChange > 0.5 {
				smoothedFPS := scene.FPSSum / len(scene.FPSHistory)
				updateTargetResolution(scene, smoothedFPS, currentTime)
			}
			adjustResolutionGradually(scene, float64(frameTime))
		}

		handleWindowResize(scene)
		HandleInputEvents(scene)

		// Render 3D scene
		scene.ViewAxes.Draw(scene.Renderer, scene.Camera)
		scene.Grid.Draw(scene.Renderer, scene.Camera)
		// scene.RenderWithTiling(128)
		scene.RenderOnThread()
		// Get rendered image and convert to RGBA
		rawImage := scene.Renderer.ToImage()
		rgbaSlice := convertToColorRGBASlice(rawImage)

		// Check if we need to resize texture
		imgWidth := rawImage.Bounds().Dx()
		imgHeight := rawImage.Bounds().Dy()
		if int(fullResTex.Width) != imgWidth || int(fullResTex.Height) != imgHeight {
			rl.UnloadTexture(fullResTex)
			fullResTex = rl.LoadTextureFromImage(&rl.Image{
				Data:    unsafe.Pointer(&rgbaSlice[0]),
				Width:   int32(imgWidth),
				Height:  int32(imgHeight),
				Mipmaps: 1,
				Format:  rl.PixelFormat(7),
			})
		} else {
			// Update existing texture
			rl.UpdateTexture(fullResTex, rgbaSlice)
		}

		// Draw everything
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		rl.DrawTexturePro(
			fullResTex,
			rl.NewRectangle(0, 0, float32(fullResTex.Width), float32(fullResTex.Height)),
			rl.NewRectangle(0, 0, float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight())),
			rl.NewVector2(0, 0),
			0,
			rl.White,
		)

		rl.DrawFPS(20, 20)
		draw_debug_stats(scene)
		drawKeyboardOverlay(keyboardTextures[currentKeyboardImage])

		rl.EndDrawing()
	}
}

func updateTargetResolution(scene *core.Scene, currentFPS int, currentTime float64) {
	// Calculate ideal scale based on FPS (inverse relationship)
	// These values can be tweaked to get the desired behavior
	minFPS := 15.0
	maxFPS := 40.0
	fpsRatio := math.Min(1.0, math.Max(0.0,
		(float64(currentFPS)-minFPS)/(maxFPS-minFPS)))

	// Map FPS ratio to resolution scale (quadratic easing for smoother transitions)
	newTarget := scene.MinResolutionScale +
		(1.0-scene.MinResolutionScale)*fpsRatio*fpsRatio

	// Only update target if significantly different
	if math.Abs(newTarget-scene.TargetResolutionScale) > 0.05 {
		scene.TargetResolutionScale = newTarget
		scene.LastScaleChange = currentTime
	}
}

func adjustResolutionGradually(scene *core.Scene, frameTime float64) {
	// Calculate maximum allowed change this frame
	maxChange := scene.ResolutionChangeSpeed * frameTime

	if scene.ResolutionScale < scene.TargetResolutionScale {
		// Move upward toward target
		scene.ResolutionScale = math.Min(
			scene.TargetResolutionScale,
			scene.ResolutionScale+maxChange)
	} else if scene.ResolutionScale > scene.TargetResolutionScale {
		// Move downward toward target
		scene.ResolutionScale = math.Max(
			scene.TargetResolutionScale,
			scene.ResolutionScale-maxChange)
	}

	// Ensure we stay within bounds
	scene.ResolutionScale = math.Max(scene.MinResolutionScale,
		math.Min(1.0, scene.ResolutionScale))

	// Resize will happen in next handleWindowResize call
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func draw_debug_stats(scene *core.Scene) {
	avgFPS := 0
	if len(scene.FPSHistory) > 0 {
		avgFPS = scene.FPSSum / len(scene.FPSHistory)
	}

	statsText := fmt.Sprintf("%s\nFPS: %d (Avg: %d)\nResolution: %.0f%% (Target: %.0f%%)\nAuto-Res: %v\nScene Triangles : %v/%v",
		core.GetMachineStats(),
		rl.GetFPS(),
		avgFPS,
		scene.ResolutionScale*100,
		scene.TargetResolutionScale*100,
		scene.AutoResolution,
		scene.DrawnTriangles,
		len(scene.Triangles))

	textWidth := rl.MeasureText(statsText, 12)
	rl.DrawRectangle(10, 10, textWidth+80, 150, rl.NewColor(0, 0, 0, 60))
	rl.DrawTextEx(debugFont, statsText, rl.NewVector2(20, 40), 12, 2, rl.LightGray)

	// Show scaling info if in auto mode
	if scene.AutoResolution {
		scalingText := fmt.Sprintf("Scaling: %.1f%%/s", scene.ResolutionChangeSpeed*100)
		rl.DrawTextEx(debugFont, scalingText, rl.NewVector2(20, 140), 12, 2, rl.LightGray)
	}
}

func drawKeyboardOverlay(tex rl.Texture2D) {
	x := 20
	y := rl.GetScreenHeight() - int(tex.Height) - 20
	rl.DrawTexture(tex, int32(x), int32(y), rl.White)
}

/*
func convertToColorRGBASlice(img *image.RGBA) []color.RGBA {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	src := img.Pix
	pixels := make([]color.RGBA, w*h)

	for i := 0; i < len(pixels); i++ {
		pixels[i] = color.RGBA{
			R: src[i*4],
			G: src[i*4+1],
			B: src[i*4+2],
			A: src[i*4+3],
		}
	}

	return pixels
}
*/

func convertToColorRGBASlice(img *image.RGBA) []color.RGBA {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	src := img.Pix
	pixels := make([]color.RGBA, w*h)

	for i := 0; i < len(pixels); i++ {
		pixels[i] = *(*color.RGBA)(unsafe.Pointer(&src[i*4]))
	}

	return pixels
}
