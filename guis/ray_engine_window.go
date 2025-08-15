package guis

import (
	"GopherEngine/core"
	"GopherEngine/widgets"
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
var display_debug_screen = false
var gui_widgets []interface{}
var appLayout *Layout

func initWindow() {
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(int32(core.SCREEN_WIDTH), int32(core.SCREEN_HEIGHT), "Gopher Engine")

	debugFont = rl.LoadFontEx("fonts/CONSOLA.TTF", 12, nil, 0)

	icon := rl.LoadImage(engine_icon_path)
	rl.SetWindowIcon(*icon)
	rl.UnloadImage(icon)

	rl.SetTargetFPS(240)

	// Initialize layout
	appLayout = NewLayout()

	// Set up panel contents
	appLayout.MenuBar.Content = drawMenuBar
	appLayout.SceneExplorer.Content = drawSceneExplorer
	appLayout.AttributeEditor.Content = drawAttributeEditor
	appLayout.AssetBrowser.Content = drawAssetBrowser

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

		handleWindowResize(scene)
		HandleInputEvents(scene)
		appLayout.HandleResize()

		// Render 3D
		scene.Render()

		// Update layout with current window size
		appLayout.Update(int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()))

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
			rl.UpdateTexture(fullResTex, rgbaSlice)
		}

		// Set render panel content function
		appLayout.RenderPanel.Content = func() {
			// Calculate the actual render area (excluding title bar)
			renderArea := rl.NewRectangle(
				appLayout.RenderPanel.Bounds.X,
				appLayout.RenderPanel.Bounds.Y+20, // Below title
				appLayout.RenderPanel.Bounds.Width,
				appLayout.RenderPanel.Bounds.Height-20, // Subtract title height
			)

			// Calculate source rectangle (entire rendered image)
			srcRect := rl.NewRectangle(0, 0, float32(fullResTex.Width), float32(fullResTex.Height))

			// Calculate destination rectangle (centered and maintaining aspect ratio)
			destWidth := renderArea.Width
			destHeight := renderArea.Height

			// Maintain 16:9 aspect ratio
			targetAspect := float32(16) / 9
			currentAspect := destWidth / destHeight

			var destRect rl.Rectangle

			if currentAspect > targetAspect {
				// Window is wider than 16:9 - add horizontal padding
				newWidth := destHeight * targetAspect
				xOffset := (destWidth - newWidth) / 2
				destRect = rl.NewRectangle(
					renderArea.X+xOffset,
					renderArea.Y,
					newWidth,
					destHeight,
				)
			} else {
				// Window is taller than 16:9 - add vertical padding
				newHeight := destWidth / targetAspect
				yOffset := (destHeight - newHeight) / 2
				destRect = rl.NewRectangle(
					renderArea.X,
					renderArea.Y+yOffset,
					destWidth,
					newHeight,
				)
			}

			// Draw the rendered texture in the render panel
			rl.DrawTexturePro(
				fullResTex,
				srcRect,
				destRect,
				rl.NewVector2(0, 0),
				0,
				rl.White,
			)

			if display_debug_screen {
				draw_debug_stats(scene, appLayout.SceneExplorer)
				draw_threading_status(scene, appLayout.SceneExplorer)
				drawKeyboardOverlay(keyboardTextures[currentKeyboardImage])
			}
		}
		// Draw everything
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		// Draw the layout
		appLayout.Draw()

		rl.EndDrawing()
	}
}

// Panel content functions
func drawMenuBar() {
	// Draw menu items here
	rl.DrawText("File", 10, 5, 12, rl.White)
	rl.DrawText("Edit", 60, 5, 12, rl.White)
	rl.DrawText("View", 110, 5, 12, rl.White)
	rl.DrawText("Help", 160, 5, 12, rl.White)
}

func drawSceneExplorer() {

}

func drawAttributeEditor() {

}

func drawAssetBrowser() {

}

