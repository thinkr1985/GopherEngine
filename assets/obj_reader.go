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

	// Create new geometry with default material
	geomName := strings.TrimSuffix(filepath, ".obj")
	if geomName == filepath { // If no .obj extension was found
		geomName = filepath // Use full path as name
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
	}

	// Temporary storage for OBJ data
	var vertices []*nomath.Vec3
	var texCoords []*nomath.Vec2
	var normals []*nomath.Vec3

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
		case "v": // Vertex position
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
			vertex := &nomath.Vec3{X: x, Y: y, Z: z}
			vertices = append(vertices, vertex)
			geom.Vertices = append(geom.Vertices, vertex)

		case "vt": // Texture coordinate
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
			uv := &nomath.Vec2{U: u, V: v}
			texCoords = append(texCoords, uv)
			geom.UVs = append(geom.UVs, uv)

		case "vn": // Vertex normal
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
			normal := nomath.Vec3{X: x, Y: y, Z: z}.Normalize()
			normals = append(normals, &normal)
			geom.Normals = append(geom.Normals, &normal)

		case "f": // Face
			if len(data) < 3 {
				return nil, fmt.Errorf("line %d: face needs at least 3 vertices", lineNum)
			}

			var faceIndices []struct {
				vIdx, tIdx, nIdx int
			}

			// Parse all vertex data for the face
			for _, vertex := range data {
				parts := strings.Split(vertex, "/")
				if len(parts) == 0 {
					return nil, fmt.Errorf("line %d: invalid face vertex format", lineNum)
				}

				// Parse vertex index (required)
				vIdx, err := strconv.Atoi(parts[0])
				if err != nil {
					return nil, fmt.Errorf("line %d: invalid vertex index: %v", lineNum, err)
				}
				if vIdx < 0 {
					vIdx = len(vertices) + vIdx + 1
				}
				vIdx-- // Convert to 0-based index

				// Parse texture coordinate index (optional)
				tIdx := -1
				if len(parts) > 1 && parts[1] != "" {
					tIdx, err = strconv.Atoi(parts[1])
					if err != nil {
						return nil, fmt.Errorf("line %d: invalid texture coordinate index: %v", lineNum, err)
					}
					if tIdx < 0 {
						tIdx = len(texCoords) + tIdx + 1
					}
					tIdx-- // Convert to 0-based index
				}

				// Parse normal index (optional)
				nIdx := -1
				if len(parts) > 2 && parts[2] != "" {
					nIdx, err = strconv.Atoi(parts[2])
					if err != nil {
						return nil, fmt.Errorf("line %d: invalid normal index: %v", lineNum, err)
					}
					if nIdx < 0 {
						nIdx = len(normals) + nIdx + 1
					}
					nIdx-- // Convert to 0-based index
				}

				faceIndices = append(faceIndices, struct{ vIdx, tIdx, nIdx int }{vIdx, tIdx, nIdx})
			}

			// Triangulate polygon (assuming convex)
			for i := 1; i < len(faceIndices)-1; i++ {
				v0 := faceIndices[0]
				v1 := faceIndices[i]
				v2 := faceIndices[i+1]

				// Create triangle with references to the geometry's vertices/normals/UVs
				tri := NewTriangle(
					geom,
					geom.Material,
					vertices[v0.vIdx],
					vertices[v1.vIdx],
					vertices[v2.vIdx],
					nil, // Normals will be set below if available
					nil,
					nil,
					nil, // UVs will be set below if available
					nil,
					nil,
				)

				// Set normals if available
				if v0.nIdx >= 0 && v1.nIdx >= 0 && v2.nIdx >= 0 {
					tri.N0 = normals[v0.nIdx]
					tri.N1 = normals[v1.nIdx]
					tri.N2 = normals[v2.nIdx]
				}

				// Set texture coordinates if available
				if v0.tIdx >= 0 && v1.tIdx >= 0 && v2.tIdx >= 0 {
					tri.UV0 = texCoords[v0.tIdx]
					tri.UV1 = texCoords[v1.tIdx]
					tri.UV2 = texCoords[v2.tIdx]
				}

				geom.Triangles = append(geom.Triangles, tri)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading OBJ file: %v", err)
	}

	// Calculate normals if not provided in file
	if len(normals) == 0 {
		geom.CalculateNormals()
	}

	// Compute bounding box
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
