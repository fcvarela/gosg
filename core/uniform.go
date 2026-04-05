package core

import (
	"unsafe"

	"github.com/fcvarela/gosg/gpu"
)

// Uniform holds a shader uniform value.
type Uniform struct {
	value any
}

// NewUniform creates a new empty uniform.
func NewUniform() *Uniform {
	return &Uniform{}
}

// Value returns the Uniform's value.
func (u *Uniform) Value() any {
	return u.value
}

// Set sets the Uniform's value.
func (u *Uniform) Set(value any) {
	u.value = value
}

// Copy returns a copy of the Uniform.
func (u *Uniform) Copy() *Uniform {
	return &Uniform{u.value}
}

// UniformBuffer holds a GPU uniform buffer.
type UniformBuffer struct {
	buffer gpu.Buffer
	size   uint64
}

// NewUniformBuffer creates a new empty uniform buffer.
func NewUniformBuffer() *UniformBuffer {
	return &UniformBuffer{}
}

// Set uploads data to the uniform buffer, creating/resizing as needed.
func (ub *UniformBuffer) Set(data unsafe.Pointer, dataLen int) {
	size := uint64(dataLen)
	if ub.size < size {
		ub.buffer.Release()
		ub.buffer = renderer.device.CreateBuffer(size, gpu.BufferUsageUniform|gpu.BufferUsageCopyDst)
		ub.size = size
	}
	renderer.queue.WriteBuffer(ub.buffer, 0, data, size)
}

// Lt is used for sorting.
func (ub *UniformBuffer) Lt(other *UniformBuffer) bool {
	if other == nil {
		return false
	}
	return uintptr(unsafe.Pointer(ub)) < uintptr(unsafe.Pointer(other))
}

// Gt is used for sorting.
func (ub *UniformBuffer) Gt(other *UniformBuffer) bool {
	if other == nil {
		return false
	}
	return uintptr(unsafe.Pointer(ub)) > uintptr(unsafe.Pointer(other))
}
