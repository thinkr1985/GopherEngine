package guis

import (
	"GopherEngine/widgets"
	_ "GopherEngine/widgets"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type PropertiesPanel struct {
	Layout      *Layout
	Bounds      rl.Rectangle
	Title       string
	IsVisible   bool
	BgColor     rl.Color
	BorderColor rl.Color
	Content     func() // Function to draw panel content
	TabWidget   widgets.TabWidget
	RenderPanel *RenderPanel // Add reference to RenderPanel
}

func NewPropertiesPanel(layout *Layout) *PropertiesPanel {
	pr := &PropertiesPanel{
		Layout:      layout,
		Title:       "PropertiesTab",
		IsVisible:   true,
		BgColor:     rl.NewColor(30, 30, 30, 255),
		BorderColor: rl.NewColor(60, 60, 60, 255),
		TabWidget:   *widgets.NewTabWidget(10, 10, 400, 500, 25, 12),
		RenderPanel: NewRenderPanel(layout), // Initialize RenderPanel
	}

	// Set up the Render tab with the RenderPanel content
	pr.TabWidget.AddCloseableTab("Render", func() {
		// Draw the RenderPanel contents
		pr.RenderPanel.DrawContents()
	})

	pr.TabWidget.AddCloseableTab("Configurations", func() {
		rl.DrawText("Configuration settings will go here", 10, 10, 14, rl.White)
	})

	return pr
}

func (pr *PropertiesPanel) Update(targetWidth, targetHeight float32) {
	pr.Layout.RenderPanel.Bounds = rl.NewRectangle(
		pr.Layout.SceneExplorerWidth,
		pr.Layout.MenuBarHeight,
		targetWidth,
		targetHeight,
	)

	widget_x_pos := float32(pr.Layout.SceneExplorer.Bounds.Width+pr.Layout.RenderPanel.Bounds.Width) + 5
	pr.TabWidget.Bounds = rl.NewRectangle(
		widget_x_pos,
		70.0,
		pr.Bounds.Width-10,
		pr.Bounds.Height-50,
	)

	// Update the RenderPanel position
	pr.RenderPanel.Bounds.X = widget_x_pos + 5
	pr.RenderPanel.Bounds.Y = 80
	pr.RenderPanel.Update(pr.TabWidget.Bounds.Width, pr.TabWidget.Bounds.Height)
	pr.TabWidget.Update()
}

func (pr *PropertiesPanel) Draw() {
	if !pr.IsVisible {
		return
	}

	rl.DrawRectangleRec(pr.Bounds, pr.BgColor)
	rl.DrawRectangleLinesEx(pr.Bounds, 1, pr.BorderColor)

	titleHeight := float32(20)
	titleRect := rl.NewRectangle(
		pr.Bounds.X, pr.Bounds.Y,
		pr.Bounds.Width, titleHeight,
	)

	rl.DrawRectangleRec(titleRect, rl.NewColor(30, 30, 30, 255))
	rl.DrawRectangleLinesEx(titleRect, 1, rl.NewColor(60, 60, 60, 255))

	if pr.Content != nil {
		contentRect := rl.NewRectangle(
			pr.Bounds.X,
			pr.Bounds.Y+titleHeight,
			pr.Bounds.Width,
			pr.Bounds.Height-titleHeight,
		)

		rl.BeginScissorMode(
			int32(contentRect.X),
			int32(contentRect.Y),
			int32(contentRect.Width),
			int32(contentRect.Height),
		)

		pr.Content()

		rl.EndScissorMode()
	}
}

func (pr *PropertiesPanel) DrawContents() {
	pr.TabWidget.Draw()
	pr.RenderPanel.DrawContents()
}
func (pr *PropertiesPanel) SetupRenderPanel() {

}
