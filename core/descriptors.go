package core

import "github.com/go-gl/mathgl/mgl32"

// Descriptors contains material properties for a specific drawable
type Descriptors struct {
	uniforms       map[string]Uniform
	uniformBuffers map[string]UniformBuffer
	textures       map[string]Texture
	custom1        mgl32.Vec4
	custom2        mgl32.Vec4
	custom3        mgl32.Vec4
	custom4        mgl32.Vec4
}

// NewDescriptors returns a new MaterialData
func NewDescriptors() Descriptors {
	s := Descriptors{
		make(map[string]Uniform),
		make(map[string]UniformBuffer),
		make(map[string]Texture),
		mgl32.Vec4{},
		mgl32.Vec4{},
		mgl32.Vec4{},
		mgl32.Vec4{},
	}
	return s
}

// Uniforms returns the state's uniform map.
func (s *Descriptors) Uniforms() map[string]Uniform {
	return s.uniforms
}

// Uniform returns the uniform with the given name.
func (s *Descriptors) Uniform(name string) Uniform {
	_, ok := s.uniforms[name]
	if ok == false {
		s.uniforms[name] = renderSystem.NewUniform()
	}
	return s.uniforms[name]
}

// SetTexture sets the material's texture named `name` to the provided texture
func (s *Descriptors) SetTexture(name string, t Texture) {
	s.textures[name] = t
}

// Textures returns the material's textures
func (s *Descriptors) Textures() map[string]Texture {
	return s.textures
}

// Texture returns the material texture with the specified name
func (s *Descriptors) Texture(name string) Texture {
	return s.textures[name]
}

// UniformBuffer returns the uniform buffer with the given name
func (s *Descriptors) UniformBuffer(name string) UniformBuffer {
	_, ok := s.uniformBuffers[name]
	if !ok {
		s.uniformBuffers[name] = renderSystem.NewUniformBuffer()
	}
	return s.uniformBuffers[name]
}

// UniformBuffers returns the state's uniform buffers
func (s *Descriptors) UniformBuffers() map[string]UniformBuffer {
	return s.uniformBuffers
}
