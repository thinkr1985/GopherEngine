package assets

import (
	"GopherEngine/lookdev"
	"GopherEngine/nomath"
	"GopherEngine/utilities"
	"fmt"
)

type Assembly struct {
	Name        string
	ID          string
	isDynamic   bool
	References  []*AssetReference
	Geometries  []*Geometry
	Materials   []*lookdev.Material
	Textures    []*lookdev.Texture
	Vertices    []*nomath.Vec3
	Triangles   []*Triangle
	Transform   *nomath.Transform
	BoundingBox *nomath.BoundingBox
	IsVisible   bool
	AssemblyMap map[string]any
}

func NewAssembly() *Assembly {
	a := Assembly{
		Name:        utilities.GenerateID(),
		ID:          utilities.GenerateUniqueID(),
		isDynamic:   false,
		IsVisible:   true,
		Transform:   nomath.NewTransform(),
		Geometries:  make([]*Geometry, 0),
		Materials:   make([]*lookdev.Material, 0),
		Textures:    make([]*lookdev.Texture, 0),
		Vertices:    make([]*nomath.Vec3, 0),
		Triangles:   make([]*Triangle, 0),
		BoundingBox: nomath.NewBoundingBox(),
		AssemblyMap: make(map[string]any),
	}
	a.ComputeBoundingBox()
	a.UpdateAssemblyMap()
	return &a
}

func (a *Assembly) UpdateAssemblyMap() {
	a.AssemblyMap["Geometries"] = a.Geometries
	a.AssemblyMap["References"] = a.References

}

func (a *Assembly) String() string {
	return fmt.Sprintf("Assembly(%v : %v, IsDynamic : %v, IsVisible : %v)", a.Name, a.ID, a.isDynamic, a.IsVisible)
}

func (a *Assembly) AddGeometry(geom *Geometry) {
	for _, g := range a.Geometries {
		if g == geom {
			return // Already present
		}
	}

	// Set parent-child relationship
	geom.Transform.SetParent(a.Transform)
	geom.Parent = a

	a.Geometries = append(a.Geometries, geom)
	a.Triangles = append(a.Triangles, geom.Triangles...)
	a.Vertices = append(a.Vertices, geom.Vertices...)
	a.Materials = append(a.Materials, geom.Material)
	a.ComputeBoundingBox()
	a.UpdateAssemblyMap()
}

func (a *Assembly) AddReference(name string, file_path string) {
	reference := NewAssetReference(name, file_path)
	a.References = append(a.References, reference)
	a.UpdateAssemblyMap()
}

func (a *Assembly) LoadReference() {
	for _, ref := range a.References {
		ref.LoadReference()
	}
}

func (a *Assembly) ClearGeometries() {
	a.Geometries = nil
	a.Vertices = nil
	a.Triangles = nil
	a.Materials = nil
	a.Textures = nil // optional, if you manage them similarly
	a.ComputeBoundingBox()
	a.UpdateAssemblyMap()
}

