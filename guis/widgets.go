package guis

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Slider struct {
	Bounds   rl.Rectangle
	Value    float32
	Min, Max float32
	Dragging bool
	Label    string
	FontSize int32
}

func NewSlider(xPos, ypos int, width, height int, min, max, value float32, label string) *Slider {
	slider := Slider{
		Bounds: rl.NewRectangle(
			float32(xPos),
			float32(ypos),
			float32(width),
			float32(height),
		),
		Value:    value,
		Min:      min,
		Max:      max,
		Dragging: false,
		Label:    label,
		FontSize: 12,
	}
	return &slider
}

func (s *Slider) PlaceSliderOnScreen(xPos, ypos int) {
	new_bounds := rl.NewRectangle(
		float32(xPos),
		float32(ypos),
		float32(s.Bounds.Width),
		float32(s.Bounds.Height),
	)
	s.Bounds = new_bounds
	s.Draw()
}

func (s *Slider) Draw() {
	fmt.Println("*****************************************************")
	// Draw label
	rl.DrawText(s.Label, int32(s.Bounds.X), int32(s.Bounds.Y-25), s.FontSize, rl.White)

	// Draw slider track
	rl.DrawRectangleRec(s.Bounds, rl.Gray)

	// Draw slider handle
	handlePos := s.Bounds.X + (s.Value-s.Min)/(s.Max-s.Min)*s.Bounds.Width
	handleRect := rl.NewRectangle(
		handlePos-5,
		s.Bounds.Y,
		10,
		s.Bounds.Height,
	)
	rl.DrawRectangleRec(handleRect, rl.Red)

	// Draw value
	valueText := fmt.Sprintf("%.2f", s.Value)
	rl.DrawText(valueText, int32(s.Bounds.X+s.Bounds.Width+10), int32(s.Bounds.Y), s.FontSize, rl.White)
}

func (s *Slider) Update() {
	fmt.Println("*********************************************")
	mousePos := rl.GetMousePosition()

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		if rl.CheckCollisionPointRec(mousePos, s.Bounds) {
			s.Dragging = true
		}
	}

	if rl.IsMouseButtonReleased(rl.MouseLeftButton) {
		s.Dragging = false
	}

	if s.Dragging {
		normalized := (mousePos.X - s.Bounds.X) / s.Bounds.Width
		s.Value = s.Min + normalized*(s.Max-s.Min)
		s.Value = float32(math.Max(float64(s.Min), math.Min(float64(s.Max), float64(s.Value))))
	}
}
