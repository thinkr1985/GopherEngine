package guis

import (
	"GopherEngine/core"
	"GopherEngine/widgets"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	resizeHandleSize    = 16
	resizeHandleHitArea = 8
)

type ResizeState struct {
	IsResizing      bool
	ResizeDirection int // 0 = left, 1 = right
	StartMouseX     float32
	StartWidth      float32
}

type Panel struct {
	Bounds      rl.Rectangle
	Title       string
	IsVisible   bool
	Content     func() // Function to draw panel content
	BgColor     rl.Color
	BorderColor rl.Color
}

type Layout struct {
	Scene                *core.Scene
	MenuBarHeight        float32
	SceneExplorerWidth   float32
	AttributeEditorWidth float32

	MenuBar       *Panel
	SceneExplorer *Panel
	// RenderPanel     *RenderPanel
	RenderPanel     *PropertiesPanel
	AttributeEditor *Panel
	AssetBrowser    *Panel
	resizeState     ResizeState
}

func NewLayout(scene *core.Scene) *Layout {
	layout := &Layout{
		Scene:                scene,
		MenuBarHeight:        30,
		SceneExplorerWidth:   250,
		AttributeEditorWidth: 300,
	}
	widgets.InitializeWidgetFont()

	layout.MenuBar = &Panel{
		Title:       "Menu",
		IsVisible:   true,
		BgColor:     rl.NewColor(50, 50, 50, 255),
		BorderColor: rl.NewColor(80, 80, 80, 255),
	}

	layout.SceneExplorer = &Panel{
		Title:       "Scene Explorer",
		IsVisible:   true,
		BgColor:     rl.NewColor(40, 40, 40, 255),
		BorderColor: rl.NewColor(70, 70, 70, 255),
	}

	properties_panel := NewPropertiesPanel(layout)
	layout.RenderPanel = properties_panel

	layout.AttributeEditor = &Panel{
		Title:       "Properties",
		IsVisible:   true,
		BgColor:     rl.NewColor(40, 40, 40, 255),
		BorderColor: rl.NewColor(70, 70, 70, 255),
	}

	layout.AssetBrowser = &Panel{
		Title:       "Asset Browser",
		IsVisible:   true,
		BgColor:     rl.NewColor(40, 40, 40, 255),
		BorderColor: rl.NewColor(70, 70, 70, 255),
	}

	return layout
}

func (l *Layout) Update(screenWidth, screenHeight int32) {
	l.MenuBar.Bounds = rl.NewRectangle(
		0, 0,
		float32(screenWidth), l.MenuBarHeight,
	)

	mainContentHeight := float32(screenHeight) - l.MenuBarHeight

	l.SceneExplorer.Bounds = rl.NewRectangle(
		0, l.MenuBarHeight,
		l.SceneExplorerWidth, mainContentHeight,
	)

	l.AttributeEditor.Bounds = rl.NewRectangle(
		float32(screenWidth)-l.AttributeEditorWidth, l.MenuBarHeight,
		l.AttributeEditorWidth, mainContentHeight,
	)

	availableWidth := float32(screenWidth) - l.SceneExplorerWidth - l.AttributeEditorWidth
	availableHeight := mainContentHeight

	targetWidth := availableWidth
	targetHeight := targetWidth * 9 / 16

	if targetHeight > availableHeight {
		targetHeight = availableHeight
		targetWidth = targetHeight * 16 / 9
	}
	l.RenderPanel.Update(targetWidth, targetHeight)

	remainingHeight := mainContentHeight - targetHeight
	assetBrowserHeight := max(100, remainingHeight)

	l.AssetBrowser.Bounds = rl.NewRectangle(
		0,
		l.MenuBarHeight+targetHeight,
		float32(screenWidth),
		assetBrowserHeight,
	)

	l.AttributeEditor.Bounds.Height = targetHeight
	l.SceneExplorer.Bounds.Height = targetHeight
}
func (l *Layout) DrawPanel(p *Panel) {
	if !p.IsVisible {
		return
	}

	// Draw panel background and title (existing code)
	rl.DrawRectangleRec(p.Bounds, p.BgColor)
	rl.DrawRectangleLinesEx(p.Bounds, 1, p.BorderColor)

	titleHeight := float32(30)
	if p.Title == "Menu" {
		titleHeight = 0
	}

	titleRect := rl.NewRectangle(
		p.Bounds.X, p.Bounds.Y,
		p.Bounds.Width, titleHeight,
	)

	rl.DrawRectangleRec(titleRect, rl.NewColor(30, 30, 30, 255))
	rl.DrawRectangleLinesEx(titleRect, 1, rl.NewColor(60, 60, 60, 255))

	if p.Title != "Menu" {
		textX := p.Bounds.X + 5
		textY := p.Bounds.Y + (titleHeight-12)/2

		rl.DrawTextEx(
			widgets.Widget_default_font,
			p.Title,
			rl.NewVector2(textX, textY),
			14,
			1.0,
			rl.White,
		)
	}

	// Draw content if available (existing code)
	if p.Content != nil {
		contentRect := rl.NewRectangle(
			p.Bounds.X,
			p.Bounds.Y+titleHeight,
			p.Bounds.Width,
			p.Bounds.Height-titleHeight,
		)

		rl.BeginScissorMode(
			int32(contentRect.X),
			int32(contentRect.Y),
			int32(contentRect.Width),
			int32(contentRect.Height),
		)

		p.Content()

		rl.EndScissorMode()
	}
}

