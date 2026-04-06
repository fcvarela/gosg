// Package gpu provides thin Go bindings for wgpu-native (WebGPU C API).
// Only the subset needed by gosg is wrapped.
package gpu

/*
#cgo CFLAGS: -I/opt/homebrew/include
#cgo LDFLAGS: -L/opt/homebrew/lib -lwgpu_native -framework Metal -framework QuartzCore -framework Foundation

#include <stdlib.h>
#include <webgpu.h>
#include <wgpu.h>

// Callback trampolines (called from C, dispatch to Go)
extern void goRequestAdapterCallback(WGPURequestAdapterStatus status, WGPUAdapter adapter, WGPUStringView message, void *userdata1, void *userdata2);
extern void goRequestDeviceCallback(WGPURequestDeviceStatus status, WGPUDevice device, WGPUStringView message, void *userdata1, void *userdata2);
*/
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"unsafe"
)

// Handle types — thin wrappers around C opaque pointers.

type Instance struct{ ref C.WGPUInstance }
type Adapter struct{ ref C.WGPUAdapter }
type Device struct{ ref C.WGPUDevice }
type Queue struct{ ref C.WGPUQueue }
type Surface struct{ ref C.WGPUSurface }
type Buffer struct{ ref C.WGPUBuffer }
type Texture struct{ ref C.WGPUTexture }
type TextureView struct{ ref C.WGPUTextureView }
type Sampler struct{ ref C.WGPUSampler }
type ShaderModule struct{ ref C.WGPUShaderModule }
type BindGroupLayout struct{ ref C.WGPUBindGroupLayout }
type BindGroup struct{ ref C.WGPUBindGroup }
type PipelineLayout struct{ ref C.WGPUPipelineLayout }
type RenderPipeline struct{ ref C.WGPURenderPipeline }
type CommandEncoder struct{ ref C.WGPUCommandEncoder }
type CommandBuffer struct{ ref C.WGPUCommandBuffer }
type RenderPassEncoder struct{ ref C.WGPURenderPassEncoder }

// Enums — only the values we need.

type TextureFormat uint32

const (
	TextureFormatBGRA8Unorm    TextureFormat = C.WGPUTextureFormat_BGRA8Unorm
	TextureFormatRGBA8Unorm    TextureFormat = C.WGPUTextureFormat_RGBA8Unorm
	TextureFormatR8Unorm       TextureFormat = C.WGPUTextureFormat_R8Unorm
	TextureFormatR16Float      TextureFormat = C.WGPUTextureFormat_R16Float
	TextureFormatR32Float      TextureFormat = C.WGPUTextureFormat_R32Float
	TextureFormatRG8Unorm      TextureFormat = C.WGPUTextureFormat_RG8Unorm
	TextureFormatRG16Float     TextureFormat = C.WGPUTextureFormat_RG16Float
	TextureFormatRG32Float     TextureFormat = C.WGPUTextureFormat_RG32Float
	TextureFormatRGBA16Float   TextureFormat = C.WGPUTextureFormat_RGBA16Float
	TextureFormatRGBA32Float   TextureFormat = C.WGPUTextureFormat_RGBA32Float
	TextureFormatDepth32Float  TextureFormat = C.WGPUTextureFormat_Depth32Float
	TextureFormatUndefined     TextureFormat = C.WGPUTextureFormat_Undefined
)

type TextureUsage uint32

const (
	TextureUsageCopySrc         TextureUsage = C.WGPUTextureUsage_CopySrc
	TextureUsageCopyDst         TextureUsage = C.WGPUTextureUsage_CopyDst
	TextureUsageTextureBinding  TextureUsage = C.WGPUTextureUsage_TextureBinding
	TextureUsageRenderAttachment TextureUsage = C.WGPUTextureUsage_RenderAttachment
)

type BufferUsage uint32

const (
	BufferUsageVertex  BufferUsage = C.WGPUBufferUsage_Vertex
	BufferUsageIndex   BufferUsage = C.WGPUBufferUsage_Index
	BufferUsageUniform BufferUsage = C.WGPUBufferUsage_Uniform
	BufferUsageCopyDst BufferUsage = C.WGPUBufferUsage_CopyDst
)

type ShaderStage uint32

const (
	ShaderStageVertex   ShaderStage = C.WGPUShaderStage_Vertex
	ShaderStageFragment ShaderStage = C.WGPUShaderStage_Fragment
)

type VertexFormat uint32

const (
	VertexFormatFloat32x2 VertexFormat = C.WGPUVertexFormat_Float32x2
	VertexFormatFloat32x3 VertexFormat = C.WGPUVertexFormat_Float32x3
	VertexFormatFloat32x4 VertexFormat = C.WGPUVertexFormat_Float32x4
	VertexFormatUnorm8x4  VertexFormat = C.WGPUVertexFormat_Unorm8x4
)

type VertexStepMode uint32

const (
	VertexStepModeVertex   VertexStepMode = C.WGPUVertexStepMode_Vertex
	VertexStepModeInstance VertexStepMode = C.WGPUVertexStepMode_Instance
)

type PrimitiveTopology uint32

const (
	PrimitiveTopologyTriangleList PrimitiveTopology = C.WGPUPrimitiveTopology_TriangleList
	PrimitiveTopologyLineList     PrimitiveTopology = C.WGPUPrimitiveTopology_LineList
	PrimitiveTopologyPointList    PrimitiveTopology = C.WGPUPrimitiveTopology_PointList
)

type FrontFace uint32

