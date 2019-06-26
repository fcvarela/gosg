package opengl

import (
	"bytes"
	"image"
	"image/draw"
	_ "image/jpeg" // registers jpeg handler
	_ "image/png"  // registers png handler
	"math"
	"runtime"
	"unsafe"

	"github.com/fcvarela/gosg/core"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/golang/glog"
	_ "golang.org/x/image/bmp" // registers bmp handler
)

// Texture is an OpenGL texture container
type Texture struct {
	id         uint32
	descriptor core.TextureDescriptor
}

// Descriptor returns the descriptor used to create this texture. Note that overriding filter and wrapMode
// will _not_ change the original descriptor used.
func (t *Texture) Descriptor() core.TextureDescriptor {
	return t.descriptor
}

// Handle implements the core.Texture interface
func (t *Texture) Handle() unsafe.Pointer {
	return unsafe.Pointer(t)
}

// Lt implements the core.Texture interface
func (t *Texture) Lt(other core.Texture) bool {
	if ot, ok := other.(*Texture); ok {
		return t.id < ot.id
	}
	return true
}

// Gt implements the core.Texture interface
func (t *Texture) Gt(other core.Texture) bool {
	if ot, ok := other.(*Texture); ok {
		return t.id > ot.id
	}
	return false
}

func textureCleanup(t *Texture) {
	glog.Info("Deleting texture: ", t.id)
}

// NewTextureFromImageData implements the core.RenderSystem interface
func (rs *RenderSystem) NewTextureFromImageData(data []byte, descriptor core.TextureDescriptor) core.Texture {
	if data == nil {
		glog.Fatal("Cannot read texture...")
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		glog.Warning("Cannot decode texture image: ", err)
	}

	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		panic("Unsupported stride")
	}
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	descriptor.Width = uint32(rgba.Rect.Size().X)
	descriptor.Height = uint32(rgba.Rect.Size().Y)
	descriptor.Target = core.TextureTarget2D
	descriptor.Format = core.TextureFormatRGBA
	descriptor.SizedFormat = core.TextureSizedFormatRGBA8
	descriptor.ComponentType = core.TextureComponentTypeUNSIGNEDBYTE

	return rs.NewTexture(descriptor, rgba.Pix)
}

// NewTexture implements the core.RenderSystem interface
func (rs *RenderSystem) NewTexture(d core.TextureDescriptor, data []byte) core.Texture {
	var target = textureTargetFromCore(d.Target)

	// sampling filter
	var minFilter, magFilter = textureFilterFromCore(d.Filter)

	// sampling wrap mode
	var wrapMode = textureWrapModeFromCore(d.WrapMode)

	// internal format
	var internalFormat uint32
	switch d.SizedFormat {
	case core.TextureSizedFormatR8:
		internalFormat = gl.R8
	case core.TextureSizedFormatR16F:
		internalFormat = gl.R16F
	case core.TextureSizedFormatR32F:
		internalFormat = gl.R32F
	case core.TextureSizedFormatRG8:
		internalFormat = gl.RG8
	case core.TextureSizedFormatRG16F:
		internalFormat = gl.RG16F
	case core.TextureSizedFormatRG32F:
		internalFormat = gl.RG32F
	case core.TextureSizedFormatRGB8:
		internalFormat = gl.RGB8
	case core.TextureSizedFormatRGB16F:
		internalFormat = gl.RGB16F
	case core.TextureSizedFormatRGB32F:
		internalFormat = gl.RGB32F
	case core.TextureSizedFormatRGBA8:
		internalFormat = gl.RGBA8
	case core.TextureSizedFormatRGBA16F:
		internalFormat = gl.RGBA16F
	case core.TextureSizedFormatRGBA32F:
		internalFormat = gl.RGBA32F
	case core.TextureSizedFormatDEPTH32F:
		internalFormat = gl.DEPTH_COMPONENT32F
	default:
		glog.Fatalf("Texture sized format %v not implemented", d.SizedFormat)
	}

	// format of data we're passing
	var format uint32
	switch d.Format {
	case core.TextureFormatR:
		format = gl.RED
	case core.TextureFormatRG:
		format = gl.RG
	case core.TextureFormatRGB:
		format = gl.RGB
	case core.TextureFormatRGBA:
		format = gl.RGBA
	case core.TextureFormatDEPTH:
		format = gl.DEPTH
	default:
		glog.Fatalf("Texture format %v not implemented", d.Format)
	}

	// component type of each channel px
	var componentType uint32
	switch d.ComponentType {
	case core.TextureComponentTypeUNSIGNEDBYTE:
		componentType = gl.UNSIGNED_BYTE
	case core.TextureComponentTypeFLOAT:
		componentType = gl.FLOAT
	default:
		glog.Fatalf("Component type %v not implemented", d.Format)
	}

	// mipmaps
	var mipmapCount = 0
	if d.Mipmaps {
		mipmapCount = int(math.Log2(float64(d.Width)))
		if d.Height < d.Width {
			mipmapCount = int(math.Log2(float64(d.Height)))
		}
	}
	mipmapCount++

	// bind the target type
	texture := uint32(0)
	gl.GenTextures(1, &texture)
	gl.BindTexture(target, texture)

	// set filtering
	gl.TexParameteri(target, gl.TEXTURE_MIN_FILTER, minFilter)
	gl.TexParameteri(target, gl.TEXTURE_MAG_FILTER, magFilter)

	// set wrap mode
	gl.TexParameteri(target, gl.TEXTURE_WRAP_S, wrapMode)
	gl.TexParameteri(target, gl.TEXTURE_WRAP_T, wrapMode)

	// allocate memory
	gl.TexStorage2D(target, int32(mipmapCount), internalFormat, int32(d.Width), int32(d.Height))

	// set data
	// this texture's storage has already been allocated, just copy the data
	if data != nil {
		gl.TexSubImage2D(target, 0, 0, 0, int32(d.Width), int32(d.Height), format, componentType, gl.Ptr(data))

		if d.Mipmaps {
			gl.GenerateMipmap(target)
		}
	}

	var t = &Texture{texture, d}
	runtime.SetFinalizer(t, textureCleanup)
	return t
}

