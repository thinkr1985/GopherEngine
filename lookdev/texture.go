package lookdev

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
)

type Texture struct {
	Width, Height int
	Pixels        []ColorRGBA
}

func LoadTexture(filename string) (*Texture, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	pixels := make([]ColorRGBA, width*height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(x+bounds.Min.X, y+bounds.Min.Y).RGBA()
			pixels[y*width+x] = ColorRGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: float64(a>>8) / 255.0, // This looks correct
			}
		}
	}

	return &Texture{
		Width:  width,
		Height: height,
		Pixels: pixels,
	}, nil
}

func (t *Texture) Sample(u, v float64) ColorRGBA {
	u = u - math.Floor(u)
	v = v - math.Floor(v)
	x := int(u * float64(t.Width-1))
	y := int(v * float64(t.Height-1))
	x = max(0, min(x, t.Width-1))
	y = max(0, min(y, t.Height-1))

	return t.Pixels[y*t.Width+x]
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