const (
	FrontFaceCCW FrontFace = C.WGPUFrontFace_CCW
	FrontFaceCW  FrontFace = C.WGPUFrontFace_CW
)

type CullMode uint32

const (
	CullModeNone  CullMode = C.WGPUCullMode_None
	CullModeFront CullMode = C.WGPUCullMode_Front
	CullModeBack  CullMode = C.WGPUCullMode_Back
)

type CompareFunction uint32

const (
	CompareFunctionNever        CompareFunction = C.WGPUCompareFunction_Never
	CompareFunctionLess         CompareFunction = C.WGPUCompareFunction_Less
	CompareFunctionEqual        CompareFunction = C.WGPUCompareFunction_Equal
	CompareFunctionLessEqual    CompareFunction = C.WGPUCompareFunction_LessEqual
	CompareFunctionAlways       CompareFunction = C.WGPUCompareFunction_Always
	CompareFunctionUndefined    CompareFunction = C.WGPUCompareFunction_Undefined
)

type BlendFactor uint32

const (
	BlendFactorZero             BlendFactor = C.WGPUBlendFactor_Zero
	BlendFactorOne              BlendFactor = C.WGPUBlendFactor_One
	BlendFactorSrcAlpha         BlendFactor = C.WGPUBlendFactor_SrcAlpha
	BlendFactorOneMinusSrcAlpha BlendFactor = C.WGPUBlendFactor_OneMinusSrcAlpha
)

type BlendOperation uint32

const (
	BlendOperationAdd BlendOperation = C.WGPUBlendOperation_Add
	BlendOperationMax BlendOperation = C.WGPUBlendOperation_Max
)

type ColorWriteMask uint32

const (
	ColorWriteMaskAll  ColorWriteMask = C.WGPUColorWriteMask_All
	ColorWriteMaskNone ColorWriteMask = C.WGPUColorWriteMask_None
)

type IndexFormat uint32

const (
	IndexFormatUint16 IndexFormat = C.WGPUIndexFormat_Uint16
	IndexFormatUint32 IndexFormat = C.WGPUIndexFormat_Uint32
)

type LoadOp uint32

const (
	LoadOpClear     LoadOp = C.WGPULoadOp_Clear
	LoadOpLoad      LoadOp = C.WGPULoadOp_Load
	LoadOpUndefined LoadOp = C.WGPULoadOp_Undefined
)

type StoreOp uint32

const (
	StoreOpStore     StoreOp = C.WGPUStoreOp_Store
	StoreOpDiscard   StoreOp = C.WGPUStoreOp_Discard
	StoreOpUndefined StoreOp = C.WGPUStoreOp_Undefined
)

type FilterMode uint32

const (
	FilterModeNearest FilterMode = C.WGPUFilterMode_Nearest
	FilterModeLinear  FilterMode = C.WGPUFilterMode_Linear
)

type MipmapFilterMode uint32

const (
	MipmapFilterModeNearest MipmapFilterMode = C.WGPUMipmapFilterMode_Nearest
	MipmapFilterModeLinear  MipmapFilterMode = C.WGPUMipmapFilterMode_Linear
)

type AddressMode uint32

const (
	AddressModeRepeat       AddressMode = C.WGPUAddressMode_Repeat
	AddressModeClampToEdge  AddressMode = C.WGPUAddressMode_ClampToEdge
	AddressModeMirrorRepeat AddressMode = C.WGPUAddressMode_MirrorRepeat
)

type TextureDimension uint32

const (
	TextureDimension2D TextureDimension = C.WGPUTextureDimension_2D
)

type TextureViewDimension uint32

const (
	TextureViewDimension2D      TextureViewDimension = C.WGPUTextureViewDimension_2D
	TextureViewDimension2DArray TextureViewDimension = C.WGPUTextureViewDimension_2DArray
)

type TextureSampleType uint32

const (
	TextureSampleTypeFloat          TextureSampleType = C.WGPUTextureSampleType_Float
	TextureSampleTypeDepth          TextureSampleType = C.WGPUTextureSampleType_Depth
	TextureSampleTypeUnfilterableFloat TextureSampleType = C.WGPUTextureSampleType_UnfilterableFloat
)

type SamplerBindingType uint32

const (
	SamplerBindingTypeFiltering    SamplerBindingType = C.WGPUSamplerBindingType_Filtering
	SamplerBindingTypeNonFiltering SamplerBindingType = C.WGPUSamplerBindingType_NonFiltering
	SamplerBindingTypeComparison   SamplerBindingType = C.WGPUSamplerBindingType_Comparison
)

type BufferBindingType uint32

const (
	BufferBindingTypeUniform BufferBindingType = C.WGPUBufferBindingType_Uniform
)

type SurfaceGetCurrentTextureStatus uint32

const (
	SurfaceGetCurrentTextureStatusSuccessOptimal    SurfaceGetCurrentTextureStatus = C.WGPUSurfaceGetCurrentTextureStatus_SuccessOptimal
	SurfaceGetCurrentTextureStatusSuccessSuboptimal SurfaceGetCurrentTextureStatus = C.WGPUSurfaceGetCurrentTextureStatus_SuccessSuboptimal
)

// --- Descriptor structs ---

type Color struct {
	R, G, B, A float64
}

type Extent3D struct {
	Width, Height, DepthOrArrayLayers uint32
}

type Origin3D struct {
	X, Y, Z uint32
}

// --- Instance creation ---

func CreateInstance() Instance {
	inst := C.wgpuCreateInstance(nil)
	return Instance{inst}
}