// Add these new methods to handle resize handles
func (l *Layout) drawResizeHandles(p *Panel) {
	mousePos := rl.GetMousePosition()
	isLeftHandle := p == l.SceneExplorer
	isRightHandle := p == l.AttributeEditor

	// Left handle (for SceneExplorer)
	if isLeftHandle {
		handleRect := rl.NewRectangle(
			p.Bounds.X+p.Bounds.Width-resizeHandleSize,
			p.Bounds.Y+p.Bounds.Height-resizeHandleSize,
			resizeHandleSize,
			resizeHandleSize,
		)

		// Check if mouse is hovering over the handle
		isHovering := rl.CheckCollisionPointRec(mousePos, handleRect)
		color := rl.NewColor(100, 100, 100, 255)
		if isHovering {
			color = rl.NewColor(200, 200, 200, 255)
		}

		// Draw L-shaped triangle
		points := []rl.Vector2{
			{handleRect.X, handleRect.Y + handleRect.Height},
			{handleRect.X + handleRect.Width, handleRect.Y + handleRect.Height},
			{handleRect.X + handleRect.Width, handleRect.Y},
		}
		rl.DrawTriangle(points[0], points[1], points[2], color)
	}

	// Right handle (for AttributeEditor)
	if isRightHandle {
		handleRect := rl.NewRectangle(
			p.Bounds.X,
			p.Bounds.Y+p.Bounds.Height-resizeHandleSize,
			resizeHandleSize,
			resizeHandleSize,
		)

		// Check if mouse is hovering over the handle
		isHovering := rl.CheckCollisionPointRec(mousePos, handleRect)
		color := rl.NewColor(100, 100, 100, 255)
		if isHovering {
			color = rl.NewColor(200, 200, 200, 255)
		}

		// Draw mirrored L-shaped triangle
		points := []rl.Vector2{
			{handleRect.X, handleRect.Y},
			{handleRect.X, handleRect.Y + handleRect.Height},
			{handleRect.X + handleRect.Width, handleRect.Y + handleRect.Height},
		}
		rl.DrawTriangle(points[0], points[1], points[2], color)
	}
}

func (l *Layout) HandleResize() {
	mousePos := rl.GetMousePosition()
	isHoveringHandle := false

	// Check left handle (SceneExplorer)
	leftHandleRect := rl.NewRectangle(
		l.SceneExplorer.Bounds.X+l.SceneExplorer.Bounds.Width-resizeHandleSize,
		l.SceneExplorer.Bounds.Y+l.SceneExplorer.Bounds.Height-resizeHandleSize,
		resizeHandleSize,
		resizeHandleSize,
	)

	// Check right handle (AttributeEditor)
	rightHandleRect := rl.NewRectangle(
		l.AttributeEditor.Bounds.X,
		l.AttributeEditor.Bounds.Y+l.AttributeEditor.Bounds.Height-resizeHandleSize,
		resizeHandleSize,
		resizeHandleSize,
	)

	if rl.CheckCollisionPointRec(mousePos, leftHandleRect) || rl.CheckCollisionPointRec(mousePos, rightHandleRect) {
		isHoveringHandle = true
		rl.SetMouseCursor(rl.MouseCursorResizeEW)
	} else if !l.resizeState.IsResizing {
		rl.SetMouseCursor(rl.MouseCursorDefault)
	}

	// Check if we should start resizing
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && isHoveringHandle {
		if rl.CheckCollisionPointRec(mousePos, leftHandleRect) {
			l.resizeState.IsResizing = true
			l.resizeState.ResizeDirection = 0 // left
			l.resizeState.StartMouseX = mousePos.X
			l.resizeState.StartWidth = l.SceneExplorerWidth
		} else if rl.CheckCollisionPointRec(mousePos, rightHandleRect) {
			l.resizeState.IsResizing = true
			l.resizeState.ResizeDirection = 1 // right
			l.resizeState.StartMouseX = mousePos.X
			l.resizeState.StartWidth = l.AttributeEditorWidth
		}
	}

	// Handle ongoing resize
	if l.resizeState.IsResizing && rl.IsMouseButtonDown(rl.MouseLeftButton) {
		delta := mousePos.X - l.resizeState.StartMouseX

		if l.resizeState.ResizeDirection == 0 { // Left resize (SceneExplorer)
			newWidth := l.resizeState.StartWidth + delta
			l.SceneExplorerWidth = max(100, min(newWidth, float32(rl.GetScreenWidth())*0.5))
		} else { // Right resize (AttributeEditor)
			newWidth := l.resizeState.StartWidth - delta
			l.AttributeEditorWidth = max(100, min(newWidth, float32(rl.GetScreenWidth())*0.5))
		}

		// Update layout immediately during resize
		l.Update(int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()))
	} else {
		l.resizeState.IsResizing = false
	}
}

func (l *Layout) Draw() {
	l.DrawPanel(l.MenuBar)
	l.DrawPanel(l.SceneExplorer)
	// l.DrawPanel(l.RenderPanel)
	l.RenderPanel.Draw()
	l.DrawPanel(l.AttributeEditor)
	l.DrawPanel(l.AssetBrowser)
	l.RenderPanel.DrawContents()

	// Only draw resize handles for SceneExplorer and AttributeEditor
	l.drawResizeHandles(l.SceneExplorer)
	l.drawResizeHandles(l.AttributeEditor)
}

func max(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
