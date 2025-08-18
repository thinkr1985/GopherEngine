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
		Title:               "Render Settings",
		IsVisible:           true,
		BgColor:             rl.NewColor(30, 30, 30, 255),
		BorderColor:         rl.NewColor(60, 60, 60, 255),
		SoftNormalCheck:     *widgets.NewCheckBox(10, 10, 20, 20, "Soft Normals"),
		FogCheck:            *widgets.NewCheckBox(10, 50, 20, 20, "Fog"),
		DOFCheck:            *widgets.NewCheckBox(10, 90, 20, 20, "Depth Of Field"),
		ShadowsCheck:        *widgets.NewCheckBox(10, 130, 20, 20, "Shadows"),
		MultiThreadingCheck: *widgets.NewCheckBox(10, 170, 20, 20, "MultiThreading"),
	}

	// Set up toggle handlers
	rp.SoftNormalCheck.OnToggle = rp.ToggleSoftNormals
	rp.FogCheck.OnToggle = rp.ToggleFog
	rp.DOFCheck.OnToggle = rp.ToggleDOF
	rp.ShadowsCheck.OnToggle = rp.ToggleShadow
	rp.MultiThreadingCheck.OnToggle = rp.ToggleMultitheading

	// Initialize checkboxes to match current scene state
	if rp.Layout != nil && rp.Layout.Scene != nil {
		rp.SoftNormalCheck.IsChecked = !rp.Layout.Scene.Renderer.OverrideSoftNormals
		rp.FogCheck.IsChecked = rp.Layout.Scene.Renderer.FogEnabled
		rp.DOFCheck.IsChecked = rp.Layout.Scene.Renderer.DOFEnabled
		rp.ShadowsCheck.IsChecked = rp.Layout.Scene.DefaultLight.Shadows
		rp.MultiThreadingCheck.IsChecked = rp.Layout.Scene.Renderer.MultiThreading
	}

	return rp
}

func (rp *RenderPanel) Update(targetWidth, targetHeight float32) {
	// Update checkbox positions relative to their container
	rp.FogCheck.Bounds = rl.NewRectangle(20, 50, 20, 20)
	rp.SoftNormalCheck.Bounds = rl.NewRectangle(20, 90, 20, 20)
	rp.DOFCheck.Bounds = rl.NewRectangle(20, 130, 20, 20)
	rp.ShadowsCheck.Bounds = rl.NewRectangle(20, 170, 20, 20)
	rp.MultiThreadingCheck.Bounds = rl.NewRectangle(20, 210, 20, 20)

	// Update all checkboxes
	rp.FogCheck.Update()
	rp.SoftNormalCheck.Update()
	rp.DOFCheck.Update()
	rp.ShadowsCheck.Update()
	rp.MultiThreadingCheck.Update()
}

func (rp *RenderPanel) DrawContents() {
	// Draw title
	rl.DrawTextEx(
		widgets.Widget_default_font,
		"Render Settings",
		rl.NewVector2(20, 10),
		18,
		1,
		rl.White,
	)

	// Draw all checkboxes
	rp.FogCheck.Draw()
	rp.SoftNormalCheck.Draw()
	rp.DOFCheck.Draw()
	rp.ShadowsCheck.Draw()
	rp.MultiThreadingCheck.Draw()
}

func (rp *RenderPanel) ToggleFog(value bool) {
	if rp.Layout != nil && rp.Layout.Scene != nil {
		rp.Layout.Scene.Renderer.FogEnabled = value
	}
}

func (rp *RenderPanel) ToggleDOF(value bool) {
	if rp.Layout != nil && rp.Layout.Scene != nil {
		rp.Layout.Scene.Renderer.DOFEnabled = value
	}
}

func (rp *RenderPanel) ToggleShadow(value bool) {
	if rp.Layout != nil && rp.Layout.Scene != nil {
		rp.Layout.Scene.DefaultLight.Shadows = value
	}
}

func (rp *RenderPanel) ToggleSoftNormals(value bool) {
	if rp.Layout != nil && rp.Layout.Scene != nil {
		rp.Layout.Scene.Renderer.OverrideSoftNormals = !value
	}
}

func (rp *RenderPanel) ToggleMultitheading(value bool) {
	if rp.Layout != nil && rp.Layout.Scene != nil {
		rp.Layout.Scene.Renderer.MultiThreading = value
	}
}
