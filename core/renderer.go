package core

import (
	"GopherEngine/lookdev"
	"image"
	"image/color"
	"image/png"
	"os"
)

type Renderer3D struct {
	Framebuffer     [][]lookdev.ColorRGBA // Changed to value type
	DepthBuffer     [][]float32           // Changed to float32 for better cache usage
	BackFaceCulling bool
}

func NewRenderer3D() *Renderer3D {
	return &Renderer3D{
		BackFaceCulling: true,
		Framebuffer:     make([][]lookdev.ColorRGBA, SCREEN_WIDTH),
		DepthBuffer:     make([][]float32, SCREEN_HEIGHT),
	}

}

func (r *Renderer3D) ToImage() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, SCREEN_WIDTH, SCREEN_HEIGHT))
	for y := 0; y < SCREEN_HEIGHT; y++ {
		for x := 0; x < SCREEN_WIDTH; x++ {
			c := r.Framebuffer[y][x]
			img.SetRGBA(x, y, color.RGBA{
				R: c.R,
				G: c.G,
				B: c.B,
				A: 1,
			})
		}
	}
	return img
}

func (r *Renderer3D) SaveToPNG(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, r.ToImage())
}
