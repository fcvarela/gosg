package core

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"unsafe"

	"github.com/fcvarela/gosg/gpu"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/golang/glog"
	_ "golang.org/x/image/bmp"
)

// InstanceData holds the per-instance-data when using instanced drawing.
type InstanceData struct {
	ModelMatrix               mgl32.Mat4
	ModelViewProjectionMatrix mgl32.Mat4
	Custom                    [4]mgl32.Vec4
}

const (
	// MaxInstances is the maximum number of instances per draw call
	MaxInstances = 2000

	// InstanceDataLen is the byte size of an InstanceData value
	InstanceDataLen = (2*16 + 4*4) * 4
)

// RenderPassColorAttachment describes a color attachment for a render pass.
type RenderPassColorAttachment struct {
	View       gpu.TextureView
	Format     gpu.TextureFormat // gpu.TextureFormatUndefined means use swap chain format
	LoadOp     gpu.LoadOp
	ClearColor mgl32.Vec4
}

// RenderPassDepthAttachment describes a depth attachment for a render pass.
type RenderPassDepthAttachment struct {
	View       gpu.TextureView
	Format     gpu.TextureFormat // gpu.TextureFormatUndefined means Depth32Float
	LoadOp     gpu.LoadOp
	ClearDepth float32
}

// RenderPassDescriptor describes how to create a render pass.
type RenderPassDescriptor struct {
	ColorAttachments []RenderPassColorAttachment
	DepthAttachment  *RenderPassDepthAttachment
	Viewport         mgl32.Vec4
}

// RenderPass wraps a gpu.RenderPassEncoder with engine-level convenience methods.
type RenderPass struct {
	encoder        gpu.RenderPassEncoder
	currentProgram *Program
	colorFormats   []gpu.TextureFormat
	depthFormat    gpu.TextureFormat
}

// SetPipeline looks up (or creates) the GPU pipeline for the given pipeline config and binds it.
// Returns false if the pipeline has no program and was not set.
func (rp *RenderPass) SetPipeline(p *Pipeline) bool {
	if p.ProgramName == "" {
		return false
	}
	program := resourceManager.Program(p.ProgramName)
	if program == nil {
		return false
	}
	rp.currentProgram = program

	pipeline := renderer.pipelines.getOrCreate(p, program, rp.colorFormats, rp.depthFormat)
	rp.encoder.SetPipeline(pipeline)
	return true
}

// SetCameraConstants creates a bind group for the UBO at group 0 and binds it.
func (rp *RenderPass) SetCameraConstants(ubo *UniformBuffer) {
	if rp.currentProgram == nil || ubo == nil {
		return
	}
	if len(rp.currentProgram.bindGroupLayouts) == 0 {
		return
	}
	bg := renderer.device.CreateBindGroup(rp.currentProgram.bindGroupLayouts[0], []gpu.BindGroupEntry{{
		Binding: 0,
		Buffer:  ubo.buffer,
		Offset:  0,
		Size:    ubo.size,
	}})
	rp.encoder.SetBindGroup(0, bg)
}

// SetMaterial creates a bind group for textures at group 1 and binds it.
func (rp *RenderPass) SetMaterial(mat *Material) {
	if rp.currentProgram == nil {
		return
	}
	if len(rp.currentProgram.bindGroupLayouts) < 2 || len(rp.currentProgram.spec.TextureBindings) == 0 {
		return
	}
	entries := make([]gpu.BindGroupEntry, 0, len(rp.currentProgram.spec.TextureBindings)*2)
	for texName, binding := range rp.currentProgram.spec.TextureBindings {
		tex := mat.Texture(texName)
		if tex == nil {
			// Pick default based on the bind group layout's expected sample type
			isDepth := false
			if len(rp.currentProgram.spec.BindGroupLayouts) > int(binding.Group) {
				for _, e := range rp.currentProgram.spec.BindGroupLayouts[binding.Group].Entries {
					if e.Binding == binding.TextureBinding && e.Texture != nil && e.Texture.SampleType == "depth" {
						isDepth = true
					}
				}
			}
			if isDepth {
				tex = renderer.defaultDepthTexture
			} else {
				tex = renderer.defaultTexture
			}
		}
		entries = append(entries,
			gpu.BindGroupEntry{Binding: binding.TextureBinding, TextureView: tex.view},
			gpu.BindGroupEntry{Binding: binding.SamplerBinding, Sampler: tex.sampler},
		)
	}
	if len(entries) > 0 {
		bg := renderer.device.CreateBindGroup(rp.currentProgram.bindGroupLayouts[1], entries)
		rp.encoder.SetBindGroup(1, bg)
	}
}

