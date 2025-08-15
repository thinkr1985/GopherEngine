package widgets

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Toggle struct {
	Bounds          rl.Rectangle
	HandleRect      rl.Rectangle
	IsOn            bool
	Label           string
	FontSize        int32
	BoxColor        rl.Color
	HandleColor     rl.Color
	TextColor       rl.Color
	TransitionSpeed float32
}

func NewToggle(xPos, yPos int, label string) *Toggle {
	toggle := Toggle{
		Bounds:          rl.NewRectangle(float32(xPos), float32(yPos), float32(50), float32(20)),
		HandleRect:      rl.NewRectangle(float32(xPos), float32(yPos), float32(20), float32(20)),
		IsOn:            false,
		Label:           label,
		FontSize:        12,
		BoxColor:        rl.Gray,
		HandleColor:     rl.White,
		TextColor:       rl.White,
		TransitionSpeed: 0.5, // Control the sliding speed
	}

	// Set the handle position based on initial state
	if toggle.IsOn {
		toggle.HandleRect.X = toggle.Bounds.X + toggle.Bounds.Width - toggle.HandleRect.Width
	}

	return &toggle
}

func (t *Toggle) Draw() {
	// Draw the box (background) of the toggle
	rl.DrawRectangleRec(t.Bounds, t.BoxColor)

	// Draw the sliding handle
	rl.DrawRectangleRec(t.HandleRect, t.HandleColor)

	// Draw the label text
	labelText := fmt.Sprintf("%s: %s", t.Label, t.StateText())
	if t.StateText() == "On" {
		t.BoxColor = rl.Green
	} else {
		t.BoxColor = rl.Gray
	}

	rl.DrawText(labelText, int32(t.Bounds.X+t.Bounds.Width+10), int32(t.Bounds.Y+5), t.FontSize, t.TextColor)
}

func (t *Toggle) StateText() string {
	if t.IsOn {
		return "On"
	}
	return "Off"
}

func (t *Toggle) Update() {
	mousePos := rl.GetMousePosition()

	// Detect if the mouse is clicked inside the toggle area
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && rl.CheckCollisionPointRec(mousePos, t.Bounds) {
		// Toggle the state
		t.IsOn = !t.IsOn
	}

	// Smoothly move the handle if the state is changing
	targetX := t.Bounds.X
	if t.IsOn {
		targetX = t.Bounds.X + t.Bounds.Width - t.HandleRect.Width
	}

	// Move the handle towards the target position
	t.HandleRect.X += (targetX - t.HandleRect.X) * t.TransitionSpeed
}
