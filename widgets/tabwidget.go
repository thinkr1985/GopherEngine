package widgets

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Tab struct {
	Label    string
	Content  func() // Function to draw tab content
	Width    float32
	IsActive bool
}

type TabWidget struct {
	Bounds        rl.Rectangle
	Tabs          []*Tab
	ActiveTab     int
	TabHeight     float32
	FontSize      int32
	TextColor     rl.Color
	ActiveColor   rl.Color
	InactiveColor rl.Color
	BorderColor   rl.Color
}

func NewTabWidget(x, y, width, height int32, tabHeight int32, fontSize int32) *TabWidget {
	tab := &TabWidget{
		Bounds:        rl.NewRectangle(float32(x), float32(y), float32(width), float32(height)),
		TabHeight:     float32(tabHeight),
		FontSize:      fontSize,
		TextColor:     rl.White,
		ActiveColor:   rl.DarkGray,
		InactiveColor: rl.NewColor(40, 40, 40, 255),
		BorderColor:   rl.NewColor(70, 70, 70, 255),
	}
	InitializeWidgetFont()
	return tab
}

func (tw *TabWidget) AddTab(label string, content func()) {
	// Calculate tab width based on label length
	textWidth := float32(rl.MeasureText(label, tw.FontSize))
	tabWidth := textWidth + 20 // Add padding

	newTab := &Tab{
		Label:    label,
		Content:  content,
		Width:    tabWidth,
		IsActive: len(tw.Tabs) == 0, // First tab is active by default
	}

	tw.Tabs = append(tw.Tabs, newTab)

	// Set first tab as active if this is the first tab added
	if len(tw.Tabs) == 1 {
		tw.ActiveTab = 0
	}
}

func (tw *TabWidget) RemoveTab(index int) {
	if index < 0 || index >= len(tw.Tabs) {
		return
	}

	// Remove the tab
	tw.Tabs = append(tw.Tabs[:index], tw.Tabs[index+1:]...)

	// Adjust active tab if needed
	if tw.ActiveTab >= len(tw.Tabs) {
		tw.ActiveTab = len(tw.Tabs) - 1
	}
	if tw.ActiveTab >= 0 && len(tw.Tabs) > 0 {
		tw.Tabs[tw.ActiveTab].IsActive = true
	}
}

func (tw *TabWidget) SetActiveTab(index int) {
	if index < 0 || index >= len(tw.Tabs) {
		return
	}

	for i := range tw.Tabs {
		tw.Tabs[i].IsActive = (i == index)
	}
	tw.ActiveTab = index
}

func (tw *TabWidget) Update() {
	mousePos := rl.GetMousePosition()

	// Check if mouse is in the tab bar area
	tabBarRect := rl.NewRectangle(
		tw.Bounds.X,
		tw.Bounds.Y,
		tw.Bounds.Width,
		tw.TabHeight,
	)

	if rl.CheckCollisionPointRec(mousePos, tabBarRect) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		// Calculate which tab was clicked
		xPos := tw.Bounds.X
		for i, tab := range tw.Tabs {
			tabRect := rl.NewRectangle(
				xPos,
				tw.Bounds.Y,
				tab.Width,
				tw.TabHeight,
			)

			if rl.CheckCollisionPointRec(mousePos, tabRect) {
				tw.SetActiveTab(i)
				break
			}

			xPos += tab.Width
		}
	}
}

func (tw *TabWidget) Draw() {
	// Draw tab bar background
	rl.DrawRectangleRec(
		rl.NewRectangle(tw.Bounds.X, tw.Bounds.Y, tw.Bounds.Width, tw.TabHeight),
		tw.InactiveColor,
	)

	// Draw tabs
	xPos := tw.Bounds.X
	for _, tab := range tw.Tabs {
		tabRect := rl.NewRectangle(
			xPos,
			tw.Bounds.Y,
			tab.Width,
			tw.TabHeight,
		)

		// Draw tab background
		tabColor := tw.InactiveColor
		if tab.IsActive {
			tabColor = tw.ActiveColor
		}
		rl.DrawRectangleRec(tabRect, tabColor)
		rl.DrawRectangleLinesEx(tabRect, 1, tw.BorderColor)

		// Draw tab label (left-aligned with padding)
		textX := float32(xPos + 5)
		textY := tw.Bounds.Y + (tw.TabHeight-float32(tw.FontSize))/2
		rl.DrawTextEx(Widget_body_font, tab.Label, rl.NewVector2(textX, textY), 14, 0, tw.TextColor)

		// For closeable tabs, draw close button
		if strings.HasPrefix(tab.Label, "× ") { // Example check for closeable tabs
			// Draw close button (small rectangle with X)
			closeBtnSize := float32(tw.FontSize) * 0.8
			closeBtnPadding := float32(5)
			closeBtnX := xPos + tab.Width - closeBtnSize - closeBtnPadding
			closeBtnY := tw.Bounds.Y + (tw.TabHeight-closeBtnSize)/2

			// Draw close button background (light color)
			closeBtnColor := rl.Red // Semi-transparent light gray
			if rl.CheckCollisionPointRec(rl.GetMousePosition(),
				rl.NewRectangle(closeBtnX, closeBtnY, closeBtnSize, closeBtnSize)) {
				closeBtnColor = rl.Green
			}
			rl.DrawRectangleRounded(
				rl.NewRectangle(closeBtnX, closeBtnY, closeBtnSize, closeBtnSize),
				0.3, // Roundness
				4,   // Segments
				closeBtnColor,
			)

			// Draw X symbol
			xColor := rl.White
			xPadding := closeBtnSize * 0.25
			rl.DrawLineEx(
				rl.Vector2{X: closeBtnX + xPadding, Y: closeBtnY + xPadding},
				rl.Vector2{X: closeBtnX + closeBtnSize - xPadding, Y: closeBtnY + closeBtnSize - xPadding},
				1.5,
				xColor,
			)
			rl.DrawLineEx(
				rl.Vector2{X: closeBtnX + closeBtnSize - xPadding, Y: closeBtnY + xPadding},
				rl.Vector2{X: closeBtnX + xPadding, Y: closeBtnY + closeBtnSize - xPadding},
				1.5,
				xColor,
			)
		}

		xPos += tab.Width
	}

	// Draw content area
	contentRect := rl.NewRectangle(
		tw.Bounds.X,
		tw.Bounds.Y+tw.TabHeight,
		tw.Bounds.Width,
		tw.Bounds.Height-tw.TabHeight,
	)
	rl.DrawRectangleRec(contentRect, rl.NewColor(40, 40, 40, 255))
	rl.DrawRectangleLinesEx(contentRect, 1, tw.BorderColor)

	// Draw active tab content
	if tw.ActiveTab >= 0 && tw.ActiveTab < len(tw.Tabs) && tw.Tabs[tw.ActiveTab].Content != nil {
		// Set scissor to content area
		rl.BeginScissorMode(
			int32(contentRect.X),
			int32(contentRect.Y),
			int32(contentRect.Width),
			int32(contentRect.Height),
		)

		// Draw tab content
		tw.Tabs[tw.ActiveTab].Content()

		// End scissor mode
		rl.EndScissorMode()
	}
}