func (i Instance) Release() {
	C.wgpuInstanceRelease(i.ref)
}

// --- Adapter request (synchronous wrapper) ---

// adapterResult is used for the async callback.
type adapterResult struct {
	adapter C.WGPUAdapter
	err     error
}

var (
	adapterChan   chan adapterResult
	adapterChanMu sync.Mutex
)

//export goRequestAdapterCallback
func goRequestAdapterCallback(status C.WGPURequestAdapterStatus, adapter C.WGPUAdapter, message C.WGPUStringView, userdata1, userdata2 unsafe.Pointer) {
	r := adapterResult{adapter: adapter}
	if status != C.WGPURequestAdapterStatus_Success {
		msg := ""
		if message.data != nil && message.length > 0 {
			msg = C.GoStringN(message.data, C.int(message.length))
		}
		r.err = fmt.Errorf("request adapter failed (status %d): %s", status, msg)
	}
	adapterChan <- r
}

func (i Instance) RequestAdapter() (Adapter, error) {
	adapterChanMu.Lock()
	adapterChan = make(chan adapterResult, 1)
	adapterChanMu.Unlock()

	opts := C.WGPURequestAdapterOptions{}
	opts.powerPreference = C.WGPUPowerPreference_HighPerformance

	cbInfo := C.WGPURequestAdapterCallbackInfo{}
	cbInfo.mode = C.WGPUCallbackMode_AllowProcessEvents
	cbInfo.callback = C.WGPURequestAdapterCallback(C.goRequestAdapterCallback)

	C.wgpuInstanceRequestAdapter(i.ref, &opts, cbInfo)
	C.wgpuInstanceProcessEvents(i.ref)

	r := <-adapterChan
	if r.err != nil {
		return Adapter{}, r.err
	}
	return Adapter{r.adapter}, nil
}

// --- Device request (synchronous wrapper) ---

type deviceResult struct {
	device C.WGPUDevice
	err    error
}

var (
	deviceChan   chan deviceResult
	deviceChanMu sync.Mutex
)

//export goRequestDeviceCallback
func goRequestDeviceCallback(status C.WGPURequestDeviceStatus, device C.WGPUDevice, message C.WGPUStringView, userdata1, userdata2 unsafe.Pointer) {
	r := deviceResult{device: device}
	if status != C.WGPURequestDeviceStatus_Success {
		msg := ""
		if message.data != nil && message.length > 0 {
			msg = C.GoStringN(message.data, C.int(message.length))
		}
		r.err = fmt.Errorf("request device failed (status %d): %s", status, msg)
	}
	deviceChan <- r
}

func (a Adapter) RequestDevice() (Device, error) {
	deviceChanMu.Lock()
	deviceChan = make(chan deviceResult, 1)
	deviceChanMu.Unlock()

	cbInfo := C.WGPURequestDeviceCallbackInfo{}
	cbInfo.mode = C.WGPUCallbackMode_AllowProcessEvents
	cbInfo.callback = C.WGPURequestDeviceCallback(C.goRequestDeviceCallback)

	// We need to process events on the instance but we don't have it here.
	// wgpu-native fires AllowProcessEvents callbacks during the request call itself.
	C.wgpuAdapterRequestDevice(a.ref, nil, cbInfo)

	r := <-deviceChan
	if r.err != nil {
		return Device{}, r.err
	}
	return Device{r.device}, nil
}

func (a Adapter) Release() {
	C.wgpuAdapterRelease(a.ref)
}

// --- Device ---

func (d Device) GetQueue() Queue {
	return Queue{C.wgpuDeviceGetQueue(d.ref)}
}

func (d Device) Release() {
	C.wgpuDeviceRelease(d.ref)
}

// --- Surface ---

// CreateMetalSurface creates a surface from a CAMetalLayer pointer.
func (i Instance) CreateMetalSurface(metalLayer unsafe.Pointer) (Surface, error) {
	// Allocate in C memory to avoid cgo pointer checks
	metalSource := (*C.WGPUSurfaceSourceMetalLayer)(C.calloc(1, C.size_t(unsafe.Sizeof(C.WGPUSurfaceSourceMetalLayer{}))))
	defer C.free(unsafe.Pointer(metalSource))
	metalSource.chain.sType = C.WGPUSType_SurfaceSourceMetalLayer
	metalSource.layer = metalLayer

	desc := (*C.WGPUSurfaceDescriptor)(C.calloc(1, C.size_t(unsafe.Sizeof(C.WGPUSurfaceDescriptor{}))))
	defer C.free(unsafe.Pointer(desc))
	desc.nextInChain = (*C.WGPUChainedStruct)(unsafe.Pointer(metalSource))

	s := C.wgpuInstanceCreateSurface(i.ref, desc)
	if s == nil {
		return Surface{}, errors.New("failed to create surface")
	}
	return Surface{s}, nil
}

func (s Surface) Configure(device Device, format TextureFormat, width, height uint32) {
	config := C.WGPUSurfaceConfiguration{}
	config.device = device.ref
	config.format = C.WGPUTextureFormat(format)
	config.width = C.uint32_t(width)
	config.height = C.uint32_t(height)
	config.usage = C.WGPUTextureUsage_RenderAttachment
	config.presentMode = C.WGPUPresentMode_Fifo // vsync

	C.wgpuSurfaceConfigure(s.ref, &config)
}

type SurfaceTexture struct {
	Texture Texture
	Status  SurfaceGetCurrentTextureStatus
}

