package core

import (
	"GopherEngine/nomath"
	"math"
	"sync"
)

var frustumMutex sync.Mutex

type PerspectiveCamera struct {
	Name          string
	Scene         *Scene
	Transform     *nomath.Transform
	FocalLength   int
	NearPlane     float64
	FarPlane      float64
	frustumPlanes [6]nomath.Vec4 // Stores precomputed frustum planes
	DirtyFrustum  bool           // Flag to avoid recalculating planes unnecessarily
	mutex         sync.Mutex
}

func NewPerspectiveCamera() *PerspectiveCamera {
	cam := &PerspectiveCamera{
		Transform:    nomath.NewTransform(),
		FocalLength:  75,      // Reasonable default
		NearPlane:    0.1,     // Should be > 0
		FarPlane:     10000.0, // Large enough to see distant objects
		DirtyFrustum: true,
	}
	cam.Transform.Position = nomath.Vec3{Z: 10, Y: 10} // Start 10 units back
	cam.Transform.Dirty = true
	return cam
}

// Frustum planes indices
const (
	PlaneNear   = 0
	PlaneFar    = 1
	PlaneLeft   = 2
	PlaneRight  = 3
	PlaneTop    = 4
	PlaneBottom = 5
)

// GetViewMatrix returns the camera's view matrix
func (c *PerspectiveCamera) GetViewMatrix() nomath.Mat4 {
	c.Transform.UpdateModelMatrix()
	view := c.Transform.GetMatrix().Inverse()
	c.DirtyFrustum = true // View matrix changed, need to update planes
	return view
}

// GetProjectionMatrix returns the camera's projection matrix
func (c *PerspectiveCamera) GetProjectionMatrix() nomath.Mat4 {
	aspect := float64(SCREEN_WIDTH) / float64(SCREEN_HEIGHT)
	f := 1.0 / math.Tan(float64(c.FocalLength)*math.Pi/360.0)

	return nomath.Mat4{
		f / aspect, 0, 0, 0,
		0, f, 0, 0,
		0, 0, (c.FarPlane + c.NearPlane) / (c.NearPlane - c.FarPlane), -1,
		0, 0, (2 * c.FarPlane * c.NearPlane) / (c.NearPlane - c.FarPlane), 0,
	}
}

func (c *PerspectiveCamera) GetFrustumPlanes() [6]nomath.Vec4 {
	frustumMutex.Lock()
	defer frustumMutex.Unlock()
	c.UpdateFrustumPlanes()
	return c.frustumPlanes
}

// UpdateFrustumPlanes calculates the 6 frustum planes in world space
func (c *PerspectiveCamera) UpdateFrustumPlanes() {
	if !c.DirtyFrustum {
		return
	}

	viewProj := c.GetProjectionMatrix().Multiply(c.GetViewMatrix())

	// Extract planes from the combined view-projection matrix
	// Using the "Fast Extraction of Viewing Frustum Planes" method
	// (http://www.cs.otago.ac.nz/postgrads/alexis/planeExtraction.pdf)

	// Left plane
	c.frustumPlanes[PlaneLeft] = nomath.Vec4{
		X: viewProj[3] + viewProj[0],
		Y: viewProj[7] + viewProj[4],
		Z: viewProj[11] + viewProj[8],
		W: viewProj[15] + viewProj[12],
	}.Normalize()

	// Right plane
	c.frustumPlanes[PlaneRight] = nomath.Vec4{
		X: viewProj[3] - viewProj[0],
		Y: viewProj[7] - viewProj[4],
		Z: viewProj[11] - viewProj[8],
		W: viewProj[15] - viewProj[12],
	}.Normalize()

	// Bottom plane
	c.frustumPlanes[PlaneBottom] = nomath.Vec4{
		X: viewProj[3] + viewProj[1],
		Y: viewProj[7] + viewProj[5],
		Z: viewProj[11] + viewProj[9],
		W: viewProj[15] + viewProj[13],
	}.Normalize()

	// Top plane
	c.frustumPlanes[PlaneTop] = nomath.Vec4{
		X: viewProj[3] - viewProj[1],
		Y: viewProj[7] - viewProj[5],
		Z: viewProj[11] - viewProj[9],
		W: viewProj[15] - viewProj[13],
	}.Normalize()

	// Near plane
	c.frustumPlanes[PlaneNear] = nomath.Vec4{
		X: viewProj[3] + viewProj[2],
		Y: viewProj[7] + viewProj[6],
		Z: viewProj[11] + viewProj[10],
		W: viewProj[15] + viewProj[14],
	}.Normalize()

	// Far plane
	c.frustumPlanes[PlaneFar] = nomath.Vec4{
		X: viewProj[3] - viewProj[2],
		Y: viewProj[7] - viewProj[6],
		Z: viewProj[11] - viewProj[10],
		W: viewProj[15] - viewProj[14],
	}.Normalize()

	c.DirtyFrustum = false
}

// IsVisible checks if a bounding box is visible in the frustum
func (c *PerspectiveCamera) IsVisible(box *nomath.BoundingBox) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.UpdateFrustumPlanes()
	center := box.Center()
	extents := box.Size().Multiply(0.5)

	for _, plane := range c.frustumPlanes {
		// Calculate signed distance from box center to plane
		dist := plane.X*center.X + plane.Y*center.Y + plane.Z*center.Z + plane.W

		// Calculate the effective radius of the box for this plane
		radius := extents.X*math.Abs(plane.X) +
			extents.Y*math.Abs(plane.Y) +
			extents.Z*math.Abs(plane.Z)

		// If the box is completely outside any plane, it's not visible
		if dist < -radius {
			return false
		}
	}
	return true
}

func (c *PerspectiveCamera) CacheMatrices() {
	c.UpdateFrustumPlanes()

	viewMatrix := c.GetViewMatrix()
	projectionMatrix := c.GetProjectionMatrix()
	viewProjMatrix := projectionMatrix.Multiply(viewMatrix)

	c.Scene.matrixMutex.Lock()
	defer c.Scene.matrixMutex.Unlock()

	c.Scene.cachedViewMatrix = viewMatrix
	c.Scene.cachedProjectionMatrix = projectionMatrix
	c.Scene.cachedViewProjMatrix = viewProjMatrix
	c.DirtyFrustum = false
}

func (c *PerspectiveCamera) Update() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.Transform.UpdateModelMatrix()
	c.CacheMatrices()

}
