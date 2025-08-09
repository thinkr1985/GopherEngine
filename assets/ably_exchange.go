package assets

import (
	"GopherEngine/lookdev"
	"GopherEngine/nomath"
	"GopherEngine/utilities"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Serializable structs

type SerializableAssembly struct {
	Name       string                       `json:"name"`
	ID         string                       `json:"id"`
	IsDynamic  bool                         `json:"is_dynamic"`
	IsVisible  bool                         `json:"is_visible"`
	Transform  nomath.SerializableTransform `json:"transform"`
	Geometries []SerializableGeomRef        `json:"geometries"`
}

type SerializableGeomRef struct {
	ID        string                       `json:"id"`
	Name      string                       `json:"name"`
	OBJPath   string                       `json:"obj_path"`
	Material  lookdev.SerializableMaterial `json:"material"`
	Transform nomath.SerializableTransform `json:"transform"`
}

// --- Transform Conversion ---

func FromSerializableTransform(st nomath.SerializableTransform) *nomath.Transform {
	return &nomath.Transform{
		Position: st.Position,
		Rotation: st.Rotation,
		Scale:    st.Scale,
	}
}

// --- OBJ Export Helper ---

func exportGeometryAsOBJ(geom *Geometry, folder string) (string, error) {
	filename := fmt.Sprintf("%s.obj", geom.Name)
	objPath := filepath.Join(folder, filename)

	f, err := os.Create(objPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var sb strings.Builder

	// Write vertices
	for _, v := range geom.Vertices {
		sb.WriteString(fmt.Sprintf("v %f %f %f\n", v.X, v.Y, v.Z))
	}

	// Write normals
	for _, n := range geom.Normals {
		sb.WriteString(fmt.Sprintf("vn %f %f %f\n", n.X, n.Y, n.Z))
	}

	// Write texture coordinates
	for _, uv := range geom.UVs {
		sb.WriteString(fmt.Sprintf("vt %f %f\n", uv.U, uv.V))
	}

	// Write faces
	for _, tri := range geom.Triangles {
		vIdx := func(v *nomath.Vec3) int {
			for i, vv := range geom.Vertices {
				if vv == v {
					return i + 1
				}
			}
			return -1
		}

		uvIdx := func(uv *nomath.Vec2) int {
			for i, u := range geom.UVs {
				if u == uv {
					return i + 1
				}
			}
			return 0
		}

		nIdx := func(n *nomath.Vec3) int {
			for i, nn := range geom.Normals {
				if nn == n {
					return i + 1
				}
			}
			return 0
		}

		a := fmt.Sprintf("%d/%d/%d", vIdx(tri.V0), uvIdx(tri.UV0), nIdx(tri.N0))
		b := fmt.Sprintf("%d/%d/%d", vIdx(tri.V1), uvIdx(tri.UV1), nIdx(tri.N1))
		c := fmt.Sprintf("%d/%d/%d", vIdx(tri.V2), uvIdx(tri.UV2), nIdx(tri.N2))

		sb.WriteString(fmt.Sprintf("f %s %s %s\n", a, b, c))
	}

	_, err = f.WriteString(sb.String())
	if err != nil {
		return "", err
	}

	return filename, nil
}

// --- Save Assembly with External Geometry References ---

func (a *Assembly) SaveAssembly(name string, folderPath string) string {
	out := SerializableAssembly{
		Name:      a.Name,
		ID:        a.ID,
		IsDynamic: a.isDynamic,
		IsVisible: a.IsVisible,
		Transform: func() nomath.SerializableTransform {
			if a.Transform != nil {
				return a.Transform.ToSerializable()
			}
			return nomath.NewTransform().ToSerializable()
		}(),
	}

	for _, geom := range a.Geometries {
		// Export geometry as .obj
		objFile, err := exportGeometryAsOBJ(geom, folderPath)
		if err != nil {
			fmt.Println("Failed to export .obj:", err)
			continue
		}

		sGeomRef := SerializableGeomRef{
			ID:      geom.ID,
			Name:    geom.Name,
			OBJPath: objFile,
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

		if geom.Material.DiffuseTexture != nil {
			texPath := geom.Material.DiffuseTexture.Path
			dst := filepath.Join(folderPath, filepath.Base(texPath))
			_ = utilities.CopyFile(texPath, dst)                      // copy it
			sGeomRef.Material.DiffuseTexture = filepath.Base(texPath) // save as relative path
		}

		if geom.Material.SpecularTexture != nil {
			texPath := geom.Material.SpecularTexture.Path
			dst := filepath.Join(folderPath, filepath.Base(texPath))
			_ = utilities.CopyFile(texPath, dst)
			sGeomRef.Material.SpecularTexture = filepath.Base(texPath)
		}
		if geom.Material.NormalTexture != nil {
			texPath := geom.Material.NormalTexture.Path
			dst := filepath.Join(folderPath, filepath.Base(texPath))
			_ = utilities.CopyFile(texPath, dst)
			sGeomRef.Material.NormalTexture = filepath.Base(texPath)
		}
		if geom.Material.TransparencyTexture != nil {
			texPath := geom.Material.TransparencyTexture.Path
			dst := filepath.Join(folderPath, filepath.Base(texPath))
			_ = utilities.CopyFile(texPath, dst)
			sGeomRef.Material.TransparencyTexture = filepath.Base(texPath)
		}

		out.Geometries = append(out.Geometries, sGeomRef)
	}

	// Write JSON
	jsonData, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		fmt.Println("Failed to marshal assembly:", err)
		return ""
	}

	filePath := filepath.Join(folderPath, name+".ably")
	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		fmt.Println("Failed to write assembly JSON:", err)
		return ""
	}

	return filePath
}

// --- Load Assembly from JSON and External OBJ Files ---

func (a *Assembly) LoadAssembly(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Failed to read assembly JSON:", err)
		return
	}

	var saved SerializableAssembly
	err = json.Unmarshal(data, &saved)
	if err != nil {
		fmt.Println("Failed to parse JSON:", err)
		return
	}

	a.Name = saved.Name
	a.ID = saved.ID
	a.isDynamic = saved.IsDynamic
	a.IsVisible = saved.IsVisible
	a.Transform = FromSerializableTransform(saved.Transform)

	assemblyDir := filepath.Dir(path)

	for _, sGeom := range saved.Geometries {
		objPath := filepath.Join(assemblyDir, sGeom.OBJPath)

		geom, err := LoadOBJ(objPath)
		if err != nil {
			fmt.Println("Failed to load .obj geometry:", err)
			continue
		}

		geom.ID = sGeom.ID
		geom.Name = sGeom.Name
		geom.Material = &lookdev.Material{
			Name:          sGeom.Material.Name,
			DiffuseColor:  sGeom.Material.DiffuseColor,
			SpecularColor: sGeom.Material.SpecularColor,
			Shininess:     sGeom.Material.Shininess,
			Transparency:  sGeom.Material.Transparency,
			Reflectivity:  sGeom.Material.Reflectivity,
		}
		geom.Transform = FromSerializableTransform(sGeom.Transform)
		geom.Transform.UpdateModelMatrix()

		if sGeom.Material.DiffuseTexture != "" {
			tex, err := lookdev.LoadTexture(filepath.Join(assemblyDir, sGeom.Material.DiffuseTexture))
			if err == nil {
				geom.Material.DiffuseTexture = tex
			}
		}
		if sGeom.Material.SpecularTexture != "" {
			tex, err := lookdev.LoadTexture(filepath.Join(assemblyDir, sGeom.Material.SpecularTexture))
			if err == nil {
				geom.Material.SpecularTexture = tex
			}
		}
		if sGeom.Material.NormalTexture != "" {
			tex, err := lookdev.LoadTexture(filepath.Join(assemblyDir, sGeom.Material.NormalTexture))
			if err == nil {
				geom.Material.NormalTexture = tex
			}
		}
		if sGeom.Material.TransparencyTexture != "" {
			tex, err := lookdev.LoadTexture(filepath.Join(assemblyDir, sGeom.Material.TransparencyTexture))
			if err == nil {
				geom.Material.TransparencyTexture = tex
			}
		}

		a.AddGeometry(geom)
	}
}
