package assets

import (
	"GopherEngine/lookdev"
	"GopherEngine/nomath"
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"log"
	_ "log"
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
	IsVisible bool
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
			IsVisible: geom.IsVisible,
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
		log.Fatalf("Failed to read Asset : %v : %v", path, err)
		return nil, err
	}

	// Decompress if you used gzip when exporting
	var packed PackedAssembly
	{
		buf := bytes.NewBuffer(data)
		var dec *gob.Decoder

		// detect gzip by checking first two bytes (gzip magic 0x1f 0x8b)
		head := buf.Bytes()
		if len(head) >= 2 && head[0] == 0x1f && head[1] == 0x8b {
			gz, gzErr := gzip.NewReader(buf)
			if gzErr != nil {
				log.Fatalf("Failed to uncompress Asset : %v : %v", path, err)
				return nil, gzErr
			}
			defer gz.Close()
			dec = gob.NewDecoder(gz)
		} else {
			dec = gob.NewDecoder(buf)
		}

		if err = dec.Decode(&packed); err != nil {
			return nil, err
		}
	}

	// Rebuild textures
	textureMap := make(map[string]*lookdev.Texture)
	for _, pt := range packed.Textures {
		if tex := lookdev.NewTextureFromBytes(pt.Data, pt.Name); tex != nil {
			textureMap[pt.Name] = tex
		}
	}

	a := NewAssembly()
	a.Name = packed.Name
	a.ID = packed.ID
	a.isDynamic = packed.IsDynamic
	a.IsVisible = packed.IsVisible
	a.Transform = FromSerializableTransform(packed.Transform)
	if a.Transform == nil {
		a.Transform = nomath.NewTransform()
	}

	// Ensure assembly slices are initialized
	if a.Geometries == nil {
		a.Geometries = make([]*Geometry, 0)
	}
	if a.Triangles == nil {
		a.Triangles = make([]*Triangle, 0)
	}

	for _, g := range packed.Geometries {
		// Convert PackedTriangle -> Triangle (without mutexes)
		var tris []*Triangle
		for _, ptri := range g.Triangles {
			tri := &Triangle{
				V0:  ptri.V0,
				V1:  ptri.V1,
				V2:  ptri.V2,
				N0:  ptri.N0,
				N1:  ptri.N1,
				N2:  ptri.N2,
				UV0: ptri.UV0,
				UV1: ptri.UV1,
				UV2: ptri.UV2,
				// BufferCache false by default; material/parent will be assigned below
			}
			tris = append(tris, tri)
		}

		geom := &Geometry{
			ID:        g.ID,
			Name:      g.Name,
			Vertices:  g.Vertices,
			Normals:   g.Normals,
			UVs:       g.UVs,
			Triangles: tris,
			IsVisible: g.IsVisible,
			Transform: FromSerializableTransform(g.Transform),
		}
		if geom.Transform == nil {
			geom.Transform = nomath.NewTransform()
		}
		geom.Transform.UpdateModelMatrix()

		// material
		mat := &lookdev.Material{
			Name:          g.Material.Name,
			DiffuseColor:  g.Material.DiffuseColor,
			SpecularColor: g.Material.SpecularColor,
			Shininess:     g.Material.Shininess,
			Transparency:  g.Material.Transparency,
			Reflectivity:  g.Material.Reflectivity,
		}
		// attach textures from embedded blob map
		if g.Material.DiffuseTexture != "" {
			if t, ok := textureMap[g.Material.DiffuseTexture]; ok {
				mat.DiffuseTexture = t
			}
		}
		if g.Material.SpecularTexture != "" {
			if t, ok := textureMap[g.Material.SpecularTexture]; ok {
				mat.SpecularTexture = t
			}
		}
		if g.Material.NormalTexture != "" {
			if t, ok := textureMap[g.Material.NormalTexture]; ok {
				mat.NormalTexture = t
			}
		}
		if g.Material.TransparencyTexture != "" {
			if t, ok := textureMap[g.Material.TransparencyTexture]; ok {
				mat.TransparencyTexture = t
			}
		}
		geom.Material = mat

		// set triangle parent and material pointers, also fill assembly triangle list
		for _, tri := range geom.Triangles {
			tri.Parent = geom
			tri.Material = geom.Material
			// Optionally: compute triangle world normal etc. later in Update
			a.Triangles = append(a.Triangles, tri)
		}

		geom.Transform.Dirty = true
		geom.Transform.UpdateModelMatrix()
		geom.Transform.Dirty = false
		// compute geometry bounding box (safe)
		geom.ComputeBoundingBox()
		// Precompute triangle buffers if needed
		geom.PrecomputeTextureBuffers()

		a.Geometries = append(a.Geometries, geom)
	}
	a.ComputeBoundingBox()
	return a, nil
}
