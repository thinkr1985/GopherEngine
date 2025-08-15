package widgets

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Icon struct {
	Texture   rl.Texture2D
	Position  rl.Vector2
	Size      rl.Vector2
	Color     rl.Color
	Scale     float32
	Rotation  float32
	IsPressed bool
	IsHovered bool
	OnClick   func()
}

func NewIcon(texture rl.Texture2D, x, y float32) *Icon {
	return &Icon{
		Texture:  texture,
		Position: rl.Vector2{X: x, Y: y},
		Size:     rl.Vector2{X: float32(texture.Width), Y: float32(texture.Height)},
		Color:    rl.White,
		Scale:    1.0,
		Rotation: 0,
	}
}

func (icon *Icon) SetSize(width, height float32) {
	icon.Size = rl.Vector2{X: width, Y: height}
}

func (icon *Icon) SetScale(scale float32) {
	icon.Scale = scale
}

func (icon *Icon) SetColor(color rl.Color) {
	icon.Color = color
}

func (icon *Icon) SetRotation(degrees float32) {
	icon.Rotation = degrees
}

func (icon *Icon) Update() {
	mousePos := rl.GetMousePosition()

	// Calculate scaled size
	scaledWidth := float32(icon.Texture.Width) * icon.Scale
	scaledHeight := float32(icon.Texture.Height) * icon.Scale

	iconBounds := rl.Rectangle{
		X:      icon.Position.X,
		Y:      icon.Position.Y,
		Width:  scaledWidth,
		Height: scaledHeight,
	}

	icon.IsHovered = rl.CheckCollisionPointRec(mousePos, iconBounds)

	if icon.IsHovered {
		if rl.IsMouseButtonDown(rl.MouseLeftButton) {
			icon.IsPressed = true
		} else {
			if icon.IsPressed && rl.IsMouseButtonReleased(rl.MouseLeftButton) {
				if icon.OnClick != nil {
					icon.OnClick()
				}
			}
			icon.IsPressed = false
		}
	} else {
		icon.IsPressed = false
	}
}

func (icon *Icon) Draw() {
	// Draw the icon with all transformations
	rl.DrawTextureEx(
		icon.Texture,
		icon.Position,
		icon.Rotation,
		icon.Scale,
		icon.Color,
	)

	// Optional: Draw hover/press effects
	if icon.IsHovered {
		rl.DrawRectangleLinesEx(
			rl.Rectangle{
				X:      icon.Position.X,
				Y:      icon.Position.Y,
				Width:  float32(icon.Texture.Width) * icon.Scale,
				Height: float32(icon.Texture.Height) * icon.Scale,
			},
			1,
			rl.Fade(rl.White, 0.5),
		)
	}
}

func (icon *Icon) SetClickHandler(handler func()) {
	icon.OnClick = handler
}

func (icon *Icon) Unload() {
	rl.UnloadTexture(icon.Texture)
}