func (s Surface) GetCurrentTexture() SurfaceTexture {
	var st C.WGPUSurfaceTexture
	C.wgpuSurfaceGetCurrentTexture(s.ref, &st)
	return SurfaceTexture{
		Texture: Texture{st.texture},
		Status:  SurfaceGetCurrentTextureStatus(st.status),
	}
}

func (s Surface) Present() {
	C.wgpuSurfacePresent(s.ref)
}

func (s Surface) Release() {
	C.wgpuSurfaceRelease(s.ref)
}

// --- Buffer ---

func (d Device) CreateBuffer(size uint64, usage BufferUsage) Buffer {
	desc := C.WGPUBufferDescriptor{}
	desc.size = C.uint64_t(size)
	desc.usage = C.WGPUBufferUsage(usage)
	return Buffer{C.wgpuDeviceCreateBuffer(d.ref, &desc)}
}

func (q Queue) WriteBuffer(buf Buffer, offset uint64, data unsafe.Pointer, size uint64) {
	C.wgpuQueueWriteBuffer(q.ref, buf.ref, C.uint64_t(offset), data, C.size_t(size))
}

func (b Buffer) Release() {
	if b.ref != nil {
		C.wgpuBufferRelease(b.ref)
	}
}

// --- Texture ---

type TextureDescriptor struct {
	Size      Extent3D
	Format    TextureFormat
	Usage     TextureUsage
	Dimension TextureDimension
	MipLevels uint32
}

func (d Device) CreateTexture(desc TextureDescriptor) Texture {
	cd := C.WGPUTextureDescriptor{}
	cd.size.width = C.uint32_t(desc.Size.Width)
	cd.size.height = C.uint32_t(desc.Size.Height)
	cd.size.depthOrArrayLayers = C.uint32_t(desc.Size.DepthOrArrayLayers)
	if cd.size.depthOrArrayLayers == 0 {
		cd.size.depthOrArrayLayers = 1
	}
	cd.format = C.WGPUTextureFormat(desc.Format)
	cd.usage = C.WGPUTextureUsage(desc.Usage)
	cd.dimension = C.WGPUTextureDimension(desc.Dimension)
	cd.mipLevelCount = C.uint32_t(desc.MipLevels)
	if cd.mipLevelCount == 0 {
		cd.mipLevelCount = 1
	}
	cd.sampleCount = 1
	return Texture{C.wgpuDeviceCreateTexture(d.ref, &cd)}
}

func (t Texture) CreateView() TextureView {
	return TextureView{C.wgpuTextureCreateView(t.ref, nil)}
}

// CreateViewArray creates a view of the full texture as a 2D array.
func (t Texture) CreateViewArray(layers uint32) TextureView {
	desc := (*C.WGPUTextureViewDescriptor)(C.calloc(1, C.size_t(unsafe.Sizeof(C.WGPUTextureViewDescriptor{}))))
	defer C.free(unsafe.Pointer(desc))
	desc.dimension = C.WGPUTextureViewDimension_2DArray
	desc.baseArrayLayer = 0
	desc.arrayLayerCount = C.uint32_t(layers)
	desc.baseMipLevel = 0
	desc.mipLevelCount = 1
	desc.format = C.WGPUTextureFormat_Depth32Float
	desc.aspect = C.WGPUTextureAspect_DepthOnly
	return TextureView{C.wgpuTextureCreateView(t.ref, desc)}
}

// CreateViewLayer creates a view of a single layer in a 2D array texture.
func (t Texture) CreateViewLayer(layer uint32) TextureView {
	desc := (*C.WGPUTextureViewDescriptor)(C.calloc(1, C.size_t(unsafe.Sizeof(C.WGPUTextureViewDescriptor{}))))
	defer C.free(unsafe.Pointer(desc))
	desc.dimension = C.WGPUTextureViewDimension_2D
	desc.baseArrayLayer = C.uint32_t(layer)
	desc.arrayLayerCount = 1
	desc.baseMipLevel = 0
	desc.mipLevelCount = 1
	desc.format = C.WGPUTextureFormat_Depth32Float
	desc.aspect = C.WGPUTextureAspect_DepthOnly
	return TextureView{C.wgpuTextureCreateView(t.ref, desc)}
}

func (t Texture) Release() {
	if t.ref != nil {
		C.wgpuTextureRelease(t.ref)
	}
}

func (tv TextureView) Release() {
	if tv.ref != nil {
		C.wgpuTextureViewRelease(tv.ref)
	}
}

type ImageCopyTexture struct {
	Texture  Texture
	MipLevel uint32
	Origin   Origin3D
}

type TextureDataLayout struct {
	Offset       uint64
	BytesPerRow  uint32
	RowsPerImage uint32
}

func (q Queue) WriteTexture(dst ImageCopyTexture, data unsafe.Pointer, dataSize uint64, layout TextureDataLayout, size Extent3D) {
	cdst := C.WGPUTexelCopyTextureInfo{}
	cdst.texture = dst.Texture.ref
	cdst.mipLevel = C.uint32_t(dst.MipLevel)
	cdst.origin.x = C.uint32_t(dst.Origin.X)
	cdst.origin.y = C.uint32_t(dst.Origin.Y)
	cdst.origin.z = C.uint32_t(dst.Origin.Z)

	clayout := C.WGPUTexelCopyBufferLayout{}
	clayout.offset = C.uint64_t(layout.Offset)
	clayout.bytesPerRow = C.uint32_t(layout.BytesPerRow)
	clayout.rowsPerImage = C.uint32_t(layout.RowsPerImage)

	csize := C.WGPUExtent3D{}
	csize.width = C.uint32_t(size.Width)
	csize.height = C.uint32_t(size.Height)
	csize.depthOrArrayLayers = C.uint32_t(size.DepthOrArrayLayers)
	if csize.depthOrArrayLayers == 0 {
		csize.depthOrArrayLayers = 1
	}

	C.wgpuQueueWriteTexture(q.ref, &cdst, data, C.size_t(dataSize), &clayout, &csize)
}