func load_gui_layout() {
	slider := widgets.NewFloatSlider(10, 100, 100, 20, 1, 200, 100, "Camera")
	checkBox := widgets.NewCheckBox(10, 200, 20, 20, "Camera")
	toggle := widgets.NewToggle(1, 300, "Camera")
	colorWheel := widgets.NewColorWheel(300, 300, 50, 12)
	dropdown := widgets.NewDropdown(500, 50, 200, 30, 20)
	dropdown.AddOption("Option 1", "value1")
	dropdown.AddOption("Option 2", "value2")
	dropdown.AddOption("Option 3", "value3")
	button := widgets.NewPushButton(700, 50, 200, 40, "Click Me", 20)
	msgBox := widgets.NewMessageBox("sdsdsdsd", "fndfbdffdfdksfbjdsbfkbkadsjfkasd", 1)

	treeWidget := widgets.NewTreeWidget(50, 50, 300, 500, 20)

	// Create the root and add some child nodes
	root := treeWidget.AddNode("Root", nil)
	treeWidget.AddNode("Child 1", root)
	treeWidget.AddNode("Child 2", root)
	treeWidget.AddNode("Child 3", root)

	// Create more nested nodes
	child2 := treeWidget.AddNode("Child 2-1", root.Children[1])
	treeWidget.AddNode("Child 2-1-1", child2)

	// Create tab widget
	tabWidget := widgets.NewTabWidget(50, 50, 600, 400, 30, 20)

	// Add regular tab
	tabWidget.AddTab("Tab 1", func() {
		rl.DrawText("Content for Tab 1", 100, 100, 20, rl.Black)
	})

	// Add closeable tab
	tabWidget.AddCloseableTab("Tab 2", func() {
		rl.DrawText("Closeable Tab Content", 100, 100, 20, rl.Black)
	})

	gui_widgets = append(gui_widgets, slider)
	gui_widgets = append(gui_widgets, checkBox)
	gui_widgets = append(gui_widgets, toggle)
	gui_widgets = append(gui_widgets, colorWheel)
	gui_widgets = append(gui_widgets, dropdown)
	gui_widgets = append(gui_widgets, button)
	gui_widgets = append(gui_widgets, msgBox)
	gui_widgets = append(gui_widgets, treeWidget)
	gui_widgets = append(gui_widgets, tabWidget)

}

func draw_gui_panels() {
	for _, widget := range gui_widgets {
		switch w := widget.(type) {
		case *widgets.CheckBox:
			w.Update()
			w.Draw()

		case *widgets.ColorWheel:
			w.Update()
			w.Draw()
		default:
			continue
		}
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
	if math.Abs(newTarget-scene.TargetResolutionScale) > 0.1 {
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

}

func draw_debug_stats(scene *core.Scene, render_panel *Panel) {
	if !display_debug_screen {
		return
	}
	avgFPS := 0
	if len(scene.FPSHistory) > 0 {
		avgFPS = scene.FPSSum / len(scene.FPSHistory)
	}
	render_width := scene.Renderer.GetWidth()
	render_height := scene.Renderer.GetHeight()
	statsText := fmt.Sprintf("Resolution : %dx%d\n%s\nFPS: %d (Avg: %d)\nResolution: %.0f%% (Target: %.0f%%)\nAuto-Res: %v\nScene Triangles : %v/%v\nCPU : %v\nGPU : %v",
		render_width, render_height,
		core.GetMachineStats(),
		rl.GetFPS(),
		avgFPS,
		scene.ResolutionScale*100,
		scene.TargetResolutionScale*100,
		scene.AutoResolution,
		scene.DrawnTriangles,
		scene.TotalTriangleCounter,
		scene.Renderer.CPU,
		scene.Renderer.GPU)

	textWidth := rl.MeasureText(statsText, 12)
	rl.DrawRectangle(render_panel.Bounds.ToInt32().Width+10, 40, textWidth+100, 190, rl.NewColor(0, 0, 0, 30))
	rl.DrawTextEx(debugFont, statsText, rl.NewVector2(render_panel.Bounds.Width+20, 80), 12, 2, rl.LightGray)

	// Show scaling info if in auto mode
	if scene.AutoResolution {
		scalingText := fmt.Sprintf("Scaling: %.1f%%/s", scene.ResolutionChangeSpeed*100)
		rl.DrawTextEx(debugFont, scalingText, rl.NewVector2(render_panel.Bounds.Width+180, 20), 12, 2, rl.LightGray)
	}
}

func draw_threading_status(scene *core.Scene, panel *Panel) {
	thread_text := fmt.Sprintf("Multi-Threading (F3) : %v", scene.Renderer.MultiThreading)

	rl.DrawTextEx(
		debugFont, thread_text,
		rl.NewVector2(panel.Bounds.Width+200, 80),
		12, 2, rl.White)

}

func drawKeyboardOverlay(tex rl.Texture2D) {
	if !display_debug_screen {
		return
	}
	x := 20
	y := rl.GetScreenHeight() - int(tex.Height) - 20
	rl.DrawTexture(tex, int32(x), int32(y), rl.White)
}

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
