package opengl

import (
	"C"
	"unsafe"

	"github.com/fcvarela/gosg/core"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl64"
)

type bufferType uint8

const (
	positionBuffer bufferType = iota
	normalBuffer
	texCoordBuffer
	indexBuffer
	instanceDataBuffer
)

type buffers struct {
	vao           uint32
	buffers       []uint32
	bufferOffsets []int
}

var (
	currentVAO = uint32(0)
)

func (b *buffers) addData(target uint32, buffer bufferType, datalen int, buf unsafe.Pointer) {
	if b.bufferOffsets[buffer] == 0 {
		gl.BindBuffer(target, b.buffers[buffer])
		gl.BufferData(target, datalen, buf, gl.STATIC_DRAW)
	} else {
		// cpu copy: get existing data, alloc new space for everything, add old data, new data
		cpuBuf := make([]byte, b.bufferOffsets[buffer])
		gl.BindBuffer(target, b.buffers[buffer])
		gl.GetBufferSubData(target, 0, b.bufferOffsets[buffer], unsafe.Pointer(&cpuBuf[0]))
		gl.BufferData(target, datalen+b.bufferOffsets[buffer], nil, gl.STATIC_DRAW)
		gl.BufferSubData(target, 0, b.bufferOffsets[buffer], unsafe.Pointer(&cpuBuf[0]))
		gl.BufferSubData(target, b.bufferOffsets[buffer], datalen, buf)
	}
	b.bufferOffsets[buffer] += datalen
}

func newBuffers() *buffers {
	bf := &buffers{}

	// create VAO
	gl.GenVertexArrays(1, &bf.vao)

	// create buffers
	bf.buffers = make([]uint32, instanceDataBuffer+1)
	bf.bufferOffsets = make([]int, instanceDataBuffer+1)

	// initialize gl buffer handles
	gl.GenBuffers(int32(len(bf.buffers)), &bf.buffers[0])

	// init attributes
	bindVAO(bf.vao)

	// position
	gl.BindBuffer(gl.ARRAY_BUFFER, bf.buffers[positionBuffer])
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)

	// normal
	gl.BindBuffer(gl.ARRAY_BUFFER, bf.buffers[normalBuffer])
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 0, nil)

	// texcoord
	gl.BindBuffer(gl.ARRAY_BUFFER, bf.buffers[texCoordBuffer])
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(2, 3, gl.FLOAT, false, 0, nil)

	// indices
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, bf.buffers[indexBuffer])

	// model matrices, prealloc for core.MaxInstances instances of 4x4 matrices with 4 bytes per float (this is our hard-max)
	gl.BindBuffer(gl.ARRAY_BUFFER, bf.buffers[instanceDataBuffer])
	gl.BufferData(gl.ARRAY_BUFFER, core.MaxInstances*core.InstanceDataLen, nil, gl.STREAM_DRAW)

	gl.EnableVertexAttribArray(3)
	gl.VertexAttribPointer(3, 4, gl.FLOAT, false, core.InstanceDataLen, gl.PtrOffset(0*16))
	gl.VertexAttribDivisor(3, 1)

	gl.EnableVertexAttribArray(4)
	gl.VertexAttribPointer(4, 4, gl.FLOAT, false, core.InstanceDataLen, gl.PtrOffset(1*16))
	gl.VertexAttribDivisor(4, 1)

	gl.EnableVertexAttribArray(5)
	gl.VertexAttribPointer(5, 4, gl.FLOAT, false, core.InstanceDataLen, gl.PtrOffset(2*16))
	gl.VertexAttribDivisor(5, 1)

	gl.EnableVertexAttribArray(6)
	gl.VertexAttribPointer(6, 4, gl.FLOAT, false, core.InstanceDataLen, gl.PtrOffset(3*16))
	gl.VertexAttribDivisor(6, 1)

	// modelviewprojection matrices
	gl.EnableVertexAttribArray(7)
	gl.VertexAttribPointer(7, 4, gl.FLOAT, false, core.InstanceDataLen, gl.PtrOffset(4*16))
	gl.VertexAttribDivisor(7, 1)

	gl.EnableVertexAttribArray(8)
	gl.VertexAttribPointer(8, 4, gl.FLOAT, false, core.InstanceDataLen, gl.PtrOffset(5*16))
	gl.VertexAttribDivisor(8, 1)

	gl.EnableVertexAttribArray(9)
	gl.VertexAttribPointer(9, 4, gl.FLOAT, false, core.InstanceDataLen, gl.PtrOffset(6*16))
	gl.VertexAttribDivisor(9, 1)

	gl.EnableVertexAttribArray(10)
	gl.VertexAttribPointer(10, 4, gl.FLOAT, false, core.InstanceDataLen, gl.PtrOffset(7*16))
	gl.VertexAttribDivisor(10, 1)

	// custom data
	gl.EnableVertexAttribArray(11)
	gl.VertexAttribPointer(11, 4, gl.FLOAT, false, core.InstanceDataLen, gl.PtrOffset(8*16))
	gl.VertexAttribDivisor(11, 1)

	gl.EnableVertexAttribArray(12)
	gl.VertexAttribPointer(12, 4, gl.FLOAT, false, core.InstanceDataLen, gl.PtrOffset(9*16))
	gl.VertexAttribDivisor(12, 1)

	gl.EnableVertexAttribArray(13)
	gl.VertexAttribPointer(13, 4, gl.FLOAT, false, core.InstanceDataLen, gl.PtrOffset(10*16))
	gl.VertexAttribDivisor(13, 1)

	gl.EnableVertexAttribArray(14)
	gl.VertexAttribPointer(14, 4, gl.FLOAT, false, core.InstanceDataLen, gl.PtrOffset(11*16))
	gl.VertexAttribDivisor(14, 1)

	bindVAO(0)
	return bf
}