func textureTargetFromCore(t core.TextureTarget) uint32 {
	var target uint32

	switch t {
	case core.TextureTarget2D:
		target = gl.TEXTURE_2D
	default:
		glog.Fatalf("Texture target %v not implemented: ", t)
	}

	return target
}

func textureFilterFromCore(tf core.TextureFilter) (minFilter, magFilter int32) {
	switch tf {
	case core.TextureFilterMipmapLinear:
		minFilter, magFilter = gl.LINEAR_MIPMAP_LINEAR, gl.LINEAR
	case core.TextureFilterLinear:
		minFilter, magFilter = gl.LINEAR, gl.LINEAR
	case core.TextureFilterNearest:
		minFilter, magFilter = gl.NEAREST, gl.NEAREST
	default:
		glog.Fatalf("Texture filter %v not implemented: ", tf)
	}

	return
}

func textureWrapModeFromCore(twm core.TextureWrapMode) int32 {
	var wrapMode int32

	switch twm {
	case core.TextureWrapModeClampBorder:
		wrapMode = gl.CLAMP_TO_BORDER
	case core.TextureWrapModeClampEdge:
		wrapMode = gl.CLAMP_TO_EDGE
	case core.TextureWrapModeRepeat:
		wrapMode = gl.REPEAT
	default:
		glog.Fatalf("Texture wrap mode %v not implemented: ", twm)
	}

	return wrapMode
}

// SetFilter sets this texture filtering mode
func (t *Texture) SetFilter(f core.TextureFilter) {
	var target = textureTargetFromCore(t.descriptor.Target)
	var minFilter, magFilter = textureFilterFromCore(f)

	gl.BindTexture(target, t.id)
	gl.TexParameteri(target, gl.TEXTURE_MIN_FILTER, minFilter)
	gl.TexParameteri(target, gl.TEXTURE_MAG_FILTER, magFilter)
}

// SetWrapMode sets this texture's wrap mode
func (t *Texture) SetWrapMode(wm core.TextureWrapMode) {
	var target = textureTargetFromCore(t.descriptor.Target)
	var wrapMode = textureWrapModeFromCore(wm)

	gl.BindTexture(target, t.id)
	gl.TexParameteri(target, gl.TEXTURE_WRAP_S, wrapMode)
	gl.TexParameteri(target, gl.TEXTURE_WRAP_T, wrapMode)
}
