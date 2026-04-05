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

// RenderCommand is a generic render command item
type RenderCommand interface{}

// DebugCommand logs a message
type DebugCommand struct {
	Message string
}

// SetViewportCommand sets the current viewport
type SetViewportCommand struct {
	Viewport mgl32.Vec4
}

// SetFramebufferCommand sets the current framebuffer
type SetFramebufferCommand struct {
	Framebuffer *Framebuffer
}

// ClearCommand clears the current framebuffer
type ClearCommand struct {
	ClearMode  ClearMode
	ClearColor mgl32.Vec4
	ClearDepth float64
}

// BindStateCommand binds the passed state
type BindStateCommand struct {
	State *State
}

// BindUniformBufferCommand binds the passed uniform buffer
type BindUniformBufferCommand struct {
	Name          string
	UniformBuffer *UniformBuffer
}

// BindDescriptorsCommand binds the passed descriptors
type BindDescriptorsCommand struct {
	Descriptors *Descriptors
}

// DrawInstancedCommand draws a mesh using instancing
type DrawInstancedCommand struct {
	Mesh          *Mesh
	InstanceCount int
	InstanceData  unsafe.Pointer
}

// DrawIMGUICommand draws an IMGUI
type DrawIMGUICommand struct{}

// DrawCommand draws a mesh
type DrawCommand struct {
	Mesh *Mesh
}

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
	currentRenderPass gpu.RenderPassEncoder
	renderPassActive  bool

	// Pending state for deferred render pass begin
	pendingFramebuffer *Framebuffer
	pendingViewport    mgl32.Vec4
	pendingClear       *ClearCommand
	pipelineSet        bool

	// Current bound state
	currentProgram *Program
	pipelines      *pipelineCache
	defaultTexture      *Texture
	defaultDepthTexture *Texture

	renderLog string
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

