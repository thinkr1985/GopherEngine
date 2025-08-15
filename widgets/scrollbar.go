package widgets

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type ScrollBar struct {
	Bounds      rl.Rectangle
	Thumb       rl.Rectangle
	IsDragging  bool
	Value       float32
	Min, Max    float32
	ContentSize float32
	VisibleSize float32
}

func NewScrollBar(x, y, height int32) *ScrollBar {
	width := float32(10)
	return &ScrollBar{
		Bounds: rl.NewRectangle(float32(x), float32(y), width, float32(height)),
		Thumb:  rl.NewRectangle(float32(x), float32(y), width, 40),
		Min:    0,
		Max:    1,
	}
}

func (sb *ScrollBar) Update() {
	mousePos := rl.GetMousePosition()

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		if rl.CheckCollisionPointRec(mousePos, sb.Thumb) {
			sb.IsDragging = true
		} else if rl.CheckCollisionPointRec(mousePos, sb.Bounds) {
			// Clicked on scrollbar but not thumb - jump to position
			relativeY := mousePos.Y - sb.Bounds.Y
			sb.Value = relativeY / sb.Bounds.Height
			sb.Value = rl.Clamp(sb.Value, 0, 1)
			sb.updateThumbPosition()
		}
	}

	if sb.IsDragging {
		if rl.IsMouseButtonDown(rl.MouseLeftButton) {
			relativeY := mousePos.Y - sb.Bounds.Y
			sb.Value = relativeY / sb.Bounds.Height
			sb.Value = rl.Clamp(sb.Value, 0, 1)
			sb.updateThumbPosition()
		} else {
			sb.IsDragging = false
		}
	}
}

func (sb *ScrollBar) updateThumbPosition() {
	// Calculate thumb height based on content ratio
	ratio := sb.VisibleSize / sb.ContentSize
	if ratio > 1 {
		ratio = 1
	}
	thumbHeight := sb.Bounds.Height * ratio
	if thumbHeight < 20 {
		thumbHeight = 20
	}
	sb.Thumb.Height = thumbHeight

	// Position thumb based on current value
	sb.Thumb.Y = sb.Bounds.Y + (sb.Bounds.Height-thumbHeight)*sb.Value
	sb.Thumb.X = sb.Bounds.X
	sb.Thumb.Width = sb.Bounds.Width
}

func (sb *ScrollBar) SetContentSize(contentSize, visibleSize float32) {
	sb.ContentSize = contentSize
	sb.VisibleSize = visibleSize
	sb.updateThumbPosition()
}

func (sb *ScrollBar) GetValue() float32 {
	return sb.Value
}

func (sb *ScrollBar) Draw() {
	// Draw scrollbar track
	rl.DrawRectangleRec(sb.Bounds, rl.NewColor(50, 50, 50, 200))

	// Draw thumb
	thumbColor := rl.NewColor(100, 100, 100, 200)
	if sb.IsDragging {
		thumbColor = rl.NewColor(150, 150, 150, 200)
	} else if rl.CheckCollisionPointRec(rl.GetMousePosition(), sb.Thumb) {
		thumbColor = rl.NewColor(120, 120, 120, 200)
	}
	rl.DrawRectangleRec(sb.Thumb, thumbColor)
}
