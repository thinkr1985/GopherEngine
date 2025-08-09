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
	Geometries  []*Geometry
	Materials   []*lookdev.Material
	Textures    []*lookdev.Texture
	Vertices    []*nomath.Vec3
	Triangles   []*Triangle
	Transform   *nomath.Transform
	BoundingBox *nomath.BoundingBox
	IsVisible   bool
}

func NewAssembly() *Assembly {
	a := Assembly{
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
	}
	a.ComputeBoundingBox()
	return &a
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

	a.Geometries = append(a.Geometries, geom)
	a.Triangles = append(a.Triangles, geom.Triangles...)
	a.Vertices = append(a.Vertices, geom.Vertices...)
	a.Materials = append(a.Materials, geom.Material)
	a.ComputeBoundingBox()
}

func (a *Assembly) ClearGeometries() {
	a.Geometries = nil
	a.Vertices = nil
	a.Triangles = nil
	a.Materials = nil
	a.Textures = nil // optional, if you manage them similarly
	a.ComputeBoundingBox()
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
	if len(a.Geometries) > 0 {
		for _, geom := range a.Geometries {
			geom.Update()
		}

	}
	a.ComputeTransformedBoundingBox()
}

func (a *Assembly) ComputeBoundingBox() {
	if len(a.Vertices) == 0 {
		return
	}

	min := *a.Vertices[0]
	max := *a.Vertices[0]

	for _, v := range a.Vertices[1:] {
		if v.X < min.X {
			min.X = v.X
		}
		if v.Y < min.Y {
			min.Y = v.Y
		}
		if v.Z < min.Z {
			min.Z = v.Z
		}
		if v.X > max.X {
			max.X = v.X
		}
		if v.Y > max.Y {
			max.Y = v.Y
		}
		if v.Z > max.Z {
			max.Z = v.Z
		}
	}

	a.BoundingBox = &nomath.BoundingBox{Min: min, Max: max}
}

func (a *Assembly) ComputeTransformedBoundingBox() {
	if len(a.Vertices) == 0 {
		return
	}

	transform := a.Transform.GetMatrix()

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
