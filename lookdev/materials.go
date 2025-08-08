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
		Transparency:  1.0,
		Shininess:     50.0,
		Reflectivity:  0.0,
	}
}

type SerializableMaterial struct {
	Name                string    `json:"name"`
	DiffuseColor        ColorRGBA `json:"diffuse_color"`
	SpecularColor       ColorRGBA `json:"specular_color"`
	Shininess           float64   `json:"shininess"`
	Transparency        float64   `json:"transparency"`
	Reflectivity        float64   `json:"reflectivity"`
	DiffuseTexture      string    `json:"diffuse_texture,omitempty"`
	SpecularTexture     string    `json:"specular_texture,omitempty"`
	NormalTexture       string    `json:"normal_texture,omitempty"`
	TransparencyTexture string    `json:"transparency_texture,omitempty"`
}
