package core

import "unsafe"

// Uniform is an interface which wraps program Uniforms. This will be replaced by PushConstants.
type Uniform interface {

	// Value returns the Uniform's value
	Value() interface{}

	// Set sets the Uniform's value
	Set(interface{})

	// Copy returns a copy of the Uniform
	Copy() Uniform
}

// UniformBuffer is an interface which wraps program UniformBuffers. This will be replaced by ConstantBuffers.
type UniformBuffer interface {
	Set(unsafe.Pointer, int)

	// Lt is used for sorting
	Lt(UniformBuffer) bool

	// Gt
	Gt(UniformBuffer) bool
}
