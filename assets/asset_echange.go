package assets

import (
	"GopherEngine/lookdev"
	"GopherEngine/nomath"
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
)

// --- Packed data structs ---

type PackedTexture struct {
	Name string
	Data []byte // raw PNG/JPEG image bytes
}

type PackedTriangle struct {
	V0, V1, V2    *nomath.Vec3
	N0, N1, N2    *nomath.Vec3
	UV0, UV1, UV2 *nomath.Vec2
}

type PackedGeometry struct {
	ID        string
	Name      string
	Vertices  []*nomath.Vec3
	Normals   []*nomath.Vec3
	UVs       []*nomath.Vec2
	Triangles []PackedTriangle
	Material  lookdev.SerializableMaterial
	Transform nomath.SerializableTransform
}

type PackedAssembly struct {
	Name       string
	ID         string
	IsDynamic  bool
	IsVisible  bool
	Transform  nomath.SerializableTransform
	Geometries []PackedGeometry
	Textures   []PackedTexture
}

// --- Asset Exporter ---

func AssetExport(assembly *Assembly, path string) error {
	var texturesMap = make(map[string][]byte)

	packed := PackedAssembly{
		Name:      assembly.Name,
		ID:        assembly.ID,
		IsDynamic: assembly.isDynamic,
		IsVisible: assembly.IsVisible,
		Transform: assembly.Transform.ToSerializable(),
	}

	for _, geom := range assembly.Geometries {
		var packedTris []PackedTriangle
		for _, tri := range geom.Triangles {
			packedTris = append(packedTris, PackedTriangle{
				V0: tri.V0, V1: tri.V1, V2: tri.V2,
				N0: tri.N0, N1: tri.N1, N2: tri.N2,
				UV0: tri.UV0, UV1: tri.UV1, UV2: tri.UV2,
			})
		}

		packedGeo := PackedGeometry{
			ID:        geom.ID,
			Name:      geom.Name,
			Vertices:  geom.Vertices,
			Normals:   geom.Normals,
			UVs:       geom.UVs,
			Triangles: packedTris,
			Material: lookdev.SerializableMaterial{
				Name:          geom.Material.Name,
				DiffuseColor:  geom.Material.DiffuseColor,
				SpecularColor: geom.Material.SpecularColor,
				Shininess:     geom.Material.Shininess,
				Transparency:  geom.Material.Transparency,
				Reflectivity:  geom.Material.Reflectivity,
			},
			Transform: geom.Transform.ToSerializable(),
		}

		// Load and embed textures
		m := geom.Material
		for _, tex := range []*lookdev.Texture{m.DiffuseTexture, m.SpecularTexture, m.NormalTexture, m.TransparencyTexture} {
			if tex != nil && tex.Path != "" {
				if _, exists := texturesMap[tex.Path]; !exists {
					data, err := os.ReadFile(tex.Path)
					if err != nil {
						fmt.Println("Warning: Could not read texture", tex.Path)
						continue
					}
					texturesMap[tex.Path] = data
				}
			}
		}

		if m.DiffuseTexture != nil {
			packedGeo.Material.DiffuseTexture = filepath.Base(m.DiffuseTexture.Path)
		}
		if m.SpecularTexture != nil {
			packedGeo.Material.SpecularTexture = filepath.Base(m.SpecularTexture.Path)
		}
		if m.NormalTexture != nil {
			packedGeo.Material.NormalTexture = filepath.Base(m.NormalTexture.Path)
		}
		if m.TransparencyTexture != nil {
			packedGeo.Material.TransparencyTexture = filepath.Base(m.TransparencyTexture.Path)
		}

		packed.Geometries = append(packed.Geometries, packedGeo)
	}

	// Add textures to PackedAssembly
	for path, data := range texturesMap {
		packed.Textures = append(packed.Textures, PackedTexture{
			Name: filepath.Base(path),
			Data: data,
		})
	}

	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	enc := gob.NewEncoder(gzipWriter)

	err := enc.Encode(packed)
	if err != nil {
		return err
	}
	gzipWriter.Close()

	err = os.WriteFile(path, buf.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}

// --- Asset Importer ---

func AssetImport(path string) (*Assembly, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var packed PackedAssembly
	buf := bytes.NewBuffer(data)

	gzipReader, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	defer gzipReader.Close()

	dec := gob.NewDecoder(gzipReader)
	err = dec.Decode(&packed)
	if err != nil {
		return nil, err
	}

	// Rebuild textures
	textureMap := make(map[string]*lookdev.Texture)
	for _, pt := range packed.Textures {
		tex := lookdev.NewTextureFromBytes(pt.Data, pt.Name)
		textureMap[pt.Name] = tex
	}

	a := NewAssembly()
	a.Name = packed.Name
	a.ID = packed.ID
	a.isDynamic = packed.IsDynamic
	a.IsVisible = packed.IsVisible
	a.Transform = FromSerializableTransform(packed.Transform)

	for _, g := range packed.Geometries {
		var tris []*Triangle
		for _, ptri := range g.Triangles {
			tris = append(tris, &Triangle{
				V0: ptri.V0, V1: ptri.V1, V2: ptri.V2,
				N0: ptri.N0, N1: ptri.N1, N2: ptri.N2,
				UV0: ptri.UV0, UV1: ptri.UV1, UV2: ptri.UV2,
			})
		}

		geom := &Geometry{
			ID:        g.ID,
			Name:      g.Name,
			Vertices:  g.Vertices,
			Normals:   g.Normals,
			UVs:       g.UVs,
			Triangles: tris,
			Transform: FromSerializableTransform(g.Transform),
		}
		geom.Transform.UpdateModelMatrix()

		mat := &lookdev.Material{
			Name:          g.Material.Name,
			DiffuseColor:  g.Material.DiffuseColor,
			SpecularColor: g.Material.SpecularColor,
			Shininess:     g.Material.Shininess,
			Transparency:  g.Material.Transparency,
			Reflectivity:  g.Material.Reflectivity,
		}
		// Reattach textures
		if tex := textureMap[g.Material.DiffuseTexture]; tex != nil {
			mat.DiffuseTexture = tex
		}
		if tex := textureMap[g.Material.SpecularTexture]; tex != nil {
			mat.SpecularTexture = tex
		}
		if tex := textureMap[g.Material.NormalTexture]; tex != nil {
			mat.NormalTexture = tex
		}
		if tex := textureMap[g.Material.TransparencyTexture]; tex != nil {
			mat.TransparencyTexture = tex
		}

		geom.Material = mat
		a.AddGeometry(geom)
	}

	return a, nil
}
