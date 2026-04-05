package core

import "github.com/go-gl/mathgl/mgl32"

// Material contains material properties for a specific drawable.
type Material struct {
	uniforms       map[string]*Uniform
	uniformBuffers map[string]*UniformBuffer
	textures       map[string]*Texture
	instanceData   [4]mgl32.Vec4
	sortKey        uint64
}

// NewMaterial returns a new Material.
func NewMaterial() Material {
	return Material{
		uniforms:       make(map[string]*Uniform),
		uniformBuffers: make(map[string]*UniformBuffer),
		textures:       make(map[string]*Texture),
	}
}

// Uniforms returns the uniform map.
func (m *Material) Uniforms() map[string]*Uniform {
	return m.uniforms
}

// Uniform returns the uniform with the given name, creating it if needed.
func (m *Material) Uniform(name string) *Uniform {
	if _, ok := m.uniforms[name]; !ok {
		m.uniforms[name] = NewUniform()
	}
	return m.uniforms[name]
}

// SetTexture sets the texture named `name` and recomputes the sort key.
func (m *Material) SetTexture(name string, t *Texture) {
	m.textures[name] = t
	m.recomputeSortKey()
}

// Textures returns the textures map.
func (m *Material) Textures() map[string]*Texture {
	return m.textures
}

// Texture returns the texture with the specified name.
func (m *Material) Texture(name string) *Texture {
	return m.textures[name]
}

// UniformBuffer returns the uniform buffer with the given name, creating it if needed.
func (m *Material) UniformBuffer(name string) *UniformBuffer {
	if _, ok := m.uniformBuffers[name]; !ok {
		m.uniformBuffers[name] = NewUniformBuffer()
	}
	return m.uniformBuffers[name]
}

// UniformBuffers returns the uniform buffers map.
func (m *Material) UniformBuffers() map[string]*UniformBuffer {
	return m.uniformBuffers
}

// InstanceData returns the per-instance data.
func (m *Material) InstanceData() [4]mgl32.Vec4 {
	return m.instanceData
}

// SetInstanceDataField sets a per-instance data field.
func (m *Material) SetInstanceDataField(index int, v mgl32.Vec4) {
	m.instanceData[index] = v
}

// SortKey returns the material's pre-computed sort key.
func (m *Material) SortKey() uint64 {
	return m.sortKey
}

func (m *Material) recomputeSortKey() {
	var key uint64
	for _, t := range m.textures {
		if t != nil {
			key ^= uint64(t.id) * 0x9e3779b97f4a7c15
		}
	}
	m.sortKey = key
}
