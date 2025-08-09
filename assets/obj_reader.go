package assets

import (
	"GopherEngine/lookdev"
	"GopherEngine/nomath"
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// LoadOBJ loads a Wavefront OBJ file and returns a Geometry object
func LoadOBJ(filepath string) (*Geometry, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	geomName := strings.TrimSuffix(filepath, ".obj")
	if geomName == filepath {
		geomName = filepath
	}

	geom := &Geometry{
		Name:        geomName,
		Transform:   nomath.NewTransform(),
		Vertices:    make([]*nomath.Vec3, 0),
		Normals:     make([]*nomath.Vec3, 0),
		UVs:         make([]*nomath.Vec2, 0),
		Triangles:   make([]*Triangle, 0),
		BoundingBox: nomath.NewBoundingBox(),
		Material:    lookdev.NewMaterial(geomName + "_material"),
		IsVisible:   true,
	}
	geom.Transform.Dirty = true
	var vertices []nomath.Vec3
	var texCoords []nomath.Vec2
	var normals []nomath.Vec3

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		prefix := fields[0]
		data := fields[1:]

		switch prefix {
		case "v":
			if len(data) < 3 {
				return nil, fmt.Errorf("line %d: vertex needs at least 3 coordinates", lineNum)
			}
			x, err := strconv.ParseFloat(data[0], 64)
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid vertex X coordinate: %v", lineNum, err)
			}
			y, err := strconv.ParseFloat(data[1], 64)
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid vertex Y coordinate: %v", lineNum, err)
			}
			z, err := strconv.ParseFloat(data[2], 64)
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid vertex Z coordinate: %v", lineNum, err)
			}
			vertices = append(vertices, nomath.Vec3{X: x, Y: y, Z: z})

		case "vt":
			if len(data) < 2 {
				return nil, fmt.Errorf("line %d: texture coordinate needs at least 2 values", lineNum)
			}
			u, err := strconv.ParseFloat(data[0], 64)
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid texture U coordinate: %v", lineNum, err)
			}
			v, err := strconv.ParseFloat(data[1], 64)
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid texture V coordinate: %v", lineNum, err)
			}
			texCoords = append(texCoords, nomath.Vec2{U: u, V: v})

		case "vn":
			if len(data) < 3 {
				return nil, fmt.Errorf("line %d: normal needs at least 3 coordinates", lineNum)
			}
			x, err := strconv.ParseFloat(data[0], 64)
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid normal X coordinate: %v", lineNum, err)
			}
			y, err := strconv.ParseFloat(data[1], 64)
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid normal Y coordinate: %v", lineNum, err)
			}
			z, err := strconv.ParseFloat(data[2], 64)
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid normal Z coordinate: %v", lineNum, err)
			}
			normals = append(normals, nomath.Vec3{X: x, Y: y, Z: z}.Normalize())

		case "f":
			if len(data) < 3 {
				return nil, fmt.Errorf("line %d: face needs at least 3 vertices", lineNum)
			}

			var faceVerts []struct{ v, vt, vn int }
			for _, vertex := range data {
				parts := strings.Split(vertex, "/")
				if len(parts) == 0 {
					return nil, fmt.Errorf("line %d: invalid face vertex format", lineNum)
				}

				// Parse vertex index
				vIdx, err := strconv.Atoi(parts[0])
				if err != nil {
					return nil, fmt.Errorf("line %d: invalid vertex index: %v", lineNum, err)
				}
				if vIdx < 0 {
					vIdx = len(vertices) + vIdx + 1
				}
				vIdx--

				// Parse texture coordinate index (if exists)
				vtIdx := -1
				if len(parts) > 1 && parts[1] != "" {
					vtIdx, err = strconv.Atoi(parts[1])
					if err != nil {
						return nil, fmt.Errorf("line %d: invalid texture coordinate index: %v", lineNum, err)
					}
					if vtIdx < 0 {
						vtIdx = len(texCoords) + vtIdx + 1
					}
					vtIdx--
				}

				// Parse normal index (if exists)
				vnIdx := -1
				if len(parts) > 2 && parts[2] != "" {
					vnIdx, err = strconv.Atoi(parts[2])
					if err != nil {
						return nil, fmt.Errorf("line %d: invalid normal index: %v", lineNum, err)
					}
					if vnIdx < 0 {
						vnIdx = len(normals) + vnIdx + 1
					}
					vnIdx--
				}

				faceVerts = append(faceVerts, struct{ v, vt, vn int }{vIdx, vtIdx, vnIdx})
			}

			// Triangulate polygon (assuming convex)
			for i := 1; i < len(faceVerts)-1; i++ {
				v0 := faceVerts[0]
				v1 := faceVerts[i]
				v2 := faceVerts[i+1]

				// Create triangle with references to the geometry's vertices/normals/UVs
				v0Ptr := &vertices[v0.v]
				v1Ptr := &vertices[v1.v]
				v2Ptr := &vertices[v2.v]

				var uv0Ptr, uv1Ptr, uv2Ptr *nomath.Vec2
				var n0Ptr, n1Ptr, n2Ptr *nomath.Vec3

				if v0.vt >= 0 && v1.vt >= 0 && v2.vt >= 0 {
					uv0Ptr = &texCoords[v0.vt]
					uv1Ptr = &texCoords[v1.vt]
					uv2Ptr = &texCoords[v2.vt]
				}

				if v0.vn >= 0 && v1.vn >= 0 && v2.vn >= 0 {
					n0Ptr = &normals[v0.vn]
					n1Ptr = &normals[v1.vn]
					n2Ptr = &normals[v2.vn]
				}

				tri := NewTriangle(
					geom,
					geom.Material,
					v0Ptr, v1Ptr, v2Ptr,
					n0Ptr, n1Ptr, n2Ptr,
					uv0Ptr, uv1Ptr, uv2Ptr,
				)

				geom.Triangles = append(geom.Triangles, tri)
			}
		}

	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading OBJ file: %v", err)
	}

	// Store the vertices/normals/UVs in the geometry
	for i := range vertices {
		geom.Vertices = append(geom.Vertices, &vertices[i])
	}
	for i := range texCoords {
		geom.UVs = append(geom.UVs, &texCoords[i])
	}
	for i := range normals {
		geom.Normals = append(geom.Normals, &normals[i])
	}

	// Calculate normals if not provided in file
	if len(normals) == 0 {
		geom.CalculateNormals()
	}
	geom.Transform.Dirty = true
	geom.Transform.UpdateModelMatrix()
	geom.ComputeBoundingBox()

	return geom, nil
}

