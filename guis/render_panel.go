package guis

import (
	"GopherEngine/widgets"
	_ "GopherEngine/widgets"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type RenderPanel struct {
	Layout              *Layout
	Bounds              rl.Rectangle
	Title               string
	IsVisible           bool
	Content             func() // Function to draw panel content
	BgColor             rl.Color
	BorderColor         rl.Color
	SoftNormalCheck     widgets.CheckBox
	FogCheck            widgets.CheckBox
	DOFCheck            widgets.CheckBox
	ShadowsCheck        widgets.CheckBox
	MultiThreadingCheck widgets.CheckBox
}

func NewRenderPanel(layout *Layout) *RenderPanel {
	rp := &RenderPanel{
		Layout:              layout,
		Title:               "Render View",
		IsVisible:           true,
		BgColor:             rl.NewColor(30, 30, 30, 255),
		BorderColor:         rl.NewColor(60, 60, 60, 255),
		SoftNormalCheck:     *widgets.NewCheckBox(10, 10, 20, 20, "Soft Normals"),
		FogCheck:            *widgets.NewCheckBox(10, 50, 20, 20, "Fog"),
		DOFCheck:            *widgets.NewCheckBox(10, 150, 20, 20, "Depth Of Field"),
		ShadowsCheck:        *widgets.NewCheckBox(10, 150, 20, 20, "Shadows"),
		MultiThreadingCheck: *widgets.NewCheckBox(10, 150, 20, 20, "MultiThreading"),
	}
	rp.SoftNormalCheck.OnToggle = rp.ToggleSoftNormals
	rp.FogCheck.OnToggle = rp.ToggleFog
	rp.DOFCheck.OnToggle = rp.ToggleDOF
	rp.ShadowsCheck.OnToggle = rp.ToggleShadow
	rp.MultiThreadingCheck.OnToggle = rp.ToggleMultitheading
	return rp

}

func (rp *RenderPanel) Update(targetWidth, targetHeight float32) {

	rp.Layout.RenderPanel.Bounds = rl.NewRectangle(
		rp.Layout.SceneExplorerWidth,
		rp.Layout.MenuBarHeight,
		targetWidth,
		targetHeight,
	)
	widget_x_pos := float32(rp.Layout.SceneExplorer.Bounds.Width+rp.Layout.RenderPanel.Bounds.Width) + 10
	rp.FogCheck.Bounds = rl.NewRectangle(widget_x_pos, float32(60), float32(rp.FogCheck.Width), float32(rp.FogCheck.Height))
	rp.SoftNormalCheck.Bounds = rl.NewRectangle(widget_x_pos, float32(90), float32(rp.SoftNormalCheck.Width), float32(rp.SoftNormalCheck.Height))
	rp.DOFCheck.Bounds = rl.NewRectangle(widget_x_pos, float32(120), float32(rp.DOFCheck.Width), float32(rp.DOFCheck.Height))
	rp.ShadowsCheck.Bounds = rl.NewRectangle(widget_x_pos, float32(150), float32(rp.ShadowsCheck.Width), float32(rp.ShadowsCheck.Height))
	rp.MultiThreadingCheck.Bounds = rl.NewRectangle(widget_x_pos, float32(180), float32(rp.MultiThreadingCheck.Width), float32(rp.MultiThreadingCheck.Height))

	rp.MultiThreadingCheck.Update()
	rp.FogCheck.Update()
	rp.SoftNormalCheck.Update()
	rp.DOFCheck.Update()
	rp.ShadowsCheck.Update()
}

func (rp *RenderPanel) Draw() {

	if !rp.IsVisible {
		return
	}

	// Draw panel background and title (existing code)
	rl.DrawRectangleRec(rp.Bounds, rp.BgColor)
	rl.DrawRectangleLinesEx(rp.Bounds, 1, rp.BorderColor)

	titleHeight := float32(20)

	titleRect := rl.NewRectangle(
		rp.Bounds.X, rp.Bounds.Y,
		rp.Bounds.Width, titleHeight,
	)

	rl.DrawRectangleRec(titleRect, rl.NewColor(30, 30, 30, 255))
	rl.DrawRectangleLinesEx(titleRect, 1, rl.NewColor(60, 60, 60, 255))

	// Draw content if available (existing code)
	if rp.Content != nil {
		contentRect := rl.NewRectangle(
			rp.Bounds.X,
			rp.Bounds.Y+titleHeight,
			rp.Bounds.Width,
			rp.Bounds.Height-titleHeight,
		)

		rl.BeginScissorMode(
			int32(contentRect.X),
			int32(contentRect.Y),
			int32(contentRect.Width),
			int32(contentRect.Height),
		)

		rp.Content()

		rl.EndScissorMode()
	}
	// rp.DrawContents() // Don't draw your contents here
}

func (rp *RenderPanel) DrawContents() {
	rp.MultiThreadingCheck.Draw()
	rp.FogCheck.Draw()
	rp.SoftNormalCheck.Draw()
	rp.DOFCheck.Draw()
	rp.ShadowsCheck.Draw()
}

func (rp *RenderPanel) ToggleFog(value bool) {
	rp.Layout.Scene.Renderer.FogEnabled = value

}

func (rp *RenderPanel) ToggleDOF(value bool) {
	rp.Layout.Scene.Renderer.DOFEnabled = value

}

func (rp *RenderPanel) ToggleShadow(value bool) {
	rp.Layout.Scene.DefaultLight.Shadows = value
}

func (rp *RenderPanel) ToggleSoftNormals(value bool) {
	rp.Layout.Scene.Renderer.OverrideSoftNormals = value

}

func (rp *RenderPanel) ToggleMultitheading(value bool) {
	rp.Layout.Scene.Renderer.MultiThreading = value

}