// --- Sampler ---

type SamplerDescriptor struct {
	AddressModeU  AddressMode
	AddressModeV  AddressMode
	AddressModeW  AddressMode
	MagFilter     FilterMode
	MinFilter     FilterMode
	MipmapFilter  MipmapFilterMode
	MaxAnisotropy uint16
	Compare       CompareFunction // set to non-zero for comparison sampler
}

func (d Device) CreateSampler(desc SamplerDescriptor) Sampler {
	cd := C.WGPUSamplerDescriptor{}
	cd.addressModeU = C.WGPUAddressMode(desc.AddressModeU)
	cd.addressModeV = C.WGPUAddressMode(desc.AddressModeV)
	cd.addressModeW = C.WGPUAddressMode(desc.AddressModeW)
	cd.magFilter = C.WGPUFilterMode(desc.MagFilter)
	cd.minFilter = C.WGPUFilterMode(desc.MinFilter)
	cd.mipmapFilter = C.WGPUMipmapFilterMode(desc.MipmapFilter)
	cd.maxAnisotropy = C.uint16_t(desc.MaxAnisotropy)
	if cd.maxAnisotropy == 0 {
		cd.maxAnisotropy = 1
	}
	if desc.Compare != 0 {
		cd.compare = C.WGPUCompareFunction(desc.Compare)
	}
	return Sampler{C.wgpuDeviceCreateSampler(d.ref, &cd)}
}

func (s Sampler) Release() {
	if s.ref != nil {
		C.wgpuSamplerRelease(s.ref)
	}
}

// --- Shader Module ---

func (d Device) CreateShaderModuleWGSL(code string) ShaderModule {
	ccode := C.CString(code)
	defer C.free(unsafe.Pointer(ccode))

	wgslSource := (*C.WGPUShaderSourceWGSL)(C.calloc(1, C.size_t(unsafe.Sizeof(C.WGPUShaderSourceWGSL{}))))
	defer C.free(unsafe.Pointer(wgslSource))
	wgslSource.chain.sType = C.WGPUSType_ShaderSourceWGSL
	wgslSource.code.data = ccode
	wgslSource.code.length = C.size_t(len(code))

	desc := (*C.WGPUShaderModuleDescriptor)(C.calloc(1, C.size_t(unsafe.Sizeof(C.WGPUShaderModuleDescriptor{}))))
	defer C.free(unsafe.Pointer(desc))
	desc.nextInChain = (*C.WGPUChainedStruct)(unsafe.Pointer(wgslSource))

	return ShaderModule{C.wgpuDeviceCreateShaderModule(d.ref, desc)}
}

func (sm ShaderModule) Release() {
	if sm.ref != nil {
		C.wgpuShaderModuleRelease(sm.ref)
	}
}

// --- Bind Group Layout ---

type BindGroupLayoutEntry struct {
	Binding    uint32
	Visibility ShaderStage
	// Only one of these should be set:
	Buffer  *BufferBindingLayout
	Sampler *SamplerBindingLayout
	Texture *TextureBindingLayout
}

type BufferBindingLayout struct {
	Type BufferBindingType
}

type SamplerBindingLayout struct {
	Type SamplerBindingType
}

type TextureBindingLayout struct {
	SampleType    TextureSampleType
	ViewDimension TextureViewDimension
}

func (d Device) CreateBindGroupLayout(entries []BindGroupLayoutEntry) BindGroupLayout {
	n := len(entries)
	centries := (*C.WGPUBindGroupLayoutEntry)(C.calloc(C.size_t(max(n, 1)), C.size_t(unsafe.Sizeof(C.WGPUBindGroupLayoutEntry{}))))
	defer C.free(unsafe.Pointer(centries))
	cslice := unsafe.Slice(centries, max(n, 1))
	for i, e := range entries {
		cslice[i].binding = C.uint32_t(e.Binding)
		cslice[i].visibility = C.WGPUShaderStage(e.Visibility)
		if e.Buffer != nil {
			cslice[i].buffer._type = C.WGPUBufferBindingType(e.Buffer.Type)
		}
		if e.Sampler != nil {
			cslice[i].sampler._type = C.WGPUSamplerBindingType(e.Sampler.Type)
		}
		if e.Texture != nil {
			cslice[i].texture.sampleType = C.WGPUTextureSampleType(e.Texture.SampleType)
			cslice[i].texture.viewDimension = C.WGPUTextureViewDimension(e.Texture.ViewDimension)
		}
	}

	desc := (*C.WGPUBindGroupLayoutDescriptor)(C.calloc(1, C.size_t(unsafe.Sizeof(C.WGPUBindGroupLayoutDescriptor{}))))
	defer C.free(unsafe.Pointer(desc))
	desc.entryCount = C.size_t(n)
	if n > 0 {
		desc.entries = centries
	}

	return BindGroupLayout{C.wgpuDeviceCreateBindGroupLayout(d.ref, desc)}
}

func (bgl BindGroupLayout) Release() {
	if bgl.ref != nil {
		C.wgpuBindGroupLayoutRelease(bgl.ref)
	}
}