// SetViewport sets the viewport on the render pass.
func (rp *RenderPass) SetViewport(x, y, w, h float32) {
	rp.encoder.SetViewport(x, y, w, h, 0.0, 1.0)
}

// SetScissorRect sets the scissor rectangle.
func (rp *RenderPass) SetScissorRect(x, y, w, h uint32) {
	rp.encoder.SetScissorRect(x, y, w, h)
}

// SetVertexBuffer binds a vertex buffer to a slot.
func (rp *RenderPass) SetVertexBuffer(slot uint32, buf gpu.Buffer, offset, size uint64) {
	rp.encoder.SetVertexBuffer(slot, buf, offset, size)
}

// SetIndexBuffer binds an index buffer.
func (rp *RenderPass) SetIndexBuffer(buf gpu.Buffer, format gpu.IndexFormat, offset, size uint64) {
	rp.encoder.SetIndexBuffer(buf, format, offset, size)
}

// SetGPUPipeline sets a raw GPU pipeline directly.
func (rp *RenderPass) SetGPUPipeline(pipeline gpu.RenderPipeline) {
	rp.encoder.SetPipeline(pipeline)
}

// SetBindGroup sets a bind group directly.
func (rp *RenderPass) SetBindGroup(group uint32, bg gpu.BindGroup) {
	rp.encoder.SetBindGroup(group, bg)
}

// DrawIndexed issues an indexed draw call.
func (rp *RenderPass) DrawIndexed(indexCount, instanceCount, firstIndex uint32, baseVertex int32, firstInstance uint32) {
	rp.encoder.DrawIndexed(indexCount, instanceCount, firstIndex, baseVertex, firstInstance)
}

// End ends the render pass.
func (rp *RenderPass) End() {
	rp.encoder.End()
	rp.encoder.Release()
}

// CurrentProgram returns the currently bound program (for ImGui rendering).
func (rp *RenderPass) CurrentProgram() *Program {
	return rp.currentProgram
}

// Renderer holds the wgpu device, queue, and rendering state.
type Renderer struct {
	instance      gpu.Instance
	device        gpu.Device
	queue         gpu.Queue
	surface       gpu.Surface
	surfaceFormat gpu.TextureFormat
	surfaceWidth  uint32
	surfaceHeight uint32

	// Per-frame state
	swapChainTexture gpu.Texture
	swapChainView    gpu.TextureView
	encoder          gpu.CommandEncoder

	// Pipeline cache and defaults
	pipelines           *pipelineCache
	defaultTexture      *Texture
	defaultDepthTexture *Texture
}

var renderer *Renderer

// InitRenderer creates and initializes the global renderer.
func InitRenderer(metalLayer unsafe.Pointer, width, height uint32) {
	r := &Renderer{
		surfaceFormat: gpu.TextureFormatBGRA8Unorm,
		surfaceWidth:  width,
		surfaceHeight: height,
	}

	r.instance = gpu.CreateInstance()

	var err error
	adapter, err := r.instance.RequestAdapter()
	if err != nil {
		glog.Fatal("Failed to get wgpu adapter: ", err)
	}

	r.device, err = adapter.RequestDevice()
	if err != nil {
		glog.Fatal("Failed to get wgpu device: ", err)
	}
	adapter.Release()

	r.queue = r.device.GetQueue()

	r.surface, err = r.instance.CreateMetalSurface(metalLayer)
	if err != nil {
		glog.Fatal("Failed to create wgpu surface: ", err)
	}

	r.surface.Configure(r.device, r.surfaceFormat, r.surfaceWidth, r.surfaceHeight)

	r.pipelines = newPipelineCache()

	// Create a default 1x1 white texture for missing texture bindings
	r.defaultTexture = r.NewTexture(TextureDescriptor{
		Width: 1, Height: 1, Target: TextureTarget2D,
		Format: TextureFormatRGBA, SizedFormat: TextureSizedFormatRGBA8,
		ComponentType: TextureComponentTypeUNSIGNEDBYTE,
		Filter: TextureFilterNearest, WrapMode: TextureWrapModeRepeat,
	}, []byte{255, 255, 255, 255})

	// Create a default 1x1 depth array texture for missing shadow bindings
	defaultDepthTex := r.device.CreateTexture(gpu.TextureDescriptor{
		Size:      gpu.Extent3D{Width: 1, Height: 1, DepthOrArrayLayers: 1},
		Format:    gpu.TextureFormatDepth32Float,
		Usage:     gpu.TextureUsageTextureBinding | gpu.TextureUsageRenderAttachment,
		Dimension: gpu.TextureDimension2D,
		MipLevels: 1,
	})
	defaultDepthView := defaultDepthTex.CreateViewArray(1)
	defaultDepthSampler := r.device.CreateSampler(gpu.SamplerDescriptor{
		AddressModeU: gpu.AddressModeClampToEdge,
		AddressModeV: gpu.AddressModeClampToEdge,
		Compare:      gpu.CompareFunctionLessEqual,
	})
	r.defaultDepthTexture = &Texture{
		id:      allocateTextureID(),
		texture: defaultDepthTex, view: defaultDepthView, sampler: defaultDepthSampler,
		descriptor: TextureDescriptor{Format: TextureFormatDEPTH, SizedFormat: TextureSizedFormatDEPTH32F},
	}

	renderer = r
	glog.Info("wgpu renderer initialized")
}

