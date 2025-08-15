package widgets

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type DropdownOption struct {
	Label string
	Value interface{}
}

type Dropdown struct {
	Bounds       rl.Rectangle
	Options      []DropdownOption
	Selected     int
	IsOpen       bool
	FontSize     int32
	TextColor    rl.Color
	BgColor      rl.Color
	HoverColor   rl.Color
	BorderColor  rl.Color
	BorderWidth  float32
	OptionHeight float32
}

func NewDropdown(x, y, width, height int32, fontSize int32) *Dropdown {
	return &Dropdown{
		Bounds:       rl.NewRectangle(float32(x), float32(y), float32(width), float32(height)),
		Selected:     -1, // No selection by default
		FontSize:     fontSize,
		TextColor:    rl.White,
		BgColor:      rl.DarkGray,
		HoverColor:   rl.Gray,
		BorderColor:  rl.LightGray,
		BorderWidth:  1,
		OptionHeight: float32(fontSize) + 10,
	}
}

func (dd *Dropdown) AddOption(label string, value interface{}) {
	dd.Options = append(dd.Options, DropdownOption{Label: label, Value: value})
	if dd.Selected == -1 && len(dd.Options) > 0 {
		dd.Selected = 0 // Select first option by default
	}
}
func (dd *Dropdown) Update() {
	mousePos := rl.GetMousePosition()

	// Check if mouse is pressed anywhere first
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		// Check if click was on main dropdown box
		if rl.CheckCollisionPointRec(mousePos, dd.Bounds) {
			dd.IsOpen = !dd.IsOpen
		} else if dd.IsOpen {
			// Check if click was on any option
			optionsHeight := float32(len(dd.Options)) * dd.OptionHeight
			optionsRect := rl.NewRectangle(dd.Bounds.X, dd.Bounds.Y+dd.Bounds.Height, dd.Bounds.Width, optionsHeight)

			if rl.CheckCollisionPointRec(mousePos, optionsRect) {
				relY := mousePos.Y - optionsRect.Y
				optionIndex := int(relY / dd.OptionHeight)

				if optionIndex >= 0 && optionIndex < len(dd.Options) {
					dd.Selected = optionIndex
				}
			}
			// Close dropdown after any click outside main box
			dd.IsOpen = false
		}
	}
}

func (dd *Dropdown) Draw() {
	// Draw main dropdown box
	color := dd.BgColor
	if rl.CheckCollisionPointRec(rl.GetMousePosition(), dd.Bounds) {
		color = dd.HoverColor
	}
	rl.DrawRectangleRec(dd.Bounds, color)
	rl.DrawRectangleLinesEx(dd.Bounds, dd.BorderWidth, dd.BorderColor)

	// Draw selected option text
	if dd.Selected >= 0 && dd.Selected < len(dd.Options) {
		text := dd.Options[dd.Selected].Label
		// textWidth := rl.MeasureText(text, dd.FontSize)
		textX := dd.Bounds.X + 5
		textY := dd.Bounds.Y + (dd.Bounds.Height-float32(dd.FontSize))/2
		rl.DrawText(text, int32(textX), int32(textY), dd.FontSize, dd.TextColor)
	}

	// Draw arrow indicator
	arrowSize := float32(dd.FontSize) / 3
	arrowX := dd.Bounds.X + dd.Bounds.Width - arrowSize - 5
	arrowY := dd.Bounds.Y + (dd.Bounds.Height-arrowSize)/2

	if dd.IsOpen {
		// Up arrow
		rl.DrawTriangle(
			rl.Vector2{X: arrowX, Y: arrowY + arrowSize},
			rl.Vector2{X: arrowX + arrowSize, Y: arrowY + arrowSize},
			rl.Vector2{X: arrowX + arrowSize/2, Y: arrowY},
			dd.TextColor,
		)
	} else {
		// Down arrow
		rl.DrawTriangle(
			rl.Vector2{X: arrowX, Y: arrowY},
			rl.Vector2{X: arrowX + arrowSize, Y: arrowY},
			rl.Vector2{X: arrowX + arrowSize/2, Y: arrowY + arrowSize},
			dd.TextColor,
		)
	}

	// Draw options if open
	if dd.IsOpen && len(dd.Options) > 0 {
		optionsHeight := float32(len(dd.Options)) * dd.OptionHeight
		optionsRect := rl.NewRectangle(dd.Bounds.X, dd.Bounds.Y+dd.Bounds.Height, dd.Bounds.Width, optionsHeight)

		// Draw options background
		rl.DrawRectangleRec(optionsRect, dd.BgColor)
		rl.DrawRectangleLinesEx(optionsRect, dd.BorderWidth, dd.BorderColor)

		// Draw each option
		for i, option := range dd.Options {
			optionRect := rl.NewRectangle(
				optionsRect.X,
				optionsRect.Y+float32(i)*dd.OptionHeight,
				optionsRect.Width,
				dd.OptionHeight,
			)

			// Highlight hovered option
			if rl.CheckCollisionPointRec(rl.GetMousePosition(), optionRect) {
				rl.DrawRectangleRec(optionRect, dd.HoverColor)
			}

			// Draw option text
			textY := optionRect.Y + (optionRect.Height-float32(dd.FontSize))/2
			rl.DrawText(option.Label, int32(optionRect.X+5), int32(textY), dd.FontSize, dd.TextColor)
		}
	}
}

func (dd *Dropdown) GetSelectedValue() interface{} {
	if dd.Selected >= 0 && dd.Selected < len(dd.Options) {
		return dd.Options[dd.Selected].Value
	}
	return nil
}

func (dd *Dropdown) SetSelectedByValue(value interface{}) {
	for i, option := range dd.Options {
		if option.Value == value {
			dd.Selected = i
			return
		}
	}
	dd.Selected = -1
}