func (a *Assembly) RemoveGeometry(geom *Geometry) {
	// Remove geometry reference
	newGeometries := make([]*Geometry, 0, len(a.Geometries))
	for _, g := range a.Geometries {
		if g != geom {
			newGeometries = append(newGeometries, g)
		}
	}
	a.Geometries = newGeometries

	// Remove geom's vertices from Assembly
	if len(geom.Vertices) > 0 {
		newVertices := make([]*nomath.Vec3, 0, len(a.Vertices))
		vertexMap := make(map[*nomath.Vec3]bool)
		for _, v := range geom.Vertices {
			vertexMap[v] = true
		}
		for _, v := range a.Vertices {
			if !vertexMap[v] {
				newVertices = append(newVertices, v)
			}
		}
		a.Vertices = newVertices

	}
	geom.Parent = nil

	// Remove geom's triangles from Assembly
	if len(geom.Triangles) > 0 {
		newTriangles := make([]*Triangle, 0, len(a.Triangles))
		triangleMap := make(map[*Triangle]bool)
		for _, t := range geom.Triangles {
			triangleMap[t] = true
		}
		for _, t := range a.Triangles {
			if !triangleMap[t] {
				newTriangles = append(newTriangles, t)
			}
		}
		a.Triangles = newTriangles
	}

	// Remove material if not used by other geometries
	if geom.Material != nil {
		// Count how many geometries use this material
		usedElsewhere := false
		for _, g := range a.Geometries {
			if g.Material == geom.Material {
				usedElsewhere = true
				break
			}
		}
		if !usedElsewhere {
			newMaterials := make([]*lookdev.Material, 0, len(a.Materials))
			for _, m := range a.Materials {
				if m != geom.Material {
					newMaterials = append(newMaterials, m)
				}
			}
			a.Materials = newMaterials
		}
	}

	// Update bounding box
	a.ComputeBoundingBox()
	a.UpdateAssemblyMap()
}

func (a *Assembly) ReplaceGeometry(oldGeom, newGeom *Geometry) {
	if oldGeom == nil || newGeom == nil {
		return
	}
	found := false
	for _, g := range a.Geometries {
		if g == oldGeom {
			found = true
			break
		}
	}
	if !found {
		return // oldGeom not part of the assembly
	}
	a.RemoveGeometry(oldGeom)
	a.AddGeometry(newGeom)
	a.UpdateAssemblyMap()
}

func (a *Assembly) GetGeometryByID(id string) *Geometry {
	for _, g := range a.Geometries {
		if g.ID == id {
			return g
		}
	}
	return nil
}

func (a *Assembly) RemoveGeometryByID(id string) {
	geom := a.GetGeometryByID(id)
	if geom != nil {
		a.RemoveGeometry(geom)
	}
	a.UpdateAssemblyMap()
}

func (a *Assembly) SetDynamic() {
	a.isDynamic = true
	a.Transform.Dirty = true
	if len(a.Geometries) > 0 {
		for _, geom := range a.Geometries {
			geom.Transform.Dirty = true
		}

	}
}

func (a *Assembly) SetStatic() {
	a.isDynamic = false
	a.Transform.Dirty = false
	if len(a.Geometries) > 0 {
		for _, geom := range a.Geometries {
			geom.Transform.Dirty = false
		}

	}
}

func (a *Assembly) Update() {
	a.Transform.UpdateModelMatrix()

	// Update all geometries
	for _, geom := range a.Geometries {
		geom.Update()
	}

	// Recompute assembly's bounding box in world space
	a.ComputeBoundingBox()
}

func (a *Assembly) ComputeBoundingBox() {
	if len(a.Geometries) == 0 {
		a.BoundingBox = nomath.NewBoundingBox()
		return
	}

	// Start with empty bounding box
	a.BoundingBox = nomath.NewBoundingBox()

	// Combine all geometry bounding boxes in world space
	for _, geom := range a.Geometries {
		// Get geometry's world transform
		worldMatrix := geom.Transform.GetWorldMatrix()

		// Transform geometry's local bounding box to world space
		worldBB := geom.BoundingBox.Transform(worldMatrix)

		// Merge with assembly's bounding box
		a.BoundingBox = a.BoundingBox.Merge(worldBB)
	}
}

func (a *Assembly) ComputeTransformedBoundingBox() {
	if len(a.Vertices) == 0 {
		return
	}

	transform := a.Transform.GetWorldMatrix()

	first := transform.MultiplyVec4(a.Vertices[0].ToVec4(1)).ToVec3()
	min := first
	max := first

	for _, v := range a.Vertices[1:] {
		tv := transform.MultiplyVec4(v.ToVec4(1)).ToVec3()
		min = nomath.Min(min, tv)
		max = nomath.Max(max, tv)
	}

	a.BoundingBox = &nomath.BoundingBox{Min: min, Max: max}
}
