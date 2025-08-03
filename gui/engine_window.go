package gui

import (
	"GopherEngine/core"
	"GopherEngine/lookdev"
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

	// Initialize resolution scaling with proper types
	scene.ResolutionScale = 1.0
	scene.AutoResolution = false
	scene.LastFPS = 60
	scene.MinResolutionScale = 0.1
	scene.LastScaleChange = rl.GetTime() // Initialize with current time
	scene.FPSHistory = make([]int, 0, 10)

	keyboardTextures := generateKeybaordTextureMap()
	defer func() {
		for _, tex := range keyboardTextures {
			rl.UnloadTexture(tex)
		}
	}()

	// Declare texture variables
	var fullResTex rl.Texture2D
	defer rl.UnloadTexture(fullResTex)

	// Create a render texture for half-resolution rendering
	halfResRenderTex := rl.LoadRenderTexture(
		int32(core.SCREEN_WIDTH/2),
		int32(core.SCREEN_HEIGHT/2),
	)
	defer rl.UnloadRenderTexture(halfResRenderTex)
	rl.SetTextureFilter(halfResRenderTex.Texture, rl.FilterBilinear)

	for !rl.WindowShouldClose() {
		frameTime := rl.GetFrameTime() // Time since last frame in seconds
		currentTime := rl.GetTime()
		currentFPS := rl.GetFPS()

		// Update FPS history for smoothing
		if len(scene.FPSHistory) >= 120 {
			scene.FPSSum -= scene.FPSHistory[0]
			scene.FPSHistory = scene.FPSHistory[1:]
		}
		scene.FPSHistory = append(scene.FPSHistory, int(currentFPS))
		scene.FPSSum += int(currentFPS)

		// Auto-resolution scaling logic
		if scene.AutoResolution {
			// Only update target scale every 0.5 seconds (cooldown period)
			if currentTime-scene.LastScaleChange > 0.5 {
				smoothedFPS := scene.FPSSum / len(scene.FPSHistory)
				updateTargetResolution(scene, smoothedFPS, currentTime)
			}

			// Gradually move toward target resolution
			adjustResolutionGradually(scene, float64(frameTime))
		}

		handleWindowResize(scene)
		HandleInputEvents(scene)

		// Clear will use correct dimensions
		if len(scene.Renderer.Framebuffer) > 0 {
			scene.Renderer.Clear(lookdev.ColorRGBA{R: 60, G: 73, B: 78, A: 1.0})
		}

		// Draw 3D content (now at lower resolution if HalfResolution=true)
		scene.ViewAxes.Draw(scene.Renderer, scene.Camera)
		scene.Grid.Draw(scene.Renderer, scene.Camera)
		scene.RenderScene()

		// Get the rendered image (smaller if half-res)
		renderedImage := scene.Renderer.ToImage()

		// Convert to raylib texture
		rgbaSlice := convertToColorRGBASlice(renderedImage)
		rlImg := rl.Image{
			Data:    unsafe.Pointer(&rgbaSlice[0]),
			Width:   int32(renderedImage.Bounds().Dx()),
			Height:  int32(renderedImage.Bounds().Dy()),
			Mipmaps: 1,
			Format:  rl.UncompressedR8g8b8a8,
		}

		// Unload previous texture if it exists
		if fullResTex.ID != 0 {
			rl.UnloadTexture(fullResTex)
		}

		// Load new texture
		fullResTex = rl.LoadTextureFromImage(&rlImg)
		rl.SetTextureFilter(fullResTex, rl.FilterBilinear)

		// Begin drawing
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
		// Draw UI elements (always at full resolution)
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

	statsText := fmt.Sprintf("%s\nFPS: %d (Avg: %d)\nResolution: %.0f%% (Target: %.0f%%)\nAuto-Res: %v",
		core.GetMachineStats(),
		rl.GetFPS(),
		avgFPS,
		scene.ResolutionScale*100,
		scene.TargetResolutionScale*100,
		scene.AutoResolution)

	textWidth := rl.MeasureText(statsText, 12)
	rl.DrawRectangle(10, 10, textWidth+80, 140, rl.NewColor(0, 0, 0, 60))
	rl.DrawTextEx(debugFont, statsText, rl.NewVector2(20, 40), 12, 2, rl.LightGray)

	// Show scaling info if in auto mode
	if scene.AutoResolution {
		scalingText := fmt.Sprintf("Scaling: %.1f%%/s", scene.ResolutionChangeSpeed*100)
		rl.DrawTextEx(debugFont, scalingText, rl.NewVector2(20, 130), 12, 2, rl.LightGray)
	}
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

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			pixels[y*w+x] = color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			}
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
