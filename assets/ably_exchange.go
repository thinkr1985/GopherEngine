package assets

import (
	"GopherEngine/lookdev"
	"GopherEngine/nomath"
	"GopherEngine/utilities"
	"encoding/json"
	"fmt"
	"log"
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
	IsVisible bool                         `json:"IsVisible"`
}

// --- Transform Conversion ---

func FromSerializableTransform(st nomath.SerializableTransform) *nomath.Transform {
	t := &nomath.Transform{
		Position: st.Position,
		Rotation: st.Rotation,
		Scale:    st.Scale,
	}
	t.Dirty = true
	t.UpdateModelMatrix() // Ensure matrix is calculated immediately
	t.Dirty = false
	return t
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

	// Write faces with explicit vertex/uv/normal indices
	for _, tri := range geom.Triangles {
		// Get indices for each component
		v0Idx := indexOfVec3(geom.Vertices, tri.V0) + 1
		v1Idx := indexOfVec3(geom.Vertices, tri.V1) + 1
		v2Idx := indexOfVec3(geom.Vertices, tri.V2) + 1

		uv0Idx := indexOfVec2(geom.UVs, tri.UV0) + 1
		uv1Idx := indexOfVec2(geom.UVs, tri.UV1) + 1
		uv2Idx := indexOfVec2(geom.UVs, tri.UV2) + 1

		n0Idx := indexOfVec3(geom.Normals, tri.N0) + 1
		n1Idx := indexOfVec3(geom.Normals, tri.N1) + 1
		n2Idx := indexOfVec3(geom.Normals, tri.N2) + 1

		// Write face with all indices
		sb.WriteString(fmt.Sprintf("f %d/%d/%d %d/%d/%d %d/%d/%d\n",
			v0Idx, uv0Idx, n0Idx,
			v1Idx, uv1Idx, n1Idx,
			v2Idx, uv2Idx, n2Idx))
	}

	_, err = f.WriteString(sb.String())
	if err != nil {
		return "", err
	}

	return filename, nil
}

// Helper functions to find indices
func indexOfVec3(slice []*nomath.Vec3, item *nomath.Vec3) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
}

func indexOfVec2(slice []*nomath.Vec2, item *nomath.Vec2) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
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
			IsVisible: geom.IsVisible,
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
		log.Fatalf("Failed to read : %v : %v\n", path, err)
		return
	}

	var saved SerializableAssembly
	err = json.Unmarshal(data, &saved)
	if err != nil {
		log.Fatalf("Failed to parse json : %v : %v\n", path, err)
		return
	}

	a.Name = saved.Name
	a.ID = saved.ID
	a.isDynamic = saved.IsDynamic
	a.IsVisible = saved.IsVisible

	// Fix: Ensure transform is properly initialized
	a.Transform = FromSerializableTransform(saved.Transform)
	if a.Transform == nil {
		a.Transform = nomath.NewTransform()
	}
	a.Transform.UpdateModelMatrix()

	assemblyDir := filepath.Dir(path)

	for _, sGeom := range saved.Geometries {
		objPath := filepath.Join(assemblyDir, sGeom.OBJPath)

		geom, err := LoadOBJ(objPath)
		if err != nil {
			log.Fatalf("ERROR : Failed to load .obj : %v : %v\n", path, err)
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

		// Fix: Ensure geometry transform is properly initialized
		geom.Transform = FromSerializableTransform(sGeom.Transform)
		if geom.Transform == nil {
			geom.Transform = nomath.NewTransform()
		}
		geom.Transform.UpdateModelMatrix()

		if sGeom.Material.DiffuseTexture != "" {
			tex, err := lookdev.LoadTexture(filepath.Join(assemblyDir, sGeom.Material.DiffuseTexture))
			if err == nil {
				geom.Material.DiffuseTexture = tex
			} else {
				log.Fatalf("Failed to load texture : %v", filepath.Join(assemblyDir, sGeom.Material.DiffuseTexture))
			}
		}
		if sGeom.Material.SpecularTexture != "" {
			tex, err := lookdev.LoadTexture(filepath.Join(assemblyDir, sGeom.Material.SpecularTexture))
			if err == nil {
				geom.Material.SpecularTexture = tex
			} else {
				log.Fatalf("Failed to load texture : %v", filepath.Join(assemblyDir, sGeom.Material.SpecularTexture))
			}
		}
		if sGeom.Material.NormalTexture != "" {
			tex, err := lookdev.LoadTexture(filepath.Join(assemblyDir, sGeom.Material.NormalTexture))
			if err == nil {
				geom.Material.NormalTexture = tex
			} else {
				log.Fatalf("Failed to load texture : %v", filepath.Join(assemblyDir, sGeom.Material.NormalTexture))
			}
		}
		if sGeom.Material.TransparencyTexture != "" {
			tex, err := lookdev.LoadTexture(filepath.Join(assemblyDir, sGeom.Material.TransparencyTexture))
			if err == nil {
				geom.Material.TransparencyTexture = tex
			} else {
				log.Fatalf("Failed to load texture : %v", filepath.Join(assemblyDir, sGeom.Material.TransparencyTexture))
			}
		}
		geom.ComputeBoundingBox()
		a.AddGeometry(geom)
	}
	a.ComputeBoundingBox()
}