var (
	sharedBuffers *buffers
)

// Mesh implements the core.Mesh interface
type Mesh struct {
	buffers           *buffers
	indexcount        int32
	indexOffset       int32
	indexBufferOffset int32
	name              string
	bounds            *core.AABB
	primitiveType     uint32
}

// IMGUIMesh implements the core.IMGUIMesh interface
type IMGUIMesh struct {
	*Mesh
}

// NewMesh implements the core.RenderSystem interface
func (r *RenderSystem) NewMesh() core.Mesh {
	m := Mesh{}

	m.bounds = core.NewAABB()
	m.buffers = sharedBuffers
	return &m
}

func bindVAO(vao uint32) {
	if currentVAO != vao {
		gl.BindVertexArray(vao)
		currentVAO = vao
	}
}

// SetPrimitiveType implements the core.Mesh interface
func (m *Mesh) SetPrimitiveType(t core.PrimitiveType) {
	switch t {
	case core.PrimitiveTypeTriangles:
		m.primitiveType = gl.TRIANGLES
	case core.PrimitiveTypeLines:
		m.primitiveType = gl.LINES
	case core.PrimitiveTypePoints:
		m.primitiveType = gl.POINTS
		gl.Enable(gl.PROGRAM_POINT_SIZE)
		gl.PointParameteri(gl.POINT_SPRITE_COORD_ORIGIN, gl.LOWER_LEFT)
	}
}

// Bounds implements the core.Mesh interface
func (m *Mesh) Bounds() *core.AABB {
	return m.bounds
}

// SetName implements the core.Mesh interface
func (m *Mesh) SetName(name string) {
	m.name = name
}

// Name implements the core.Mesh interface
func (m *Mesh) Name() string {
	return m.name
}

// SetPositions implements the core.Mesh interface
func (m *Mesh) SetPositions(positions []float32) {
	// the index offset is equal to the number of primitives already in the position buffer
	m.indexOffset = int32(m.buffers.bufferOffsets[positionBuffer]) / (4 * 3)
	m.buffers.addData(gl.ARRAY_BUFFER, positionBuffer, len(positions)*4, gl.Ptr(positions))

	// grow our bounds
	for i := 0; i < len(positions); i += 3 {
		m.bounds.ExtendWithPoint(
			mgl64.Vec3{
				float64(positions[i+0]),
				float64(positions[i+1]),
				float64(positions[i+2])})
	}
}

// SetNormals implements the core.Mesh interface
func (m *Mesh) SetNormals(normals []float32) {
	m.buffers.addData(gl.ARRAY_BUFFER, normalBuffer, len(normals)*4, gl.Ptr(normals))
}

// SetTextureCoordinates implements the core.Mesh interface
func (m *Mesh) SetTextureCoordinates(texcoords []float32) {
	m.buffers.addData(gl.ARRAY_BUFFER, texCoordBuffer, len(texcoords)*4, gl.Ptr(texcoords))
}

// SetIndices implements the core.Mesh interface
func (m *Mesh) SetIndices(indices []uint16) {
	m.indexBufferOffset = int32(m.buffers.bufferOffsets[indexBuffer])
	m.indexcount = int32(len(indices))
	m.buffers.addData(gl.ELEMENT_ARRAY_BUFFER, indexBuffer, len(indices)*2, gl.Ptr(indices))
}

// Draw implements the core.Mesh interface
func (m *Mesh) Draw() {
	bindVAO(m.buffers.vao)
	gl.DrawElementsBaseVertex(
		m.primitiveType, m.indexcount, gl.UNSIGNED_SHORT,
		gl.PtrOffset(int(m.indexBufferOffset)), m.indexOffset)
}

// DrawInstanced implements the core.Mesh interface
func (m *Mesh) DrawInstanced(instanceCount int, instanceData unsafe.Pointer) {
	// copy per instance data and orphan last buffer
	gl.BindBuffer(gl.ARRAY_BUFFER, m.buffers.buffers[instanceDataBuffer])
	gl.BufferData(gl.ARRAY_BUFFER, core.MaxInstances*core.InstanceDataLen, nil, gl.STREAM_DRAW)
	gl.BufferData(gl.ARRAY_BUFFER, core.MaxInstances*core.InstanceDataLen, instanceData, gl.STREAM_DRAW)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	bindVAO(m.buffers.vao)
	gl.DrawElementsInstancedBaseVertex(
		m.primitiveType, m.indexcount, gl.UNSIGNED_SHORT,
		gl.PtrOffset(int(m.indexBufferOffset)), int32(instanceCount), m.indexOffset)
}

// Lt implements the core.Mesh interface
func (m *Mesh) Lt(other core.Mesh) bool {
	return m.buffers.vao < other.(*Mesh).buffers.vao
}

// Gt implements the core.Mesh interface
func (m *Mesh) Gt(other core.Mesh) bool {
	return m.buffers.vao > other.(*Mesh).buffers.vao
}
