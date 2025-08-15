package widgets

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type FloatSlider struct {
	Bounds       rl.Rectangle
	HandleRect   rl.Rectangle
	Value        float32
	Min, Max     float32
	Dragging     bool
	Label        string
	FontSize     int32
	Transparency uint8
	SliderColor  rl.Color
	HandleColor  rl.Color
	TextColor    rl.Color
}

func NewFloatSlider(xPos, ypos int, width, height int, min, max, value float32, label string) *FloatSlider {
	slider := FloatSlider{
		Bounds: rl.NewRectangle(
			float32(xPos)+50,
			float32(ypos),
			float32(width),
			float32(height),
		),
		Value:        value,
		Min:          min,
		Max:          max,
		Dragging:     false,
		Label:        label,
		FontSize:     12,
		Transparency: uint8(200),
		SliderColor:  rl.Gray,
		HandleColor:  rl.DarkGray,
		TextColor:    rl.White,
	}

	handlePos := slider.Bounds.X + (slider.Value-slider.Min)/(slider.Max-slider.Min)*slider.Bounds.Width
	slider.HandleRect = rl.NewRectangle(
		handlePos,
		slider.Bounds.Y,
		10,
		slider.Bounds.Height,
	)
	slider.SliderColor.A = slider.Transparency
	slider.HandleColor.A = slider.Transparency
	slider.TextColor.A = slider.Transparency
	return &slider
}

func (s *FloatSlider) PlaceSliderOnScreen(xPos, yPos int) {
	s.Bounds.X = float32(xPos)
	s.Bounds.Y = float32(yPos)

	s.Draw()
}

func (s *FloatSlider) Draw() {
	// Draw label
	rl.DrawText(s.Label, int32(s.Bounds.X-50), int32(s.Bounds.Y+3), s.FontSize, s.TextColor)

	// Draw slider track
	rl.DrawRectangleRec(s.Bounds, s.SliderColor)
	handlePos := s.Bounds.X + (s.Value-s.Min-5)/(s.Max-s.Min+15)*s.Bounds.Width
	s.HandleRect.X = handlePos
	rl.DrawRectangleRec(s.HandleRect, s.HandleColor)

	// Draw value
	valueText := fmt.Sprintf("%.2f", s.Value)
	rl.DrawText(valueText, int32(s.Bounds.X+s.Bounds.Width+10), int32(s.Bounds.Y+3), s.FontSize, s.TextColor)
}

func (s *FloatSlider) Update() {
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