// --- Bind Group ---

type BindGroupEntry struct {
	Binding     uint32
	Buffer      Buffer
	Offset      uint64
	Size        uint64
	Sampler     Sampler
	TextureView TextureView
}

func (d Device) CreateBindGroup(layout BindGroupLayout, entries []BindGroupEntry) BindGroup {
	n := len(entries)
	centries := (*C.WGPUBindGroupEntry)(C.calloc(C.size_t(max(n, 1)), C.size_t(unsafe.Sizeof(C.WGPUBindGroupEntry{}))))
	defer C.free(unsafe.Pointer(centries))
	cslice := unsafe.Slice(centries, max(n, 1))
	for i, e := range entries {
		cslice[i].binding = C.uint32_t(e.Binding)
		cslice[i].buffer = e.Buffer.ref
		cslice[i].offset = C.uint64_t(e.Offset)
		cslice[i].size = C.uint64_t(e.Size)
		cslice[i].sampler = e.Sampler.ref
		cslice[i].textureView = e.TextureView.ref
	}

	desc := (*C.WGPUBindGroupDescriptor)(C.calloc(1, C.size_t(unsafe.Sizeof(C.WGPUBindGroupDescriptor{}))))
	defer C.free(unsafe.Pointer(desc))
	desc.layout = layout.ref
	desc.entryCount = C.size_t(n)
	if n > 0 {
		desc.entries = centries
	}

	return BindGroup{C.wgpuDeviceCreateBindGroup(d.ref, desc)}
}

func (bg BindGroup) Release() {
	if bg.ref != nil {
		C.wgpuBindGroupRelease(bg.ref)
	}
}

// --- Pipeline Layout ---

func (d Device) CreatePipelineLayout(bindGroupLayouts []BindGroupLayout) PipelineLayout {
	n := len(bindGroupLayouts)
	crefs := (*C.WGPUBindGroupLayout)(C.calloc(C.size_t(max(n, 1)), C.size_t(unsafe.Sizeof(C.WGPUBindGroupLayout(nil)))))
	defer C.free(unsafe.Pointer(crefs))
	cslice := unsafe.Slice(crefs, max(n, 1))
	for i, bgl := range bindGroupLayouts {
		cslice[i] = bgl.ref
	}

	desc := (*C.WGPUPipelineLayoutDescriptor)(C.calloc(1, C.size_t(unsafe.Sizeof(C.WGPUPipelineLayoutDescriptor{}))))
	defer C.free(unsafe.Pointer(desc))
	desc.bindGroupLayoutCount = C.size_t(n)
	if n > 0 {
		desc.bindGroupLayouts = crefs
	}

	return PipelineLayout{C.wgpuDeviceCreatePipelineLayout(d.ref, desc)}
}

func (pl PipelineLayout) Release() {
	if pl.ref != nil {
		C.wgpuPipelineLayoutRelease(pl.ref)
	}
}

// --- Render Pipeline ---

type VertexAttribute struct {
	Format         VertexFormat
	Offset         uint64
	ShaderLocation uint32
}

type VertexBufferLayout struct {
	ArrayStride uint64
	StepMode    VertexStepMode
	Attributes  []VertexAttribute
}

type BlendComponent struct {
	SrcFactor BlendFactor
	DstFactor BlendFactor
	Operation BlendOperation
}

type BlendState struct {
	Color BlendComponent
	Alpha BlendComponent
}

type ColorTargetState struct {
	Format    TextureFormat
	Blend     *BlendState
	WriteMask ColorWriteMask
}

type DepthStencilState struct {
	Format            TextureFormat
	DepthWriteEnabled bool
	DepthCompare      CompareFunction
}

type RenderPipelineDescriptor struct {
	Layout         PipelineLayout
	VertexModule   ShaderModule
	VertexEntry    string
	FragmentModule ShaderModule
	FragmentEntry  string
	Buffers        []VertexBufferLayout
	Targets        []ColorTargetState
	Primitive      PrimitiveTopology
	FrontFace      FrontFace
	CullMode       CullMode
	DepthStencil   *DepthStencilState
}

