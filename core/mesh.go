package core

import (
	"unsafe"
)

// PrimitiveType is a raster primitive type.
type PrimitiveType uint8

// Supported primitive types
const (
	PrimitiveTypeTriangles PrimitiveType = iota
	PrimitiveTypePoints
	PrimitiveTypeLines
)

// Mesh is an interface which wraps handling of geometry.
type Mesh interface {
	SetPrimitiveType(PrimitiveType)

	SetPositions(positions []float32)
	SetNormals(normals []float32)
	SetTextureCoordinates(coordinates []float32)
	SetIndices(indices []uint16)
	SetInstanceCount(count int)
	SetModelMatrices(matrices unsafe.Pointer)

	SetName(name string)
	Name() string

	Draw()

	Bounds() *AABB

	Lt(Mesh) bool
	Gt(Mesh) bool
}

// IMGUIMesh is an interface which wraps a Mesh used for IMGUI primitives.
type IMGUIMesh interface {
	Mesh
}

var (
	aabbMesh Mesh
)

// NewScreenQuadMesh returns a mesh to be drawn by an orthographic projection camera.
func NewScreenQuadMesh(width, height float32) Mesh {
	positions := []float32{
		width * 0.0, height * 0.0, 0.0,
		width * 1.0, height * 0.0, 0.0,
		width * 1.0, height * 1.0, 0.0,
		width * 0.0, height * 1.0, 0.0,
	}

	normals := []float32{
		0.0, 0.0, 0.0,
		0.0, 1.0, 0.0,
		1.0, 1.0, 0.0,
		1.0, 0.0, 0.0,
	}

	tcoords := []float32{
		0.0, 1.0, 0.0,
		1.0, 1.0, 0.0,
		1.0, 0.0, 0.0,
		0.0, 0.0, 0.0,
	}

	indices := []uint16{0, 1, 2, 2, 3, 0}

	m := renderSystem.NewMesh()
	m.SetPrimitiveType(PrimitiveTypeTriangles)
	m.SetPositions(positions)
	m.SetNormals(normals)
	m.SetTextureCoordinates(tcoords)
	m.SetIndices(indices)
	m.SetName("ScreenQuadMesh")
	return m
}

// AABBMesh returns a normalized cube centered at the origin. This is
// used to draw bounding boxes by translating and scaling it according to node bounds.
func AABBMesh() Mesh {
	if aabbMesh != nil {
		return aabbMesh
	}

	positions := []float32{
		-0.5, -0.5, -0.5,
		+0.5, -0.5, -0.5,
		+0.5, +0.5, -0.5,
		-0.5, +0.5, -0.5,
		-0.5, -0.5, +0.5,
		+0.5, -0.5, +0.5,
		+0.5, +0.5, +0.5,
		-0.5, +0.5, +0.5,
	}

	indices := []uint16{
		0, 1,
		1, 2,
		2, 3,
		3, 0,
		4, 5,
		5, 6,
		6, 7,
		7, 4,
		0, 4,
		1, 5,
		2, 6,
		3, 7}

	aabbMesh = renderSystem.NewMesh()
	aabbMesh.SetPrimitiveType(PrimitiveTypeLines)
	aabbMesh.SetPositions(positions)
	aabbMesh.SetNormals(positions)
	aabbMesh.SetTextureCoordinates(positions)
	aabbMesh.SetIndices(indices)
	aabbMesh.SetInstanceCount(1)
	aabbMesh.SetName("AABB")

	return aabbMesh
}
