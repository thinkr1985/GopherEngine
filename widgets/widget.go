package widgets

import rl "github.com/gen2brain/raylib-go/raylib"

// Widget interface that all widgets must implement
type Widget interface {
	Draw()
	Update()
	SetParent(parent *BaseWidget)
	GetRect() rl.Rectangle
}

// BaseWidget is a concrete implementation
type BaseWidget struct {
	Rect     rl.Rectangle
	Children []Widget
	Parent   *BaseWidget
}

// Create a new BaseWidget with given rectangle
func NewWidget(rect rl.Rectangle) *BaseWidget {
	return &BaseWidget{
		Rect:     rect,
		Children: []Widget{},
	}
}

// Append a child widget and set its parent
func (w *BaseWidget) AppendChild(child Widget) {
	child.SetParent(w)
	// Ensure child is inside parent's rect
	r := child.GetRect()
	if r.X < w.Rect.X {
		r.X = w.Rect.X
	}
	if r.Y < w.Rect.Y {
		r.Y = w.Rect.Y
	}
	if r.X+r.Width > w.Rect.X+w.Rect.Width {
		r.X = w.Rect.X + w.Rect.Width - r.Width
	}
	if r.Y+r.Height > w.Rect.Y+w.Rect.Height {
		r.Y = w.Rect.Y + w.Rect.Height - r.Height
	}

	// (Optionally you could cast and update the rect if child is *BaseWidget)
	if bw, ok := child.(*BaseWidget); ok {
		bw.Rect = r
	}

	w.Children = append(w.Children, child)
}

// Set parent widget
func (w *BaseWidget) SetParent(parent *BaseWidget) {
	w.Parent = parent
}

// Return rectangle
func (w *BaseWidget) GetRect() rl.Rectangle {
	return w.Rect
}

// Default draw: draw self and children
func (w *BaseWidget) Draw() {
	rl.DrawRectangleRec(w.Rect, rl.Gray) // default fill
	for _, child := range w.Children {
		child.Draw()
	}
}

// Default update: update children
func (w *BaseWidget) Update() {
	for _, child := range w.Children {
		child.Update()
	}
}
