package widgets

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	MessageTypeInfo = iota
	MessageTypeWarning
	MessageTypeError
)

type MessageBox struct {
	Title       string
	Message     string
	Type        int
	IsVisible   bool
	Bounds      rl.Rectangle
	TitleHeight float32
	Button      *PushButton
}

func NewMessageBox(title, message string, messageType int) *MessageBox {
	// Calculate screen dimensions
	screenWidth := float32(rl.GetScreenWidth())
	screenHeight := float32(rl.GetScreenHeight())

	// Set default dimensions (60% of screen width, auto height)
	width := screenWidth * 0.6
	height := screenHeight * 0.3

	// Create bounds (centered)
	bounds := rl.NewRectangle(
		(screenWidth-width)/2,
		(screenHeight-height)/2,
		width,
		height,
	)

	// Create button (centered horizontally at bottom)
	buttonWidth := float32(80)
	buttonHeight := float32(30)
	button := NewPushButton(
		int32(bounds.X+(bounds.Width-buttonWidth)/2),
		int32(bounds.Y+bounds.Height-buttonHeight-20),
		int32(buttonWidth),
		int32(buttonHeight),
		"OK",
		20,
	)

	// Set button click handler to hide the message box
	button.SetClickHandler(func() {
		// Handler will be set when showing the message box
	})

	return &MessageBox{
		Title:       title,
		Message:     message,
		Type:        messageType,
		IsVisible:   false,
		Bounds:      bounds,
		TitleHeight: 30,
		Button:      button,
	}
}

func (mb *MessageBox) Show() {
	mb.IsVisible = true
	// Update button handler each time we show
	mb.Button.SetClickHandler(func() {
		mb.IsVisible = false
	})
}

func (mb *MessageBox) Hide() {
	mb.IsVisible = false
}

func (mb *MessageBox) Update() {
	if !mb.IsVisible {
		return
	}
	mb.Button.Update()
}

// drawWrappedText draws text with word wrapping within bounds
func (mb *MessageBox) drawWrappedText(text string, bounds rl.Rectangle, fontSize int32, color rl.Color) {
	words := strings.Fields(text)
	if len(words) == 0 {
		return
	}

	var lines []string
	currentLine := words[0]

	for _, word := range words[1:] {
		testLine := currentLine + " " + word
		if rl.MeasureText(testLine, fontSize) > int32(bounds.Width) {
			lines = append(lines, currentLine)
			currentLine = word
		} else {
			currentLine = testLine
		}
	}
	lines = append(lines, currentLine)

	y := bounds.Y
	for _, line := range lines {
		rl.DrawText(line, int32(bounds.X), int32(y), fontSize, color)
		y += float32(fontSize) + 2 // Add small line spacing
	}
}

func (mb *MessageBox) Draw() {
	if !mb.IsVisible {
		return
	}

	// Determine colors based on message type
	var titleColor rl.Color
	var borderColor rl.Color

	switch mb.Type {
	case MessageTypeWarning:
		titleColor = rl.NewColor(255, 165, 0, 255) // Orange
		borderColor = rl.NewColor(255, 165, 0, 150)
	case MessageTypeError:
		titleColor = rl.NewColor(220, 20, 60, 255) // Crimson red
		borderColor = rl.NewColor(220, 20, 60, 150)
	default: // MessageTypeInfo
		titleColor = rl.NewColor(30, 144, 255, 255) // Dodger blue
		borderColor = rl.NewColor(30, 144, 255, 150)
	}

	// Draw message box background
	rl.DrawRectangleRec(mb.Bounds, rl.LightGray)
	rl.DrawRectangleLinesEx(mb.Bounds, 2, borderColor)

	// Draw title bar
	titleBar := rl.NewRectangle(mb.Bounds.X, mb.Bounds.Y, mb.Bounds.Width, mb.TitleHeight)
	rl.DrawRectangleRec(titleBar, titleColor)
	rl.DrawRectangleLinesEx(titleBar, 1, rl.Black)

	// Draw title text (centered vertically, left-aligned with padding)
	titleFontSize := int32(20)
	titleTextX := int32(mb.Bounds.X + 10)
	titleTextY := int32(mb.Bounds.Y + (mb.TitleHeight-float32(titleFontSize))/2)
	rl.DrawText(mb.Title, titleTextX, titleTextY, titleFontSize, rl.White)

	// Draw message text (with word wrapping)
	messageFontSize := int32(18)
	textColor := rl.Black
	textPadding := float32(20)

	// Calculate text area bounds
	textBounds := rl.NewRectangle(
		mb.Bounds.X+textPadding,
		mb.Bounds.Y+mb.TitleHeight+textPadding,
		mb.Bounds.Width-textPadding*2,
		mb.Bounds.Height-mb.TitleHeight-mb.Button.Bounds.Height-textPadding*3,
	)

	// Draw wrapped text
	mb.drawWrappedText(mb.Message, textBounds, messageFontSize, textColor)

	// Draw the OK button
	mb.Button.Draw()
}

// Helper functions to create pre-configured message boxes
func ShowInfo(title, message string) *MessageBox {
	mb := NewMessageBox(title, message, MessageTypeInfo)
	mb.Show()
	return mb
}

func ShowWarning(title, message string) *MessageBox {
	mb := NewMessageBox(title, message, MessageTypeWarning)
	mb.Show()
	return mb
}

func ShowError(title, message string) *MessageBox {
	mb := NewMessageBox(title, message, MessageTypeError)
	mb.Show()
	return mb
}
