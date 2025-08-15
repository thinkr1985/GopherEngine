package widgets

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ColorWheel struct {
	Bounds       rl.Rectangle
	Hue          float32
	Saturation   float32
	Value        float32
	Selected     rl.Color
	IsDragging   bool
	WheelRadius  float32
	wheelTexture rl.RenderTexture2D
	valueTexture rl.RenderTexture2D
	needsRedraw  bool
	fontSize     int32
}

func NewColorWheel(x, y, radius int32, fontSize int32) *ColorWheel {
	cw := &ColorWheel{
		Bounds:      rl.NewRectangle(float32(x), float32(y), float32(radius*2), float32(radius*2)+50), // Extra space for RGB labels
		Hue:         0,
		Saturation:  1,
		Value:       1,
		Selected:    rl.Red,
		WheelRadius: float32(radius),
		needsRedraw: true,
		fontSize:    fontSize,
	}

	// Create render textures
	cw.wheelTexture = rl.LoadRenderTexture(int32(cw.Bounds.Width), int32(cw.WheelRadius*2))
	cw.valueTexture = rl.LoadRenderTexture(30, int32(cw.WheelRadius*2))

	// Initial draw
	cw.generateWheelTexture()
	cw.generateValueTexture()

	return cw
}

func (cw *ColorWheel) GetColorAtPosition(pos rl.Vector2) rl.Color {
	center := rl.Vector2{X: cw.Bounds.X + cw.Bounds.Width/2, Y: cw.Bounds.Y + cw.WheelRadius}
	delta := rl.Vector2Subtract(pos, center)
	distance := rl.Vector2Length(delta) / cw.WheelRadius

	if distance > 1 {
		distance = 1
	}

	angle := float32(math.Atan2(float64(delta.Y), float64(delta.X)))
	if angle < 0 {
		angle += 2 * math.Pi
	}
	hue := angle / (2 * math.Pi)

	return rl.ColorFromHSV(hue*360, distance, cw.Value)
}

func (cw *ColorWheel) generateWheelTexture() {
	rl.BeginTextureMode(cw.wheelTexture)
	rl.ClearBackground(rl.Blank)

	center := rl.Vector2{X: cw.Bounds.Width / 2, Y: cw.WheelRadius}

	// Draw color wheel using colored pixels
	for y := float32(0); y < cw.WheelRadius*2; y++ {
		for x := float32(0); x < cw.Bounds.Width; x++ {
			dx := x - center.X
			dy := y - center.Y
			distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))

			if distance <= cw.WheelRadius {
				angle := float32(math.Atan2(float64(dy), float64(dx)))
				if angle < 0 {
					angle += 2 * math.Pi
				}
				hue := angle / (2 * math.Pi)
				saturation := distance / cw.WheelRadius
				rl.DrawPixel(int32(x), int32(y), rl.ColorFromHSV(hue*360, saturation, 1))
			}
		}
	}

	rl.EndTextureMode()
}

func (cw *ColorWheel) generateValueTexture() {
	rl.BeginTextureMode(cw.valueTexture)
	rl.ClearBackground(rl.Blank)

	for y := int32(0); y < int32(cw.WheelRadius*2); y++ {
		value := 1 - float32(y)/(cw.WheelRadius*2)
		col := rl.ColorFromHSV(cw.Hue*360, cw.Saturation, value)
		rl.DrawLine(0, y, 30, y, col)
	}

	rl.EndTextureMode()
}

func (cw *ColorWheel) Update() {
	mousePos := rl.GetMousePosition()
	center := rl.Vector2{X: cw.Bounds.X + cw.Bounds.Width/2, Y: cw.Bounds.Y + cw.WheelRadius}

	// Handle color wheel selection
	if rl.CheckCollisionPointCircle(mousePos, center, cw.WheelRadius) {
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			cw.IsDragging = true
		}
	}

	if rl.IsMouseButtonReleased(rl.MouseLeftButton) {
		cw.IsDragging = false
	}

	if cw.IsDragging {
		delta := rl.Vector2Subtract(mousePos, center)
		angle := float32(math.Atan2(float64(delta.Y), float64(delta.X)))
		if angle < 0 {
			angle += 2 * math.Pi
		}
		cw.Hue = angle / (2 * math.Pi)

		distance := rl.Vector2Length(delta) / cw.WheelRadius
		if distance > 1 {
			distance = 1
		}
		cw.Saturation = distance

		cw.Selected = rl.ColorFromHSV(cw.Hue*360, cw.Saturation, cw.Value)
		cw.generateValueTexture() // Update value slider when hue/saturation changes
	}

	// Handle value slider
	valueRect := rl.NewRectangle(cw.Bounds.X+cw.Bounds.Width+10, cw.Bounds.Y, 30, cw.WheelRadius*2)
	if rl.CheckCollisionPointRec(mousePos, valueRect) && rl.IsMouseButtonDown(rl.MouseLeftButton) {
		cw.Value = 1 - (mousePos.Y-valueRect.Y)/valueRect.Height
		cw.Selected = rl.ColorFromHSV(cw.Hue*360, cw.Saturation, cw.Value)
	}
}

