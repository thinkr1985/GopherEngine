package lookdev

import (
	"errors"
	"fmt"
	"math"
)

type ColorRGB struct {
	R uint8 // Red (0-255)
	G uint8 // Green (0-255)
	B uint8 // Blue (0-255)
}

// ColorRGBA represents a color with 8-bit channels (0-255) and alpha (0-1)
type ColorRGBA struct {
	R uint8   // Red (0-255)
	G uint8   // Green (0-255)
	B uint8   // Blue (0-255)
	A float64 // Alpha (0.0-1.0)
}

// NewColorRGBA creates a new color with default values (white, fully opaque)
func NewColorRGBA() *ColorRGBA {
	return &ColorRGBA{
		R: 0,
		G: 0,
		B: 0,
		A: 1.0,
	}
}

// NewColorRGBA creates a new color to highlight things.
func NewWarningColorRGBA() *ColorRGBA {
	return &ColorRGBA{
		R: 226,
		G: 59,
		B: 255,
		A: 1.0,
	}
}

// NewColorRGB creates a color with specified RGB values (alpha = 1.0)
func NewColorRGB(r, g, b uint8) *ColorRGB {
	return &ColorRGB{
		R: r,
		G: g,
		B: b,
	}
}

// NewColorRGBA creates a color with specified RGBA values
func NewColorRGBAValues(r, g, b uint8, a float64) (*ColorRGBA, error) {
	if a < 0 || a > 1.0 {
		return nil, errors.New("alpha must be between 0.0 and 1.0")
	}
	return &ColorRGBA{
		R: r,
		G: g,
		B: b,
		A: a,
	}, nil
}

// FromHex creates a color from a hexadecimal string (format: "#RRGGBB" or "#RRGGBBAA")
func FromHex(hex string) (*ColorRGBA, error) {
	// Remove leading '#' if present
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}

	var r, g, b uint8
	var a float64 = 1.0

	switch len(hex) {
	case 3: // #RGB format
		_, err := fmt.Sscanf(hex, "%1x%1x%1x", &r, &g, &b)
		if err != nil {
			return nil, fmt.Errorf("invalid hex color: %v", err)
		}
		// Scale up from 4-bit to 8-bit
		r *= 17
		g *= 17
		b *= 17

	case 6: // #RRGGBB format
		_, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
		if err != nil {
			return nil, fmt.Errorf("invalid hex color: %v", err)
		}

	case 8: // #RRGGBBAA format
		var alpha uint8
		_, err := fmt.Sscanf(hex, "%02x%02x%02x%02x", &r, &g, &b, &alpha)
		if err != nil {
			return nil, fmt.Errorf("invalid hex color: %v", err)
		}
		a = float64(alpha) / 255.0

	default:
		return nil, errors.New("hex color must be 3, 6, or 8 characters long (excluding #)")
	}

	return &ColorRGBA{
		R: r,
		G: g,
		B: b,
		A: a,
	}, nil
}

// ToHex returns the hexadecimal representation of the color
func (c *ColorRGBA) ToHex() string {
	return fmt.Sprintf("#%02X%02X%02X%02X", c.R, c.G, c.B, uint8(c.A*255))
}

// Lerp performs linear interpolation between two colors
func (c *ColorRGBA) Lerp(other *ColorRGBA, t float64) *ColorRGBA {
	if t <= 0 {
		return c
	}
	if t >= 1 {
		return other
	}

	return &ColorRGBA{
		R: uint8(float64(c.R) + (float64(other.R)-float64(c.R))*t),
		G: uint8(float64(c.G) + (float64(other.G)-float64(c.G))*t),
		B: uint8(float64(c.B) + (float64(other.B)-float64(c.B))*t),
		A: c.A + (other.A-c.A)*t,
	}
}

// Add blends two colors additively
func (c *ColorRGBA) Add(other *ColorRGBA) *ColorRGBA {
	return &ColorRGBA{
		R: uint8(math.Min(float64(c.R)+float64(other.R), 255)),
		G: uint8(math.Min(float64(c.G)+float64(other.G), 255)),
		B: uint8(math.Min(float64(c.B)+float64(other.B), 255)),
		A: math.Min(c.A+other.A, 1.0),
	}
}

// Multiply blends two colors multiplicatively
func (c *ColorRGBA) Multiply(other *ColorRGBA) *ColorRGBA {
	return &ColorRGBA{
		R: uint8(float64(c.R) * float64(other.R) / 255),
		G: uint8(float64(c.G) * float64(other.G) / 255),
		B: uint8(float64(c.B) * float64(other.B) / 255),
		A: c.A * other.A,
	}
}

// Scale scales the color by a factor (keeping alpha unchanged)
func (c *ColorRGBA) Scale(factor float64) *ColorRGBA {
	return &ColorRGBA{
		R: uint8(math.Min(float64(c.R)*factor, 255)),
		G: uint8(math.Min(float64(c.G)*factor, 255)),
		B: uint8(math.Min(float64(c.B)*factor, 255)),
		A: c.A,
	}
}

// WithAlpha returns a new color with modified alpha
func (c *ColorRGBA) WithAlpha(alpha float64) (*ColorRGBA, error) {
	if alpha < 0 || alpha > 1.0 {
		return nil, errors.New("alpha must be between 0.0 and 1.0")
	}
	return &ColorRGBA{
		R: c.R,
		G: c.G,
		B: c.B,
		A: alpha,
	}, nil
}

// ToFloat32 returns the color as float32 values (0-1 range)
func (c *ColorRGBA) ToFloat32() (r, g, b, a float32) {
	return float32(c.R) / 255, float32(c.G) / 255, float32(c.B) / 255, float32(c.A)
}

// Equals checks if two colors are equal (with epsilon comparison for alpha)
func (c *ColorRGBA) Equals(other *ColorRGBA) bool {
	const epsilon = 0.001
	return c.R == other.R &&
		c.G == other.G &&
		c.B == other.B &&
		math.Abs(c.A-other.A) < epsilon
}

// Grayscale converts the color to grayscale
func (c *ColorRGBA) Grayscale() *ColorRGBA {
	gray := uint8(0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B))
	return &ColorRGBA{
		R: gray,
		G: gray,
		B: gray,
		A: c.A,
	}
}

// String returns a string representation of the color
func (c *ColorRGBA) String() string {
	return fmt.Sprintf("ColorRGBA(%d, %d, %d, %.2f)", c.R, c.G, c.B, c.A)
}