// GetRenderer returns the global renderer.
func GetRenderer() *Renderer {
	return renderer
}

// ProgramExtension returns the file extension for program specs.
func (r *Renderer) ProgramExtension() string {
	return "wgpu.json"
}

// NewProgram creates a new program from spec data.
func (r *Renderer) NewProgram(name string, data []byte) *Program {
	return loadProgram(name, data)
}

// NewTexture creates a new texture from raw data.
func (r *Renderer) NewTexture(d TextureDescriptor, data []byte) *Texture {
	format := sizedFormatToGPU(d.SizedFormat)

	mipLevels := uint32(1)
	if d.Mipmaps {
		mipLevels = uint32(math.Log2(float64(min(d.Width, d.Height)))) + 1
	}

	usage := gpu.TextureUsageTextureBinding | gpu.TextureUsageCopyDst
	// Textures that will be used as framebuffer attachments need RenderAttachment
	if d.Format == TextureFormatDEPTH || !d.Mipmaps {
		usage |= gpu.TextureUsageRenderAttachment
	}

	tex := r.device.CreateTexture(gpu.TextureDescriptor{
		Size:      gpu.Extent3D{Width: d.Width, Height: d.Height, DepthOrArrayLayers: 1},
		Format:    format,
		Usage:     gpu.TextureUsage(usage),
		Dimension: gpu.TextureDimension2D,
		MipLevels: mipLevels,
	})

	if data != nil {
		bytesPerPixel := bytesPerPixelForFormat(d.SizedFormat)
		r.queue.WriteTexture(
			gpu.ImageCopyTexture{Texture: tex, MipLevel: 0},
			unsafe.Pointer(&data[0]),
			uint64(len(data)),
			gpu.TextureDataLayout{BytesPerRow: d.Width * bytesPerPixel, RowsPerImage: d.Height},
			gpu.Extent3D{Width: d.Width, Height: d.Height, DepthOrArrayLayers: 1},
		)
	}

	view := tex.CreateView()
	sampler := r.createSampler(d)

	return &Texture{texture: tex, view: view, sampler: sampler, descriptor: d, id: allocateTextureID()}
}

// NewTextureFromImageData creates a texture from encoded image bytes.
func (r *Renderer) NewTextureFromImageData(data []byte, d TextureDescriptor) *Texture {
	if data == nil {
		glog.Fatal("Cannot read texture: nil data")
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		glog.Warning("Cannot decode texture image: ", err)
		return nil
	}

	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)

	d.Width = uint32(rgba.Rect.Size().X)
	d.Height = uint32(rgba.Rect.Size().Y)
	d.Target = TextureTarget2D
	d.Format = TextureFormatRGBA
	d.SizedFormat = TextureSizedFormatRGBA8
	d.ComponentType = TextureComponentTypeUNSIGNEDBYTE

	return r.NewTexture(d, rgba.Pix)
}

// NewFramebuffer creates a new framebuffer.
func (r *Renderer) NewFramebuffer() *Framebuffer {
	return NewFramebuffer()
}