// CalculateNormals computes vertex normals by averaging face normals
func (g *Geometry) CalculateNormals() {
	// First pass: calculate face normals and accumulate at vertices
	vertexNormals := make([]nomath.Vec3, len(g.Vertices))
	faceNormals := make([]nomath.Vec3, len(g.Triangles))

	for i, tri := range g.Triangles {
		// Calculate face normal
		edge1 := tri.V1.Subtract(*tri.V0)
		edge2 := tri.V2.Subtract(*tri.V0)
		normal := edge1.Cross(edge2).Normalize()
		faceNormals[i] = normal

		// Accumulate at each vertex
		vertexNormals[getVertexIndex(g.Vertices, tri.V0)] = vertexNormals[getVertexIndex(g.Vertices, tri.V0)].Add(normal)
		vertexNormals[getVertexIndex(g.Vertices, tri.V1)] = vertexNormals[getVertexIndex(g.Vertices, tri.V1)].Add(normal)
		vertexNormals[getVertexIndex(g.Vertices, tri.V2)] = vertexNormals[getVertexIndex(g.Vertices, tri.V2)].Add(normal)
	}

	// Second pass: normalize vertex normals and assign to triangles
	for i := range vertexNormals {
		vertexNormals[i] = vertexNormals[i].Normalize()
	}

	// Assign normals to triangles
	for _, tri := range g.Triangles {
		if tri.N0 == nil {
			tri.N0 = &vertexNormals[getVertexIndex(g.Vertices, tri.V0)]
		}
		if tri.N1 == nil {
			tri.N1 = &vertexNormals[getVertexIndex(g.Vertices, tri.V1)]
		}
		if tri.N2 == nil {
			tri.N2 = &vertexNormals[getVertexIndex(g.Vertices, tri.V2)]
		}
	}
}

// Helper function to find index of a vertex pointer in the slice
func getVertexIndex(vertices []*nomath.Vec3, v *nomath.Vec3) int {
	for i, vertex := range vertices {
		if vertex == v {
			return i
		}
	}
	return -1
}
