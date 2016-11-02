package core

import "unsafe"

// TextureTarget specifies a texture target type (1D, 2D, 2DArray, Cubemap, etc)
type TextureTarget int

// TextureTargetXXX are the different texture types
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

// These are the several supported texture component layouts
const (
	TextureFormatR TextureFormat = iota
	TextureFormatRG
	TextureFormatRGB
	TextureFormatRGBA
	TextureFormatDEPTH
)

// TextureSizedFormat specified the format and size of a texture's components
type TextureSizedFormat int

// These are the several supportex texture component sizes
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

// These are the supported texture component storate types
const (
	TextureComponentTypeUNSIGNEDBYTE TextureComponentType = iota
	TextureComponentTypeFLOAT
)

// TextureWrapMode specifies the type of wrap around a sampler of this texture will use
type TextureWrapMode int

// These are the supported texture wrap modes
const (
	TextureWrapModeClampEdge TextureWrapMode = iota
	TextureWrapModeClampBorder
	TextureWrapModeRepeat
)

// TextureFilter specifies the type of interpolation a sampler of this texture will use
type TextureFilter int

// These are the supported texture filtering modes
const (
	TextureFilterNearest TextureFilter = iota
	TextureFilterLinear
	TextureFilterMipmapLinear
)

// TextureDescriptor contains the full description of a texture and its sampling parameters
// It is used as input to texture creation functions and at runtime inside rendersystems
// to setup samplers and memory allocation
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

// Texture is an interface which wraps both a texture and settings for samplers sampling it
type Texture interface {
	Descriptor() TextureDescriptor

	Handle() unsafe.Pointer

	// Lt is used for sorting
	Lt(Texture) bool

	// Gt
	Gt(Texture) bool

	// SetFilter
	SetFilter(TextureFilter)

	// SetWrapMode
	SetWrapMode(TextureWrapMode)
}