// AddCloseableTab creates a tab with a close button
func (tw *TabWidget) AddCloseableTab(label string, content func()) {
	closeBtnSize := float32(tw.FontSize)
	textWidth := float32(rl.MeasureText(label, tw.FontSize))
	tabWidth := textWidth + 20 + closeBtnSize + 5

	newTab := &Tab{
		Label:    label,
		Content:  content,
		Width:    tabWidth,
		IsActive: len(tw.Tabs) == 0,
	}

	tw.Tabs = append(tw.Tabs, newTab)

	if len(tw.Tabs) == 1 {
		tw.ActiveTab = 0
	}
}

// Update the DrawCloseableTab method to show a visible close button
func (tw *TabWidget) DrawCloseableTab(tab *Tab, tabRect rl.Rectangle) {
	// Draw tab background
	tabColor := tw.InactiveColor
	if tab.IsActive {
		tabColor = tw.ActiveColor
	}
	rl.DrawRectangleRec(tabRect, tabColor)
	rl.DrawRectangleLinesEx(tabRect, 1, tw.BorderColor)

	// Draw tab label (left-aligned with padding)
	textX := tabRect.X + 5
	textY := tabRect.Y + (tabRect.Height-float32(tw.FontSize))/2
	rl.DrawTextEx(Widget_body_font, tab.Label, rl.NewVector2(float32(textX), float32(textY)), 12, 0, tw.TextColor)

	// Draw close button (small rectangle with X)
	closeBtnSize := float32(tw.FontSize) * 0.8
	closeBtnPadding := float32(5)
	closeBtnX := tabRect.X + tabRect.Width - closeBtnSize - closeBtnPadding
	closeBtnY := tabRect.Y + (tabRect.Height-closeBtnSize)/2

	// Draw close button background (light color)
	closeBtnColor := rl.Red
	if rl.CheckCollisionPointRec(rl.GetMousePosition(),
		rl.NewRectangle(closeBtnX, closeBtnY, closeBtnSize, closeBtnSize)) {
		closeBtnColor = rl.Green
	}
	rl.DrawRectangleRounded(
		rl.NewRectangle(closeBtnX, closeBtnY, closeBtnSize, closeBtnSize),
		0.3, // Roundness
		4,   // Segments
		closeBtnColor,
	)

	// Draw X symbol
	xColor := rl.Orange
	xPadding := closeBtnSize * 0.25
	rl.DrawLineEx(
		rl.Vector2{X: closeBtnX + xPadding, Y: closeBtnY + xPadding},
		rl.Vector2{X: closeBtnX + closeBtnSize - xPadding, Y: closeBtnY + closeBtnSize - xPadding},
		1.5,
		xColor,
	)
	rl.DrawLineEx(
		rl.Vector2{X: closeBtnX + closeBtnSize - xPadding, Y: closeBtnY + xPadding},
		rl.Vector2{X: closeBtnX + xPadding, Y: closeBtnY + closeBtnSize - xPadding},
		1.5,
		xColor,
	)
}

// Update the UpdateCloseableTabs method to use the same button dimensions
func (tw *TabWidget) UpdateCloseableTabs() {
	mousePos := rl.GetMousePosition()
	tabBarRect := rl.NewRectangle(tw.Bounds.X, tw.Bounds.Y, tw.Bounds.Width, tw.TabHeight)

	if rl.CheckCollisionPointRec(mousePos, tabBarRect) && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		xPos := tw.Bounds.X
		for i, tab := range tw.Tabs {
			tabRect := rl.NewRectangle(xPos, tw.Bounds.Y, tab.Width, tw.TabHeight)

			// Calculate close button area (must match DrawCloseableTab)
			closeBtnSize := float32(tw.FontSize) * 0.8
			closeBtnPadding := float32(5)
			closeBtnRect := rl.NewRectangle(
				xPos+tab.Width-closeBtnSize-closeBtnPadding,
				tw.Bounds.Y+(tw.TabHeight-closeBtnSize)/2,
				closeBtnSize,
				closeBtnSize,
			)

			if rl.CheckCollisionPointRec(mousePos, closeBtnRect) {
				tw.RemoveTab(i)
				break
			} else if rl.CheckCollisionPointRec(mousePos, tabRect) {
				tw.SetActiveTab(i)

			}

			xPos += tab.Width
		}
	}
}
