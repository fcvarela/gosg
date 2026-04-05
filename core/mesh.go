package core

import (
	"sync/atomic"
	"unsafe"

	"github.com/fcvarela/gosg/gpu"
	"github.com/go-gl/mathgl/mgl64"
)

// PrimitiveType is a raster primitive type.
type PrimitiveType uint8

// Supported primitive types
const (
	PrimitiveTypeTriangles PrimitiveType = iota
	PrimitiveTypePoints
	PrimitiveTypeLines
)

var nextMeshID uint32

// Mesh holds geometry data backed by GPU buffers.
type Mesh struct {
	id             uint32
	name           string
	bounds         *AABB
	primitiveType  PrimitiveType
	indexCount     uint32
	positionBuffer gpu.Buffer
	normalBuffer   gpu.Buffer
	texCoordBuffer gpu.Buffer
	indexBuffer    gpu.Buffer
	instanceBuffer gpu.Buffer
	positionSize   uint64
	normalSize     uint64
	texCoordSize   uint64
	indexSize      uint64
}

// NewMesh creates a new empty mesh.
func NewMesh() *Mesh {
	m := &Mesh{
		id:     atomic.AddUint32(&nextMeshID, 1),
		bounds: NewAABB(),
	}
	// Pre-allocate instance buffer for instanced drawing
	if renderer != nil {
		m.instanceBuffer = renderer.device.CreateBuffer(
			uint64(MaxInstances*InstanceDataLen),
			gpu.BufferUsageVertex|gpu.BufferUsageCopyDst,
		)
	}
	return m
}

func (m *Mesh) SetPrimitiveType(t PrimitiveType) { m.primitiveType = t }
func (m *Mesh) SetName(name string)              { m.name = name }
func (m *Mesh) Name() string                     { return m.name }
func (m *Mesh) Bounds() *AABB                    { return m.bounds }

func (m *Mesh) SetPositions(positions []float32) {
	m.positionSize = uint64(len(positions) * 4)
	m.positionBuffer = renderer.device.CreateBuffer(m.positionSize, gpu.BufferUsageVertex|gpu.BufferUsageCopyDst)
	renderer.queue.WriteBuffer(m.positionBuffer, 0, unsafe.Pointer(&positions[0]), m.positionSize)

	// grow bounds
	for i := 0; i < len(positions); i += 3 {
		m.bounds.ExtendWithPoint(mgl64.Vec3{
			float64(positions[i+0]),
			float64(positions[i+1]),
			float64(positions[i+2]),
		})
	}
}

func (m *Mesh) SetNormals(normals []float32) {
	m.normalSize = uint64(len(normals) * 4)
	m.normalBuffer = renderer.device.CreateBuffer(m.normalSize, gpu.BufferUsageVertex|gpu.BufferUsageCopyDst)
	renderer.queue.WriteBuffer(m.normalBuffer, 0, unsafe.Pointer(&normals[0]), m.normalSize)
}

func (m *Mesh) SetTextureCoordinates(texcoords []float32) {
	m.texCoordSize = uint64(len(texcoords) * 4)
	m.texCoordBuffer = renderer.device.CreateBuffer(m.texCoordSize, gpu.BufferUsageVertex|gpu.BufferUsageCopyDst)
	renderer.queue.WriteBuffer(m.texCoordBuffer, 0, unsafe.Pointer(&texcoords[0]), m.texCoordSize)
}

func (m *Mesh) SetIndices(indices []uint16) {
	m.indexCount = uint32(len(indices))
	rawSize := uint64(len(indices) * 2)
	// wgpu requires buffer sizes and copy sizes aligned to 4 bytes
	m.indexSize = (rawSize + 3) &^ 3
	m.indexBuffer = renderer.device.CreateBuffer(m.indexSize, gpu.BufferUsageIndex|gpu.BufferUsageCopyDst)
	// Pad data to aligned size
	padded := make([]byte, m.indexSize)
	copy(padded, (*[1 << 30]byte)(unsafe.Pointer(&indices[0]))[:rawSize])
	renderer.queue.WriteBuffer(m.indexBuffer, 0, unsafe.Pointer(&padded[0]), m.indexSize)
}

func (m *Mesh) Draw(rp *RenderPass) {
	rp.SetVertexBuffer(0, m.positionBuffer, 0, m.positionSize)
	rp.SetVertexBuffer(1, m.normalBuffer, 0, m.normalSize)
	rp.SetVertexBuffer(2, m.texCoordBuffer, 0, m.texCoordSize)
	rp.SetIndexBuffer(m.indexBuffer, gpu.IndexFormatUint16, 0, m.indexSize)
	rp.DrawIndexed(m.indexCount, 1, 0, 0, 0)
}

func (m *Mesh) DrawInstanced(rp *RenderPass, instanceCount int, instanceData unsafe.Pointer) {
	dataSize := uint64(instanceCount * InstanceDataLen)
	renderer.queue.WriteBuffer(m.instanceBuffer, 0, instanceData, dataSize)
	rp.SetVertexBuffer(0, m.positionBuffer, 0, m.positionSize)
	rp.SetVertexBuffer(1, m.normalBuffer, 0, m.normalSize)
	rp.SetVertexBuffer(2, m.texCoordBuffer, 0, m.texCoordSize)
	rp.SetVertexBuffer(3, m.instanceBuffer, 0, dataSize)
	rp.SetIndexBuffer(m.indexBuffer, gpu.IndexFormatUint16, 0, m.indexSize)
	rp.DrawIndexed(m.indexCount, uint32(instanceCount), 0, 0, 0)
}

func (m *Mesh) ID() uint32 { return m.id }

var aabbMesh *Mesh

// NewScreenQuadMesh returns a mesh to be drawn by an orthographic projection camera.
func NewScreenQuadMesh(width, height float32) *Mesh {
	m := NewMesh()
	m.SetPrimitiveType(PrimitiveTypeTriangles)
	m.SetPositions([]float32{
		width * 0.0, height * 0.0, 0.0,
		width * 1.0, height * 0.0, 0.0,
		width * 1.0, height * 1.0, 0.0,
		width * 0.0, height * 1.0, 0.0,
	})
	m.SetNormals([]float32{0, 0, 0, 0, 1, 0, 1, 1, 0, 1, 0, 0})
	m.SetTextureCoordinates([]float32{0, 1, 0, 1, 1, 0, 1, 0, 0, 0, 0, 0})
	m.SetIndices([]uint16{0, 1, 2, 2, 3, 0})
	m.SetName("ScreenQuadMesh")
	return m
}

// AABBMesh returns a normalized cube centered at the origin.
func AABBMesh() *Mesh {
	if aabbMesh != nil {
		return aabbMesh
	}

	aabbMesh = NewMesh()
	aabbMesh.SetPrimitiveType(PrimitiveTypeLines)
	aabbMesh.SetPositions([]float32{
		-0.5, -0.5, -0.5, +0.5, -0.5, -0.5, +0.5, +0.5, -0.5, -0.5, +0.5, -0.5,
		-0.5, -0.5, +0.5, +0.5, -0.5, +0.5, +0.5, +0.5, +0.5, -0.5, +0.5, +0.5,
	})
	dummyNormals := []float32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	aabbMesh.SetNormals(dummyNormals)
	aabbMesh.SetTextureCoordinates(dummyNormals)
	aabbMesh.SetIndices([]uint16{0, 1, 1, 2, 2, 3, 3, 0, 4, 5, 5, 6, 6, 7, 7, 4, 0, 4, 1, 5, 2, 6, 3, 7})
	aabbMesh.SetName("AABB")

	return aabbMesh
}