// NewMesh creates a new mesh.
func (r *Renderer) NewMesh() *Mesh {
	return NewMesh()
}

// NewUniform creates a new uniform.
func (r *Renderer) NewUniform() *Uniform {
	return NewUniform()
}

// NewUniformBuffer creates a new uniform buffer.
func (r *Renderer) NewUniformBuffer() *UniformBuffer {
	return NewUniformBuffer()
}

// CanBatch returns whether two materials can be batched (same textures).
func (r *Renderer) CanBatch(a *Material, b *Material) bool {
	return a.sortKey == b.sortKey
}

// BeginFrame acquires the swap chain texture and creates a command encoder.
func (r *Renderer) BeginFrame() {
	st := r.surface.GetCurrentTexture()
	if st.Status != gpu.SurfaceGetCurrentTextureStatusSuccessOptimal &&
		st.Status != gpu.SurfaceGetCurrentTextureStatusSuccessSuboptimal {
		glog.Warning("Failed to acquire swap chain texture, skipping frame")
		return
	}
	r.swapChainTexture = st.Texture
	r.swapChainView = r.swapChainTexture.CreateView()
	r.encoder = r.device.CreateCommandEncoder()
}

// BeginRenderPass creates a new render pass from the descriptor.
func (r *Renderer) BeginRenderPass(desc RenderPassDescriptor) *RenderPass {
	gpuDesc := gpu.RenderPassDescriptor{}

	var colorFormats []gpu.TextureFormat
	depthFormat := gpu.TextureFormatUndefined

	for _, ca := range desc.ColorAttachments {
		colorView := ca.View
		if colorView == (gpu.TextureView{}) {
			colorView = r.swapChainView
		}
		colorFmt := ca.Format
		if colorFmt == gpu.TextureFormatUndefined {
			colorFmt = r.surfaceFormat
		}
		colorFormats = append(colorFormats, colorFmt)

		clearColor := gpu.Color{
			R: float64(ca.ClearColor[0]),
			G: float64(ca.ClearColor[1]),
			B: float64(ca.ClearColor[2]),
			A: float64(ca.ClearColor[3]),
		}

		gpuDesc.ColorAttachments = append(gpuDesc.ColorAttachments, gpu.RenderPassColorAttachment{
			View:       colorView,
			LoadOp:     ca.LoadOp,
			StoreOp:    gpu.StoreOpStore,
			ClearValue: clearColor,
		})
	}

	if desc.DepthAttachment != nil {
		depthFormat = desc.DepthAttachment.Format
		if depthFormat == gpu.TextureFormatUndefined {
			depthFormat = gpu.TextureFormatDepth32Float
		}
		gpuDesc.DepthStencilAttachment = &gpu.RenderPassDepthStencilAttachment{
			View:            desc.DepthAttachment.View,
			DepthLoadOp:     desc.DepthAttachment.LoadOp,
			DepthStoreOp:    gpu.StoreOpStore,
			DepthClearValue: desc.DepthAttachment.ClearDepth,
		}
	}

	enc := r.encoder.BeginRenderPass(gpuDesc)

	rp := &RenderPass{
		encoder:      enc,
		colorFormats: colorFormats,
		depthFormat:  depthFormat,
	}

	if desc.Viewport != (mgl32.Vec4{}) {
		rp.SetViewport(desc.Viewport[0], desc.Viewport[1], desc.Viewport[2], desc.Viewport[3])
	}

	return rp
}

// Flush submits the current command encoder and starts a new one.
// Use between draw sequences that write to the same buffers.
func (r *Renderer) Flush() {
	cmdBuf := r.encoder.Finish()
	r.queue.Submit(cmdBuf)
	r.encoder = r.device.CreateCommandEncoder()
}

// EndFrame submits the command buffer and presents.
func (r *Renderer) EndFrame() {
	cmdBuf := r.encoder.Finish()
	r.queue.Submit(cmdBuf)
	r.surface.Present()

	r.swapChainView.Release()
	r.swapChainView = gpu.TextureView{}
	r.swapChainTexture = gpu.Texture{}
}