func (cw *ColorWheel) Draw() {
	center := rl.Vector2{X: cw.Bounds.X + cw.Bounds.Width/2, Y: cw.Bounds.Y + cw.WheelRadius}

	// Draw color wheel
	rl.DrawTextureRec(
		cw.wheelTexture.Texture,
		rl.NewRectangle(0, 0, cw.Bounds.Width, -cw.WheelRadius*2),
		rl.Vector2{X: cw.Bounds.X, Y: cw.Bounds.Y},
		rl.White,
	)

	// Draw value slider
	valueRect := rl.NewRectangle(cw.Bounds.X+cw.Bounds.Width+10, cw.Bounds.Y, 30, cw.WheelRadius*2)
	rl.DrawTextureRec(
		cw.valueTexture.Texture,
		rl.NewRectangle(0, 0, 30, -cw.WheelRadius*2),
		rl.Vector2{X: valueRect.X, Y: valueRect.Y},
		rl.White,
	)

	// Draw selection indicator
	if cw.Saturation > 0 {
		selX := center.X + cw.Saturation*cw.WheelRadius*float32(math.Cos(float64(cw.Hue*2*math.Pi)))
		selY := center.Y + cw.Saturation*cw.WheelRadius*float32(math.Sin(float64(cw.Hue*2*math.Pi)))
		rl.DrawCircleLines(int32(selX), int32(selY), 10, rl.White)
		rl.DrawCircleLines(int32(selX), int32(selY), 9, rl.Black)
	}

	// Draw current value indicator
	valueY := valueRect.Y + valueRect.Height*(1-cw.Value)
	rl.DrawRectangleLinesEx(valueRect, 2, rl.White)
	rl.DrawRectangle(int32(valueRect.X), int32(valueY)-2, int32(valueRect.Width), 4, rl.Black)

	// Draw selected color preview
	previewRect := rl.NewRectangle(valueRect.X+valueRect.Width+10, valueRect.Y, 50, 50)
	rl.DrawRectangleRec(previewRect, cw.Selected)
	rl.DrawRectangleLinesEx(previewRect, 2, rl.White)

	// Draw RGB values
	rgbText := fmt.Sprintf("R: %d G: %d B: %d", cw.Selected.R, cw.Selected.G, cw.Selected.B)
	textY := cw.Bounds.Y + cw.WheelRadius*2 + 10
	rl.DrawText(rgbText, int32(cw.Bounds.X), int32(textY), cw.fontSize, rl.White)

	// Draw individual color channels with colored text
	rText := fmt.Sprintf("R: %d", cw.Selected.R)
	gText := fmt.Sprintf("G: %d", cw.Selected.G)
	bText := fmt.Sprintf("B: %d", cw.Selected.B)

	spacing := float32(10)
	totalWidth := rl.MeasureText(rText, cw.fontSize) + rl.MeasureText(gText, cw.fontSize) +
		rl.MeasureText(bText, cw.fontSize) + 2*int32(spacing)

	startX := cw.Bounds.X + cw.Bounds.Width - float32(totalWidth)/float32(2)

	// Red channel
	rl.DrawText(rText, int32(startX), int32(textY), cw.fontSize, rl.Red)
	startX += float32(rl.MeasureText(rText, cw.fontSize)) + spacing

	// Green channel
	rl.DrawText(gText, int32(startX), int32(textY), cw.fontSize, rl.Green)
	startX += float32(rl.MeasureText(gText, cw.fontSize)) + spacing

	// Blue channel
	rl.DrawText(bText, int32(startX), int32(textY), cw.fontSize, rl.Blue)
}

func (cw *ColorWheel) Unload() {
	rl.UnloadRenderTexture(cw.wheelTexture)
	rl.UnloadRenderTexture(cw.valueTexture)
}
