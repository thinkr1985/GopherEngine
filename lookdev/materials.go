package lookdev

type Material struct {
	Name                string
	DiffuseColor        ColorRGBA
	SpecularColor       ColorRGBA
	Shininess           float64
	Transparency        float64
	Reflectivity        float64
	DiffuseTexture      *Texture
	SpecularTexture     *Texture
	NormalTexture       *Texture
	TransparencyTexture *Texture
}

func NewMaterial(name string) *Material {
	return &Material{
		Name:          name,
		DiffuseColor:  ColorRGBA{R: 166, G: 166, B: 166, A: 1.0},
		SpecularColor: ColorRGBA{R: 0, G: 0, B: 0, A: 1},
		Transparency:  0.0,
		Shininess:     32.0,
		Reflectivity:  0.0,
	}
}