func (d Device) CreateRenderPipeline(desc RenderPipelineDescriptor) RenderPipeline {
	var pinner runtime.Pinner
	defer pinner.Unpin()

	cVertexEntry := C.CString(desc.VertexEntry)
	defer C.free(unsafe.Pointer(cVertexEntry))
	cFragEntry := C.CString(desc.FragmentEntry)
	defer C.free(unsafe.Pointer(cFragEntry))

	// Convert vertex buffer layouts
	attrArrays := make([][]C.WGPUVertexAttribute, len(desc.Buffers))
	cBufferLayouts := make([]C.WGPUVertexBufferLayout, len(desc.Buffers))
	for i, buf := range desc.Buffers {
		attrArrays[i] = make([]C.WGPUVertexAttribute, len(buf.Attributes))
		for j, attr := range buf.Attributes {
			attrArrays[i][j].format = C.WGPUVertexFormat(attr.Format)
			attrArrays[i][j].offset = C.uint64_t(attr.Offset)
			attrArrays[i][j].shaderLocation = C.uint32_t(attr.ShaderLocation)
		}
		cBufferLayouts[i].arrayStride = C.uint64_t(buf.ArrayStride)
		cBufferLayouts[i].stepMode = C.WGPUVertexStepMode(buf.StepMode)
		cBufferLayouts[i].attributeCount = C.size_t(len(buf.Attributes))
		if len(buf.Attributes) > 0 {
			pinner.Pin(&attrArrays[i][0])
			cBufferLayouts[i].attributes = &attrArrays[i][0]
		}
	}

	// Convert color targets
	cBlendStates := make([]C.WGPUBlendState, len(desc.Targets))
	cTargets := make([]C.WGPUColorTargetState, len(desc.Targets))
	for i, t := range desc.Targets {
		cTargets[i].format = C.WGPUTextureFormat(t.Format)
		cTargets[i].writeMask = C.WGPUColorWriteMask(t.WriteMask)
		if t.Blend != nil {
			cBlendStates[i].color.srcFactor = C.WGPUBlendFactor(t.Blend.Color.SrcFactor)
			cBlendStates[i].color.dstFactor = C.WGPUBlendFactor(t.Blend.Color.DstFactor)
			cBlendStates[i].color.operation = C.WGPUBlendOperation(t.Blend.Color.Operation)
			cBlendStates[i].alpha.srcFactor = C.WGPUBlendFactor(t.Blend.Alpha.SrcFactor)
			cBlendStates[i].alpha.dstFactor = C.WGPUBlendFactor(t.Blend.Alpha.DstFactor)
			cBlendStates[i].alpha.operation = C.WGPUBlendOperation(t.Blend.Alpha.Operation)
			pinner.Pin(&cBlendStates[i])
			cTargets[i].blend = &cBlendStates[i]
		}
	}

	cd := C.WGPURenderPipelineDescriptor{}
	cd.layout = desc.Layout.ref

	cd.vertex.module = desc.VertexModule.ref
	cd.vertex.entryPoint.data = cVertexEntry
	cd.vertex.entryPoint.length = C.size_t(len(desc.VertexEntry))
	cd.vertex.bufferCount = C.size_t(len(cBufferLayouts))
	if len(cBufferLayouts) > 0 {
		pinner.Pin(&cBufferLayouts[0])
		cd.vertex.buffers = &cBufferLayouts[0]
	}

	fragState := C.WGPUFragmentState{}
	fragState.module = desc.FragmentModule.ref
	fragState.entryPoint.data = cFragEntry
	fragState.entryPoint.length = C.size_t(len(desc.FragmentEntry))
	fragState.targetCount = C.size_t(len(cTargets))
	if len(cTargets) > 0 {
		pinner.Pin(&cTargets[0])
		fragState.targets = &cTargets[0]
	}
	pinner.Pin(&fragState)
	cd.fragment = &fragState

	cd.primitive.topology = C.WGPUPrimitiveTopology(desc.Primitive)
	cd.primitive.frontFace = C.WGPUFrontFace(desc.FrontFace)
	cd.primitive.cullMode = C.WGPUCullMode(desc.CullMode)

	// Depth stencil
	var cDepthStencil C.WGPUDepthStencilState
	if desc.DepthStencil != nil {
		cDepthStencil.format = C.WGPUTextureFormat(desc.DepthStencil.Format)
		if desc.DepthStencil.DepthWriteEnabled {
			cDepthStencil.depthWriteEnabled = C.WGPUOptionalBool_True
		} else {
			cDepthStencil.depthWriteEnabled = C.WGPUOptionalBool_False
		}
		cDepthStencil.depthCompare = C.WGPUCompareFunction(desc.DepthStencil.DepthCompare)
		pinner.Pin(&cDepthStencil)
		cd.depthStencil = &cDepthStencil
	}

	// Multisample
	cd.multisample.count = 1
	cd.multisample.mask = 0xFFFFFFFF

	return RenderPipeline{C.wgpuDeviceCreateRenderPipeline(d.ref, &cd)}
}

func (rp RenderPipeline) Release() {
	if rp.ref != nil {
		C.wgpuRenderPipelineRelease(rp.ref)
	}
}

// --- Command Encoder ---

func (d Device) CreateCommandEncoder() CommandEncoder {
	desc := C.WGPUCommandEncoderDescriptor{}
	return CommandEncoder{C.wgpuDeviceCreateCommandEncoder(d.ref, &desc)}
}

type RenderPassColorAttachment struct {
	View       TextureView
	LoadOp     LoadOp
	StoreOp    StoreOp
	ClearValue Color
}

type RenderPassDepthStencilAttachment struct {
	View            TextureView
	DepthLoadOp     LoadOp
	DepthStoreOp    StoreOp
	DepthClearValue float32
}

type RenderPassDescriptor struct {
	ColorAttachments       []RenderPassColorAttachment
	DepthStencilAttachment *RenderPassDepthStencilAttachment
}