// RenderLog returns the render log.
func (r *Renderer) RenderLog() string {
	return r.renderLog
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

	return &Texture{texture: tex, view: view, sampler: sampler, descriptor: d}
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

// Dispatch executes a render command.
func (r *Renderer) Dispatch(cmd RenderCommand) {
	switch t := cmd.(type) {
	case *DebugCommand:
		glog.Info(t.Message)

	case *SetFramebufferCommand:
		// End current render pass when framebuffer changes
		r.endCurrentRenderPass()
		r.pendingFramebuffer = t.Framebuffer
		r.pendingClear = nil

	case *SetViewportCommand:
		r.pendingViewport = t.Viewport
		if r.renderPassActive {
			r.currentRenderPass.SetViewport(
				t.Viewport[0], t.Viewport[1],
				t.Viewport[2], t.Viewport[3],
				0.0, 1.0,
			)
		}

	case *ClearCommand:
		r.pendingClear = t

	case *BindStateCommand:
		r.ensureRenderPass()
		if t.State.ProgramName == "" {
			break
		}
		program := resourceManager.Program(t.State.ProgramName)
		if program == nil {
			break
		}
		r.currentProgram = program

		// Determine current target formats
		colorFormat := r.surfaceFormat
		depthFormat := gpu.TextureFormatUndefined
		if r.pendingFramebuffer != nil {
			if len(r.pendingFramebuffer.colorAttachments) > 0 {
				colorFormat = sizedFormatToGPU(r.pendingFramebuffer.colorAttachments[0].descriptor.SizedFormat)
			} else {
				colorFormat = gpu.TextureFormatUndefined
			}
			if r.pendingFramebuffer.depthAttachment != nil {
				depthFormat = sizedFormatToGPU(r.pendingFramebuffer.depthAttachment.descriptor.SizedFormat)
			}
		}

		pipeline := r.pipelines.getOrCreate(t.State, program, colorFormat, depthFormat)
		r.currentRenderPass.SetPipeline(pipeline)
		r.pipelineSet = true

	case *BindUniformBufferCommand:
		// Create a bind group for the UBO at group 0
		if r.currentProgram == nil || !r.renderPassActive || t.UniformBuffer == nil {
			break
		}
		if len(r.currentProgram.bindGroupLayouts) == 0 {
			break
		}
		bg := renderer.device.CreateBindGroup(r.currentProgram.bindGroupLayouts[0], []gpu.BindGroupEntry{{
			Binding: 0,
			Buffer:  t.UniformBuffer.buffer,
			Offset:  0,
			Size:    t.UniformBuffer.size,
		}})
		r.currentRenderPass.SetBindGroup(0, bg)

	case *BindDescriptorsCommand:
		// Create bind group for textures at group 1
		if r.currentProgram == nil || !r.renderPassActive {
			break
		}
		if len(r.currentProgram.bindGroupLayouts) < 2 || len(r.currentProgram.spec.TextureBindings) == 0 {
			break
		}
		entries := make([]gpu.BindGroupEntry, 0, len(r.currentProgram.spec.TextureBindings)*2)
		for texName, binding := range r.currentProgram.spec.TextureBindings {
			tex := t.Descriptors.Texture(texName)
			if tex == nil {
				// Pick default based on the bind group layout's expected sample type
				isDepth := false
				if len(r.currentProgram.spec.BindGroupLayouts) > int(binding.Group) {
					for _, e := range r.currentProgram.spec.BindGroupLayouts[binding.Group].Entries {
						if e.Binding == binding.TextureBinding && e.Texture != nil && e.Texture.SampleType == "depth" {
							isDepth = true
						}
					}
				}
				if isDepth {
					tex = r.defaultDepthTexture
				} else {
					tex = r.defaultTexture
				}
			}
			entries = append(entries,
				gpu.BindGroupEntry{Binding: binding.TextureBinding, TextureView: tex.view},
				gpu.BindGroupEntry{Binding: binding.SamplerBinding, Sampler: tex.sampler},
			)
		}
		if len(entries) > 0 {
			bg := renderer.device.CreateBindGroup(r.currentProgram.bindGroupLayouts[1], entries)
			r.currentRenderPass.SetBindGroup(1, bg)
		}

	case *DrawInstancedCommand:
		r.ensureRenderPass()
		if t.Mesh != nil {
			t.Mesh.DrawInstanced(t.InstanceCount, t.InstanceData)
		}

	case *DrawCommand:
		r.ensureRenderPass()
		if t.Mesh != nil {
			t.Mesh.Draw()
		}

	case *DrawIMGUICommand:
		r.ensureRenderPass()
		if r.currentProgram != nil {
			imgui.draw(r.currentProgram)
		}

	default:
		glog.Errorf("Unsupported command type: %T", t)
	}
}

// CanBatch returns whether two descriptors can be batched.
func (r *Renderer) CanBatch(a *Descriptors, b *Descriptors) bool {
	for name, tb := range b.Textures() {
		ta, ok := a.Textures()[name]
		if !ok || ta != tb {
			return false
		}
	}
	return true
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
	r.renderPassActive = false
	r.pendingClear = nil
	r.pendingFramebuffer = nil
}

// Flush ends the current render pass, submits, and starts a new command encoder.
// Use between draw sequences that write to the same buffers.
func (r *Renderer) Flush() {
	r.endCurrentRenderPass()
	cmdBuf := r.encoder.Finish()
	r.queue.Submit(cmdBuf)
	r.encoder = r.device.CreateCommandEncoder()
}

// EndFrame ends the current render pass, submits, and presents.
func (r *Renderer) EndFrame() {
	r.endCurrentRenderPass()

	cmdBuf := r.encoder.Finish()
	r.queue.Submit(cmdBuf)
	r.surface.Present()

	r.swapChainView.Release()
	r.swapChainView = gpu.TextureView{}
	r.swapChainTexture = gpu.Texture{}
}

func (r *Renderer) endCurrentRenderPass() {
	if r.renderPassActive {
		r.currentRenderPass.End()
		r.currentRenderPass.Release()
		r.currentRenderPass = gpu.RenderPassEncoder{}
		r.renderPassActive = false
		r.pipelineSet = false
	}
}

func (r *Renderer) ensureRenderPass() {
	if r.renderPassActive {
		return
	}

	desc := gpu.RenderPassDescriptor{}

	// Determine if we have a color attachment
	hasColor := true
	if r.pendingFramebuffer != nil && len(r.pendingFramebuffer.colorAttachments) == 0 {
		hasColor = false
	}

	if hasColor {
		var colorView gpu.TextureView
		if r.pendingFramebuffer != nil && len(r.pendingFramebuffer.colorAttachments) > 0 {
			colorView = r.pendingFramebuffer.colorAttachments[0].view
		} else {
			colorView = r.swapChainView
		}

		colorLoadOp := gpu.LoadOpLoad
		colorClear := gpu.Color{R: 0, G: 0, B: 0, A: 1}
		if r.pendingClear != nil && r.pendingClear.ClearMode&ClearColor != 0 {
			colorLoadOp = gpu.LoadOpClear
			colorClear = gpu.Color{
				R: float64(r.pendingClear.ClearColor[0]),
				G: float64(r.pendingClear.ClearColor[1]),
				B: float64(r.pendingClear.ClearColor[2]),
				A: float64(r.pendingClear.ClearColor[3]),
			}
		}

		desc.ColorAttachments = []gpu.RenderPassColorAttachment{{
			View:       colorView,
			LoadOp:     colorLoadOp,
			StoreOp:    gpu.StoreOpStore,
			ClearValue: colorClear,
		}}
	}

	// Depth attachment
	var depthView gpu.TextureView
	if r.pendingFramebuffer != nil && r.pendingFramebuffer.depthAttachment != nil {
		depthView = r.pendingFramebuffer.depthAttachment.view
	}
	if depthView != (gpu.TextureView{}) {
		depthLoadOp := gpu.LoadOpLoad
		depthClearValue := float32(1.0)
		if r.pendingClear != nil && r.pendingClear.ClearMode&ClearDepth != 0 {
			depthLoadOp = gpu.LoadOpClear
			depthClearValue = float32(r.pendingClear.ClearDepth)
		}
		desc.DepthStencilAttachment = &gpu.RenderPassDepthStencilAttachment{
			View:            depthView,
			DepthLoadOp:     depthLoadOp,
			DepthStoreOp:    gpu.StoreOpStore,
			DepthClearValue: depthClearValue,
		}
	}

	r.currentRenderPass = r.encoder.BeginRenderPass(desc)
	r.renderPassActive = true

	// Apply viewport
	if r.pendingViewport != (mgl32.Vec4{}) {
		r.currentRenderPass.SetViewport(
			r.pendingViewport[0], r.pendingViewport[1],
			r.pendingViewport[2], r.pendingViewport[3],
			0.0, 1.0,
		)
	}

	r.pendingClear = nil
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

// DefaultRenderTechnique does z pre-pass, diffuse pass, transparency pass
func DefaultRenderTechnique(camera *Camera, materialBuckets map[*State][]*Node) {
	renderer.Dispatch(&SetFramebufferCommand{camera.framebuffer})
	renderer.Dispatch(&SetViewportCommand{camera.viewport})
	renderer.Dispatch(&ClearCommand{camera.clearMode, camera.clearColor, camera.clearDepth})

	// draw zpass
	for m, nodes := range materialBuckets {
		if len(nodes) == 0 || m.Blending {
			continue
		}
		var zpassState = resourceManager.State(fmt.Sprintf("%s-z", nodes[0].state.Name))
		renderer.Dispatch(&BindStateCommand{zpassState})
		renderer.Dispatch(&BindUniformBufferCommand{"cameraConstants", camera.constants.buffer})
		RenderBatchedNodes(camera, nodes)
	}

	// Flush — opaque pass reuses the same mesh instance buffers
	renderer.Flush()
	renderer.Dispatch(&SetFramebufferCommand{camera.framebuffer})
	renderer.Dispatch(&SetViewportCommand{camera.viewport})

	// draw opaque pass
	for m, nodes := range materialBuckets {
		if len(nodes) == 0 || m.Blending {
			continue
		}
		renderer.Dispatch(&BindStateCommand{nodes[0].state})
		renderer.Dispatch(&BindUniformBufferCommand{"cameraConstants", camera.constants.buffer})
		RenderBatchedNodes(camera, nodes)
	}

	// transparent pass
	for m, nodes := range materialBuckets {
		if len(nodes) == 0 || !m.Blending {
			continue
		}
		renderer.Dispatch(&BindStateCommand{nodes[0].state})
		renderer.Dispatch(&BindUniformBufferCommand{"cameraConstants", camera.constants.buffer})
		RenderBatchedNodes(camera, nodes)
	}
}

// AABBRenderTechnique draws AABBs for all nodes
func AABBRenderTechnique(camera *Camera, materialBuckets map[*State][]*Node) {
	renderer.Dispatch(&BindStateCommand{resourceManager.State("aabb")})
	renderer.Dispatch(&BindUniformBufferCommand{"cameraConstants", camera.constants.buffer})

	// Collect all nodes across all buckets into a single instance array
	var instanceData [MaxInstances]InstanceData
	count := 0
	for _, nodes := range materialBuckets {
		for _, n := range nodes {
			if count >= MaxInstances {
				break
			}
			var center = n.worldBounds.Center()
			var size = n.worldBounds.Size()
			var transform64 = mgl64.Translate3D(center[0], center[1], center[2]).Mul4(mgl64.Scale3D(size[0], size[1], size[2]))
			instanceData[count].ModelMatrix = Mat4DoubleToFloat(transform64)
			count++
		}
	}
	if count > 0 {
		renderer.Dispatch(&DrawInstancedCommand{Mesh: AABBMesh(), InstanceCount: count, InstanceData: unsafe.Pointer(&instanceData)})
	}
}

// DebugRenderTechnique runs the default render path followed by AABBs
func DebugRenderTechnique(camera *Camera, materialBuckets map[*State][]*Node) {
	DefaultRenderTechnique(camera, materialBuckets)
	AABBRenderTechnique(camera, materialBuckets)
}

// RenderBatchedNodes splits a list of nodes into batched items
func RenderBatchedNodes(camera *Camera, nodes []*Node) {
	var lastBatchIndex = 0
	for i := 1; i < len(nodes); i++ {
		if !renderer.CanBatch(nodes[i].MaterialData(), nodes[i-1].MaterialData()) {
			RenderBatch(camera, nodes[lastBatchIndex:i])
			lastBatchIndex = i
		}
	}
	RenderBatch(camera, nodes[lastBatchIndex:])
}

// RenderBatch renders a batch of drawables sharing the same state and descriptors
func RenderBatch(camera *Camera, nodes []*Node) {
	if len(nodes) == 0 {
		return
	}

	var instanceData [MaxInstances]InstanceData
	for i, n := range nodes {
		mMatrix64 := n.WorldTransform()
		mvpMatrix64 := camera.projectionMatrix.Mul4(camera.viewMatrix.Mul4(mMatrix64))
		instanceData[i].ModelMatrix = Mat4DoubleToFloat(mMatrix64)
		instanceData[i].ModelViewProjectionMatrix = Mat4DoubleToFloat(mvpMatrix64)
		instanceData[i].Custom = n.materialData.instanceData
	}
	renderer.Dispatch(&BindDescriptorsCommand{Descriptors: &nodes[0].materialData})
	renderer.Dispatch(&DrawInstancedCommand{Mesh: nodes[0].mesh, InstanceCount: len(nodes), InstanceData: unsafe.Pointer(&instanceData)})
}
