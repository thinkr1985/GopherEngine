package gui

import (
	"GopherEngine/core"
	"GopherEngine/lookdev"
	"image"
	"image/color"
	_ "math"
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
		draw_debug_stats()
		drawKeyboardOverlay(keyboardTextures[currentKeyboardImage])

		rl.EndDrawing()
	}
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