func (ce CommandEncoder) BeginRenderPass(desc RenderPassDescriptor) RenderPassEncoder {
	nColors := len(desc.ColorAttachments)

	// Allocate all structs in C memory to avoid cgo pointer checks
	colorAttachments := (*C.WGPURenderPassColorAttachment)(nil)
	if nColors > 0 {
		colorAttachments = (*C.WGPURenderPassColorAttachment)(C.calloc(C.size_t(nColors), C.size_t(unsafe.Sizeof(C.WGPURenderPassColorAttachment{}))))
		defer C.free(unsafe.Pointer(colorAttachments))
		caSlice := unsafe.Slice(colorAttachments, nColors)
		for i, ca := range desc.ColorAttachments {
			caSlice[i].view = ca.View.ref
			caSlice[i].depthSlice = C.WGPU_DEPTH_SLICE_UNDEFINED
			caSlice[i].loadOp = C.WGPULoadOp(ca.LoadOp)
			caSlice[i].storeOp = C.WGPUStoreOp(ca.StoreOp)
			caSlice[i].clearValue = C.WGPUColor{
				r: C.double(ca.ClearValue.R),
				g: C.double(ca.ClearValue.G),
				b: C.double(ca.ClearValue.B),
				a: C.double(ca.ClearValue.A),
			}
		}
	}

	cd := (*C.WGPURenderPassDescriptor)(C.calloc(1, C.size_t(unsafe.Sizeof(C.WGPURenderPassDescriptor{}))))
	defer C.free(unsafe.Pointer(cd))
	cd.colorAttachmentCount = C.size_t(nColors)
	cd.colorAttachments = colorAttachments

	var cDepthPtr *C.WGPURenderPassDepthStencilAttachment
	if desc.DepthStencilAttachment != nil {
		cDepthPtr = (*C.WGPURenderPassDepthStencilAttachment)(C.calloc(1, C.size_t(unsafe.Sizeof(C.WGPURenderPassDepthStencilAttachment{}))))
		defer C.free(unsafe.Pointer(cDepthPtr))
		cDepthPtr.view = desc.DepthStencilAttachment.View.ref
		cDepthPtr.depthLoadOp = C.WGPULoadOp(desc.DepthStencilAttachment.DepthLoadOp)
		cDepthPtr.depthStoreOp = C.WGPUStoreOp(desc.DepthStencilAttachment.DepthStoreOp)
		cDepthPtr.depthClearValue = C.float(desc.DepthStencilAttachment.DepthClearValue)
		cDepthPtr.stencilLoadOp = C.WGPULoadOp_Undefined
		cDepthPtr.stencilStoreOp = C.WGPUStoreOp_Undefined
		cd.depthStencilAttachment = cDepthPtr
	}

	return RenderPassEncoder{C.wgpuCommandEncoderBeginRenderPass(ce.ref, cd)}
}

func (ce CommandEncoder) Finish() CommandBuffer {
	return CommandBuffer{C.wgpuCommandEncoderFinish(ce.ref, nil)}
}

func (cb CommandBuffer) Release() {
	if cb.ref != nil {
		C.wgpuCommandBufferRelease(cb.ref)
	}
}

func (ce CommandEncoder) Release() {
	if ce.ref != nil {
		C.wgpuCommandEncoderRelease(ce.ref)
	}
}

// --- Render Pass Encoder ---

func (rpe RenderPassEncoder) SetPipeline(pipeline RenderPipeline) {
	C.wgpuRenderPassEncoderSetPipeline(rpe.ref, pipeline.ref)
}

func (rpe RenderPassEncoder) SetVertexBuffer(slot uint32, buffer Buffer, offset, size uint64) {
	C.wgpuRenderPassEncoderSetVertexBuffer(rpe.ref, C.uint32_t(slot), buffer.ref, C.uint64_t(offset), C.uint64_t(size))
}

func (rpe RenderPassEncoder) SetIndexBuffer(buffer Buffer, format IndexFormat, offset, size uint64) {
	C.wgpuRenderPassEncoderSetIndexBuffer(rpe.ref, buffer.ref, C.WGPUIndexFormat(format), C.uint64_t(offset), C.uint64_t(size))
}

func (rpe RenderPassEncoder) SetBindGroup(groupIndex uint32, group BindGroup) {
	C.wgpuRenderPassEncoderSetBindGroup(rpe.ref, C.uint32_t(groupIndex), group.ref, 0, nil)
}

func (rpe RenderPassEncoder) DrawIndexed(indexCount, instanceCount, firstIndex uint32, baseVertex int32, firstInstance uint32) {
	C.wgpuRenderPassEncoderDrawIndexed(rpe.ref, C.uint32_t(indexCount), C.uint32_t(instanceCount), C.uint32_t(firstIndex), C.int32_t(baseVertex), C.uint32_t(firstInstance))
}

func (rpe RenderPassEncoder) SetViewport(x, y, width, height, minDepth, maxDepth float32) {
	C.wgpuRenderPassEncoderSetViewport(rpe.ref, C.float(x), C.float(y), C.float(width), C.float(height), C.float(minDepth), C.float(maxDepth))
}

func (rpe RenderPassEncoder) SetScissorRect(x, y, width, height uint32) {
	C.wgpuRenderPassEncoderSetScissorRect(rpe.ref, C.uint32_t(x), C.uint32_t(y), C.uint32_t(width), C.uint32_t(height))
}

func (rpe RenderPassEncoder) End() {
	C.wgpuRenderPassEncoderEnd(rpe.ref)
}

func (rpe RenderPassEncoder) Release() {
	if rpe.ref != nil {
		C.wgpuRenderPassEncoderRelease(rpe.ref)
	}
}

// --- Queue ---

func (q Queue) Submit(cmdBuf CommandBuffer) {
	C.wgpuQueueSubmit(q.ref, 1, &cmdBuf.ref)
}

func (q Queue) Release() {
	C.wgpuQueueRelease(q.ref)
}

// --- Logging ---

func SetLogLevel(level int) {
	C.wgpuSetLogLevel(C.WGPULogLevel(level))
}

const (
	LogLevelOff   = int(C.WGPULogLevel_Off)
	LogLevelError = int(C.WGPULogLevel_Error)
	LogLevelWarn  = int(C.WGPULogLevel_Warn)
	LogLevelInfo  = int(C.WGPULogLevel_Info)
	LogLevelDebug = int(C.WGPULogLevel_Debug)
	LogLevelTrace = int(C.WGPULogLevel_Trace)
)