func (r *Renderer) createSampler(d TextureDescriptor) gpu.Sampler {
	desc := gpu.SamplerDescriptor{}

	switch d.Filter {
	case TextureFilterNearest:
		desc.MinFilter = gpu.FilterModeNearest
		desc.MagFilter = gpu.FilterModeNearest
		desc.MipmapFilter = gpu.MipmapFilterModeNearest
	case TextureFilterLinear:
		desc.MinFilter = gpu.FilterModeLinear
		desc.MagFilter = gpu.FilterModeLinear
		desc.MipmapFilter = gpu.MipmapFilterModeNearest
	case TextureFilterMipmapLinear:
		desc.MinFilter = gpu.FilterModeLinear
		desc.MagFilter = gpu.FilterModeLinear
		desc.MipmapFilter = gpu.MipmapFilterModeLinear
	}

	switch d.WrapMode {
	case TextureWrapModeClampEdge, TextureWrapModeClampBorder:
		desc.AddressModeU = gpu.AddressModeClampToEdge
		desc.AddressModeV = gpu.AddressModeClampToEdge
		desc.AddressModeW = gpu.AddressModeClampToEdge
	case TextureWrapModeRepeat:
		desc.AddressModeU = gpu.AddressModeRepeat
		desc.AddressModeV = gpu.AddressModeRepeat
		desc.AddressModeW = gpu.AddressModeRepeat
	}

	// Depth textures get a comparison sampler for hardware shadow mapping
	if d.Format == TextureFormatDEPTH {
		desc.Compare = gpu.CompareFunctionLessEqual
	}

	return r.device.CreateSampler(desc)
}

func sizedFormatToGPU(f TextureSizedFormat) gpu.TextureFormat {
	switch f {
	case TextureSizedFormatR8:
		return gpu.TextureFormatR8Unorm
	case TextureSizedFormatR16F:
		return gpu.TextureFormatR16Float
	case TextureSizedFormatR32F:
		return gpu.TextureFormatR32Float
	case TextureSizedFormatRG8:
		return gpu.TextureFormatRG8Unorm
	case TextureSizedFormatRG16F:
		return gpu.TextureFormatRG16Float
	case TextureSizedFormatRG32F:
		return gpu.TextureFormatRG32Float
	case TextureSizedFormatRGBA8:
		return gpu.TextureFormatRGBA8Unorm
	case TextureSizedFormatRGBA16F:
		return gpu.TextureFormatRGBA16Float
	case TextureSizedFormatRGBA32F:
		return gpu.TextureFormatRGBA32Float
	case TextureSizedFormatDEPTH32F:
		return gpu.TextureFormatDepth32Float
	default:
		glog.Fatalf("Unsupported texture format: %d", f)
		return gpu.TextureFormatUndefined
	}
}

func bytesPerPixelForFormat(f TextureSizedFormat) uint32 {
	switch f {
	case TextureSizedFormatR8:
		return 1
	case TextureSizedFormatRG8:
		return 2
	case TextureSizedFormatRGB8:
		return 3
	case TextureSizedFormatRGBA8:
		return 4
	case TextureSizedFormatR16F:
		return 2
	case TextureSizedFormatRG16F:
		return 4
	case TextureSizedFormatRGBA16F:
		return 8
	case TextureSizedFormatR32F:
		return 4
	case TextureSizedFormatRG32F:
		return 8
	case TextureSizedFormatRGBA32F:
		return 16
	case TextureSizedFormatDEPTH32F:
		return 4
	default:
		return 4
	}
}

// DefaultRenderTechnique does z pre-pass, opaque pass, transparency pass.
func DefaultRenderTechnique(camera *Camera, materialBuckets map[*Pipeline][]*Node) {
	// Z-prepass
	zDesc := camera.MakeRenderPassDescriptor(true, true)
	zPass := renderer.BeginRenderPass(zDesc)
	for m, nodes := range materialBuckets {
		if len(nodes) == 0 || m.Blending {
			continue
		}
		zpassState := resourceManager.Pipeline(fmt.Sprintf("%s-z", nodes[0].pipeline.Name))
		if !zPass.SetPipeline(zpassState) {
			continue
		}
		zPass.SetCameraConstants(camera.constants.buffer)
		RenderBatchedNodes(zPass, camera, nodes)
	}
	zPass.End()

	// Flush — opaque pass reuses the same mesh instance buffers
	renderer.Flush()

	// Opaque pass
	opaqueDesc := camera.MakeRenderPassDescriptor(false, false)
	opaquePass := renderer.BeginRenderPass(opaqueDesc)
	for m, nodes := range materialBuckets {
		if len(nodes) == 0 || m.Blending {
			continue
		}
		if !opaquePass.SetPipeline(nodes[0].pipeline) {
			continue
		}
		opaquePass.SetCameraConstants(camera.constants.buffer)
		RenderBatchedNodes(opaquePass, camera, nodes)
	}

	// Transparent pass (same render pass as opaque)
	for m, nodes := range materialBuckets {
		if len(nodes) == 0 || !m.Blending {
			continue
		}
		if !opaquePass.SetPipeline(nodes[0].pipeline) {
			continue
		}
		opaquePass.SetCameraConstants(camera.constants.buffer)
		RenderBatchedNodes(opaquePass, camera, nodes)
	}
	opaquePass.End()
}

