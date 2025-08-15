package widgets

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type CheckBox struct {
	Bounds    rl.Rectangle
	IsChecked bool
	Label     string
	FontSize  int32
	BoxColor  rl.Color
	TextColor rl.Color
}

func NewCheckBox(xPos, yPos, width, height int, label string) *CheckBox {
	InitializeWidgetFont()
	// defer rl.UnloadFont(widget_default_font)

	return &CheckBox{
		Bounds:    rl.NewRectangle(float32(xPos), float32(yPos), float32(width), float32(height)),
		IsChecked: false,
		Label:     label,
		FontSize:  12,
		BoxColor:  rl.DarkGray,
		TextColor: rl.White,
	}
}

func (cb *CheckBox) Draw() {
	// Draw checkbox box
	if cb.IsChecked {
		rl.DrawRectangleRec(cb.Bounds, rl.Green) // Green if checked
	} else {
		rl.DrawRectangleRec(cb.Bounds, rl.Gray) // Gray if unchecked
	}

	// Draw checkbox label text
	labelText := fmt.Sprintf("%s: %s", cb.Label, cb.StateText())
	rl.DrawTextEx(
		Widget_default_font, // Font (using default font)
		labelText,           // Text to draw
		rl.Vector2{ // Position (convert from X,Y to Vector2)
			X: cb.Bounds.X + cb.Bounds.Width + 10,
			Y: cb.Bounds.Y,
		},
		float32(cb.FontSize), // Font size (convert int32 to float32)
		1.0,                  // Spacing (default 1.0)
		cb.TextColor,         // Text color
	)
	// 	rl.DrawText(labelText, int32(cb.Bounds.X+cb.Bounds.Width+10), int32(cb.Bounds.Y), cb.FontSize, cb.TextColor)
}

func (cb *CheckBox) StateText() string {
	if cb.IsChecked {
		return "On"
	}
	return "Off"
}

func (cb *CheckBox) Update() {
	mousePos := rl.GetMousePosition()

	// Check if the mouse is pressed and within the bounds of the checkbox
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && rl.CheckCollisionPointRec(mousePos, cb.Bounds) {
		cb.IsChecked = !cb.IsChecked
	}
}
