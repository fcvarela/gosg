package core

import "github.com/go-gl/mathgl/mgl32"

// Descriptors contains material properties for a specific drawable
type Descriptors struct {
	uniforms       map[string]*Uniform
	uniformBuffers map[string]*UniformBuffer
	textures       map[string]*Texture
	instanceData   [4]mgl32.Vec4
}

// NewDescriptors returns a new Descriptors
func NewDescriptors() Descriptors {
	return Descriptors{
		uniforms:       make(map[string]*Uniform),
		uniformBuffers: make(map[string]*UniformBuffer),
		textures:       make(map[string]*Texture),
	}
}

// Uniforms returns the uniform map.
func (s *Descriptors) Uniforms() map[string]*Uniform {
	return s.uniforms
}

// Uniform returns the uniform with the given name, creating it if needed.
func (s *Descriptors) Uniform(name string) *Uniform {
	if _, ok := s.uniforms[name]; !ok {
		s.uniforms[name] = NewUniform()
	}
	return s.uniforms[name]
}

// SetTexture sets the texture named `name`.
func (s *Descriptors) SetTexture(name string, t *Texture) {
	s.textures[name] = t
}

// Textures returns the textures map.
func (s *Descriptors) Textures() map[string]*Texture {
	return s.textures
}

// Texture returns the texture with the specified name.
func (s *Descriptors) Texture(name string) *Texture {
	return s.textures[name]
}

// UniformBuffer returns the uniform buffer with the given name, creating it if needed.
func (s *Descriptors) UniformBuffer(name string) *UniformBuffer {
	if _, ok := s.uniformBuffers[name]; !ok {
		s.uniformBuffers[name] = NewUniformBuffer()
	}
	return s.uniformBuffers[name]
}

// UniformBuffers returns the uniform buffers map.
func (s *Descriptors) UniformBuffers() map[string]*UniformBuffer {
	return s.uniformBuffers
}

// InstanceData returns the per-instance data.
func (s *Descriptors) InstanceData() [4]mgl32.Vec4 {
	return s.instanceData
}

// SetInstanceDataField sets a per-instance data field.
func (s *Descriptors) SetInstanceDataField(index int, v mgl32.Vec4) {
	s.instanceData[index] = v
}