// AABBRenderTechnique draws AABBs for all nodes.
func AABBRenderTechnique(camera *Camera, materialBuckets map[*Pipeline][]*Node) {
	desc := camera.MakeRenderPassDescriptor(false, false)
	pass := renderer.BeginRenderPass(desc)
	pass.SetPipeline(resourceManager.Pipeline("aabb"))
	pass.SetCameraConstants(camera.constants.buffer)

	// Collect all nodes across all buckets into a single instance array
	var instanceData [MaxInstances]InstanceData
	count := 0
	for _, nodes := range materialBuckets {
		for _, n := range nodes {
			if count >= MaxInstances {
				break
			}
			center := n.worldBounds.Center()
			size := n.worldBounds.Size()
			transform64 := mgl64.Translate3D(center[0], center[1], center[2]).Mul4(mgl64.Scale3D(size[0], size[1], size[2]))
			instanceData[count].ModelMatrix = Mat4DoubleToFloat(transform64)
			count++
		}
	}
	if count > 0 {
		AABBMesh().DrawInstanced(pass, count, unsafe.Pointer(&instanceData))
	}
	pass.End()
}

// DebugRenderTechnique runs the default render path followed by AABBs.
func DebugRenderTechnique(camera *Camera, materialBuckets map[*Pipeline][]*Node) {
	DefaultRenderTechnique(camera, materialBuckets)
	AABBRenderTechnique(camera, materialBuckets)
}

// PostProcessRenderTechnique renders a single pass with no z-prepass.
// Suitable for fullscreen post-process passes (FXAA, SSGI, etc.).
func PostProcessRenderTechnique(camera *Camera, materialBuckets map[*Pipeline][]*Node) {
	desc := camera.MakeRenderPassDescriptor(
		camera.clearMode&ClearColor != 0,
		camera.clearMode&ClearDepth != 0,
	)
	pass := renderer.BeginRenderPass(desc)
	for _, nodes := range materialBuckets {
		if len(nodes) == 0 {
			continue
		}
		if !pass.SetPipeline(nodes[0].pipeline) {
			continue
		}
		pass.SetCameraConstants(camera.constants.buffer)
		RenderBatchedNodes(pass, camera, nodes)
	}
	pass.End()
}

// RenderBatchedNodes splits a list of nodes into batched items.
func RenderBatchedNodes(pass *RenderPass, camera *Camera, nodes []*Node) {
	lastBatchIndex := 0
	for i := 1; i < len(nodes); i++ {
		if !renderer.CanBatch(nodes[i].Material(), nodes[i-1].Material()) {
			RenderBatch(pass, camera, nodes[lastBatchIndex:i])
			lastBatchIndex = i
		}
	}
	RenderBatch(pass, camera, nodes[lastBatchIndex:])
}

// RenderBatch renders a batch of drawables sharing the same state and descriptors.
func RenderBatch(pass *RenderPass, camera *Camera, nodes []*Node) {
	if len(nodes) == 0 {
		return
	}

	var instanceData [MaxInstances]InstanceData
	for i, n := range nodes {
		mMatrix64 := n.WorldTransform()
		mvpMatrix64 := camera.projectionMatrix.Mul4(camera.viewMatrix.Mul4(mMatrix64))
		instanceData[i].ModelMatrix = Mat4DoubleToFloat(mMatrix64)
		instanceData[i].ModelViewProjectionMatrix = Mat4DoubleToFloat(mvpMatrix64)
		instanceData[i].Custom = n.material.instanceData
	}
	pass.SetMaterial(&nodes[0].material)
	nodes[0].mesh.DrawInstanced(pass, len(nodes), unsafe.Pointer(&instanceData))
}
