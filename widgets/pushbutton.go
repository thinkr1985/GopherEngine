package widgets

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	IconAlignLeft = iota
	IconAlignRight
)

type PushButton struct {
	Bounds      rl.Rectangle
	Text        string
	FontSize    int32
	TextColor   rl.Color
	BgColor     rl.Color
	HoverColor  rl.Color
	PressColor  rl.Color
	BorderColor rl.Color
	BorderWidth float32
	IsPressed   bool
	IsHovered   bool
	OnClick     func()
	Icon        *Icon
	IconAlign   int
	IconPadding float32
}

func NewPushButton(x, y, width, height int32, text string, fontSize int32) *PushButton {
	return &PushButton{
		Bounds:      rl.NewRectangle(float32(x), float32(y), float32(width), float32(height)),
		Text:        text,
		FontSize:    fontSize,
		TextColor:   rl.White,
		BgColor:     rl.NewColor(70, 70, 70, 255),
		HoverColor:  rl.NewColor(100, 100, 100, 255),
		PressColor:  rl.NewColor(50, 50, 50, 255),
		BorderColor: rl.LightGray,
		BorderWidth: 1,
		IconAlign:   IconAlignLeft,
		IconPadding: 5,
	}
}

func (btn *PushButton) SetIcon(icon *Icon, align int, padding float32) {
	btn.Icon = icon
	btn.IconAlign = align
	btn.IconPadding = padding

	// Position the icon appropriately
	btn.updateIconPosition()
}

func (btn *PushButton) updateIconPosition() {
	if btn.Icon == nil {
		return
	}

	iconWidth := float32(btn.Icon.Texture.Width) * btn.Icon.Scale
	iconHeight := float32(btn.Icon.Texture.Height) * btn.Icon.Scale

	// Calculate Y position (centered vertically)
	iconY := btn.Bounds.Y + (btn.Bounds.Height-iconHeight)/2

	if btn.IconAlign == IconAlignLeft {
		btn.Icon.Position = rl.Vector2{
			X: btn.Bounds.X + btn.IconPadding,
			Y: iconY,
		}
	} else {
		btn.Icon.Position = rl.Vector2{
			X: btn.Bounds.X + btn.Bounds.Width - iconWidth - btn.IconPadding,
			Y: iconY,
		}
	}
}

func (btn *PushButton) Update() {
	mousePos := rl.GetMousePosition()
	btn.IsHovered = rl.CheckCollisionPointRec(mousePos, btn.Bounds)

	// Update icon state if it exists
	if btn.Icon != nil {
		btn.Icon.Update()

		// If icon is clicked, treat it as button click
		if btn.Icon.IsPressed && rl.IsMouseButtonReleased(rl.MouseLeftButton) {
			if btn.OnClick != nil {
				btn.OnClick()
			}
		}
	}

	if btn.IsHovered {
		if rl.IsMouseButtonDown(rl.MouseLeftButton) {
			btn.IsPressed = true
		} else {
			if btn.IsPressed && rl.IsMouseButtonReleased(rl.MouseLeftButton) {
				if btn.OnClick != nil {
					btn.OnClick()
				}
			}
			btn.IsPressed = false
		}
	} else {
		btn.IsPressed = false
	}
}

func (btn *PushButton) Draw() {
	// Determine button color based on state
	var btnColor rl.Color
	if btn.IsPressed {
		btnColor = btn.PressColor
	} else if btn.IsHovered {
		btnColor = btn.HoverColor
	} else {
		btnColor = btn.BgColor
	}

	// Draw button background
	rl.DrawRectangleRec(btn.Bounds, btnColor)
	rl.DrawRectangleLinesEx(btn.Bounds, btn.BorderWidth, btn.BorderColor)

	// Draw icon if it exists
	if btn.Icon != nil {
		btn.Icon.Draw()
	}

	// Calculate text position
	var textX float32
	textWidth := float32(rl.MeasureText(btn.Text, btn.FontSize))
	textY := btn.Bounds.Y + (btn.Bounds.Height-float32(btn.FontSize))/2

	if btn.Icon == nil {
		// Center text if no icon
		textX = btn.Bounds.X + (btn.Bounds.Width-textWidth)/2
	} else {
		iconWidth := float32(btn.Icon.Texture.Width) * btn.Icon.Scale
		if btn.IconAlign == IconAlignLeft {
			// Text after icon
			textX = btn.Icon.Position.X + iconWidth + btn.IconPadding
		} else {
			// Text before icon
			textX = btn.Bounds.X + btn.IconPadding
		}
	}

	// Draw button text
	if btn.Text != "" {
		rl.DrawText(btn.Text, int32(textX), int32(textY), btn.FontSize, btn.TextColor)
	}
}

func (btn *PushButton) SetClickHandler(handler func()) {
	btn.OnClick = handler
	if btn.Icon != nil {
		btn.Icon.OnClick = handler
	}
}
