package core

import (
	"unsafe"

	"github.com/fcvarela/gosg/gpu"
)

// Texture holds a GPU texture, its view, sampler, and descriptor.
type Texture struct {
	texture    gpu.Texture
	view       gpu.TextureView
	sampler    gpu.Sampler
	descriptor TextureDescriptor
}

// Descriptor returns the descriptor used to create this texture.
func (t *Texture) Descriptor() TextureDescriptor {
	return t.descriptor
}

// Handle returns a pointer to this texture (used by ImGui).
func (t *Texture) Handle() unsafe.Pointer {
	return unsafe.Pointer(t)
}

// Lt is used for sorting textures.
func (t *Texture) Lt(other *Texture) bool {
	if other == nil {
		return false
	}
	return uintptr(unsafe.Pointer(t)) < uintptr(unsafe.Pointer(other))
}

// Gt is used for sorting textures.
func (t *Texture) Gt(other *Texture) bool {
	if other == nil {
		return false
	}
	return uintptr(unsafe.Pointer(t)) > uintptr(unsafe.Pointer(other))
}

// SetFilter recreates the sampler with the given filter mode.
func (t *Texture) SetFilter(f TextureFilter) {
	t.descriptor.Filter = f
	t.sampler = renderer.createSampler(t.descriptor)
}

// SetWrapMode recreates the sampler with the given wrap mode.
func (t *Texture) SetWrapMode(wm TextureWrapMode) {
	t.descriptor.WrapMode = wm
	t.sampler = renderer.createSampler(t.descriptor)
}

// TextureTarget specifies a texture target type
type TextureTarget int

const (
	TextureTarget1D TextureTarget = iota
	TextureTarget1DArray
	TextureTarget2D
	TextureTarget2DArray
	TextureTargetCubemapXPositive
	TextureTargetCubemapXNegative
	TextureTargetCubemapYPositive
	TextureTargetCubemapYNegative
	TextureTargetCubemapZPositive
	TextureTargetCubemapZNegative
)

// TextureFormat holds the texture component layout
type TextureFormat int

const (
	TextureFormatR TextureFormat = iota
	TextureFormatRG
	TextureFormatRGB
	TextureFormatRGBA
	TextureFormatDEPTH
)

// TextureSizedFormat specifies the format and size of a texture's components
type TextureSizedFormat int

const (
	TextureSizedFormatR8 TextureSizedFormat = iota
	TextureSizedFormatR16F
	TextureSizedFormatR32F
	TextureSizedFormatRG8
	TextureSizedFormatRG16F
	TextureSizedFormatRG32F
	TextureSizedFormatRGB8
	TextureSizedFormatRGB16F
	TextureSizedFormatRGB32F
	TextureSizedFormatRGBA8
	TextureSizedFormatRGBA16F
	TextureSizedFormatRGBA32F
	TextureSizedFormatDEPTH32F
)

// TextureComponentType specifies the texture component storage type
type TextureComponentType int

const (
	TextureComponentTypeUNSIGNEDBYTE TextureComponentType = iota
	TextureComponentTypeFLOAT
)

// TextureWrapMode specifies the type of wrap around
type TextureWrapMode int

const (
	TextureWrapModeClampEdge TextureWrapMode = iota
	TextureWrapModeClampBorder
	TextureWrapModeRepeat
)

// TextureFilter specifies the type of interpolation
type TextureFilter int

const (
	TextureFilterNearest TextureFilter = iota
	TextureFilterLinear
	TextureFilterMipmapLinear
)

// TextureDescriptor contains the full description of a texture
type TextureDescriptor struct {
	Width         uint32
	Height        uint32
	Mipmaps       bool
	Target        TextureTarget
	Format        TextureFormat
	SizedFormat   TextureSizedFormat
	ComponentType TextureComponentType
	Filter        TextureFilter
	WrapMode      TextureWrapMode
}
