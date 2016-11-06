package core

import (
	"fmt"
	"unsafe"

	"github.com/fcvarela/gosg/protos"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
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
	Framebuffer Framebuffer
}

// ClearCommand clears the current framebuffer
type ClearCommand struct {
	ClearMode  ClearMode
	ClearColor mgl32.Vec4
	ClearDepth float64
}

// BindStateCommand binds the passed state
type BindStateCommand struct {
	State *protos.State
}

// BindUniformBufferCommand binds the passed uniform buffer
type BindUniformBufferCommand struct {
	Name          string
	UniformBuffer UniformBuffer
}

// BindDescriptorsCommand bids the passed descriptors
type BindDescriptorsCommand struct {
	Descriptors *Descriptors
}

// DrawInstancedCommand draws a mesh using instancing
type DrawInstancedCommand struct {
	Mesh          Mesh
	InstanceCount int
	ModelMatrices unsafe.Pointer
}

// DrawCommand draws a mesh
type DrawCommand struct {
	Mesh Mesh
}

// SwapBuffersCommand swaps the front/back buffers of a framebuffer
type SwapBuffersCommand struct {
	Window *glfw.Window
}

// RenderSystem is an interface which wraps all logic related to rendering and memory management of
// GPU buffers.
type RenderSystem interface {
	// PrepareWindow initializes a new window
	PreMakeWindow()
	PostMakeWindow(cfg WindowConfig, window *glfw.Window)

	// NewMesh retuns a new mesh.
	NewMesh() Mesh

	// NewIMGUIMesh returns a new IMGUI mesh.
	NewIMGUIMesh() IMGUIMesh

	// ProgramExtension exposes the resource extension of program definitions for the implementation.
	ProgramExtension() string

	// NewProgram creates a new program from a list of subprogram source files.
	NewProgram(name string, data []byte) Program

	// NewTexture creates a new texture from a byte buffer containing an image file, not raw bitmap.
	// This always generates RGBA, unsigned byte and will generate mipmaps levels from
	// smallest dimension, ie: 2048x1024 = 10 mipmap levels; log2(1024)
	// It also defaults to ClampEdge and mipmapped filtering.
	NewTextureFromImageData(r []byte, d TextureDescriptor) Texture

	// NewUniform creates a new empty uniform
	NewUniform() Uniform

	// NewUniformBuffer creates a new empty uniform buffer
	NewUniformBuffer() UniformBuffer

	// NewRawTexture creates a new texture and allocates storage for it
	NewTexture(descriptor TextureDescriptor, data []byte) Texture

	// NewFramebuffer returns a newly created framebuffer
	NewFramebuffer() Framebuffer

	// Run runs a command. When sync is true, the call blocks for a return error. Async always returns nil
	Run(cmd RenderCommand, sync bool) error

	// CommandQueue returns the command queue that the rendersystem consumes every frame
	CommandQueue() chan RenderCommand

	// FlushQueue tells the renderSystem to flush its queue until it gets a nil command
	Flush()

	// CanBatch returns whether two nodes can be batched in the same drawcall
	CanBatch(a *Descriptors, b *Descriptors) bool

	// RenderLog returns a log of the render plan
	RenderLog() string
}

// RenderPass represents one operation of rendering a set of nodes with a given rasterstate and a program
type RenderPass struct {
	Nodes []*Node
	Name  string
	State *protos.State
}

var (
	renderSystem RenderSystem
)

// InstanceData holds the per-instance-data when using instanced drawing. It is automatically
// populated by all the included RenderTechniques. Use node.GetInstanceData() to change its custom
// values
type InstanceData struct {
	ModelMatrix mgl32.Mat4
	Custom1     mgl32.Vec4
	Custom2     mgl32.Vec4
	Custom3     mgl32.Vec4
	Custom4     mgl32.Vec4
}

const (
	// MaxInstances is the maximum number of modelMatrices in a mesh instance attribute buffer
	MaxInstances = 2000

	// InstanceDataLen is the byte size of an InstanceData value
	InstanceDataLen = (1*16 + 4*4) * 4
)

// SetRenderSystem is meant to be called from RenderSystem implementations on their init method
// to cause the side-effect of setting the core RenderSystem to their provided one.
func SetRenderSystem(rs RenderSystem) {
	renderSystem = rs
}

// GetRenderSystem returns the renderSystem, thereby exposing it to any package importing core.
func GetRenderSystem() RenderSystem {
	return renderSystem
}

// DefaultRenderTechnique does z pre-pass, diffuse pass, transparency pass
func DefaultRenderTechnique(camera *Camera, materialBuckets map[*protos.State][]*Node, cmdBuf chan RenderCommand) {
	cmdBuf <- &SetFramebufferCommand{camera.framebuffer}
	cmdBuf <- &SetViewportCommand{camera.viewport}
	cmdBuf <- &ClearCommand{camera.clearMode, camera.clearColor, camera.clearDepth}

	// draw zpass
	for m, nodes := range materialBuckets {
		if len(nodes) == 0 {
			continue
		}
		if m.Blending == true {
			continue
		}

		var zpassState = resourceManager.State(fmt.Sprintf("%s-z", nodes[0].state.Name))
		cmdBuf <- &BindStateCommand{zpassState}
		cmdBuf <- &BindUniformBufferCommand{"cameraConstants", camera.constants.buffer}
		RenderBatchedNodes(nodes, cmdBuf)
	}

	// draw opaque pass
	for m, nodes := range materialBuckets {
		if len(nodes) == 0 {
			continue
		}
		if m.Blending == true {
			continue
		}

		cmdBuf <- &BindStateCommand{nodes[0].state}
		cmdBuf <- &BindUniformBufferCommand{"cameraConstants", camera.constants.buffer}
		RenderBatchedNodes(nodes, cmdBuf)
	}

	// transparent pass
	for m, nodes := range materialBuckets {
		if len(nodes) == 0 {
			continue
		}
		if m.Blending == false {
			continue
		}
		//sort.Sort(sort.Reverse(NodesByCameraDistanceNearToFar{nodes, camera.node}))
		cmdBuf <- &BindStateCommand{nodes[0].state}
		cmdBuf <- &BindUniformBufferCommand{"cameraConstants", camera.constants.buffer}
		RenderBatchedNodes(nodes, cmdBuf)
	}
}

// AABBRenderTechnique draws AABBs for all nodes
func AABBRenderTechnique(camera *Camera, materialBuckets map[*protos.State][]*Node, cmdBuf chan RenderCommand) {
	cmdBuf <- &BindStateCommand{resourceManager.State("aabb")}
	cmdBuf <- &BindUniformBufferCommand{"cameraConstants", camera.constants.buffer}
	for _, nodes := range materialBuckets {
		if len(nodes) == 0 {
			continue
		}

		var instanceData [MaxInstances]InstanceData
		for i, n := range nodes {
			var center = n.worldBounds.Center()
			var size = n.worldBounds.Size()

			// need a setuniform command
			var transform64 = mgl64.Translate3D(center[0], center[1], center[2]).Mul4(mgl64.Scale3D(size[0], size[1], size[2]))
			var transform32 = Mat4DoubleToFloat(transform64)
			instanceData[i].ModelMatrix = transform32
		}
		cmdBuf <- &DrawInstancedCommand{Mesh: AABBMesh(), InstanceCount: len(nodes), ModelMatrices: unsafe.Pointer(&instanceData)}
	}
}

// DebugRenderTechnique runs the default render path followed by AABBs
func DebugRenderTechnique(camera *Camera, materialBuckets map[*protos.State][]*Node, cmdBuf chan RenderCommand) {
	DefaultRenderTechnique(camera, materialBuckets, cmdBuf)
	AABBRenderTechnique(camera, materialBuckets, cmdBuf)
}

// RenderBatchedNodes splits a list of nodes into batched items per buffers and streams draw calls
// It assumes all batches share materials and should be used long with Mater
func RenderBatchedNodes(nodes []*Node, cmdBuf chan RenderCommand) {
	var lastBatchIndex = 0
	for i := 1; i < len(nodes); i++ {
		if renderSystem.CanBatch(nodes[i].MaterialData(), nodes[i-1].MaterialData()) {
			RenderBatch(nodes[lastBatchIndex:i], cmdBuf)
			lastBatchIndex = i
		}
	}

	// close last batch
	RenderBatch(nodes[lastBatchIndex:], cmdBuf)
}

// RenderBatch renders a batch of drawables sharing the same state and descriptors
func RenderBatch(nodes []*Node, cmdBuf chan RenderCommand) {
	if len(nodes) == 0 {
		return
	}

	var instanceData [MaxInstances]InstanceData
	for i, n := range nodes {
		transform64 := n.WorldTransform()
		transform32 := Mat4DoubleToFloat(transform64)
		instanceData[i].ModelMatrix = transform32
		instanceData[i].Custom1 = n.materialData.custom1
		instanceData[i].Custom2 = n.materialData.custom2
		instanceData[i].Custom3 = n.materialData.custom3
		instanceData[i].Custom4 = n.materialData.custom4
	}
	cmdBuf <- &BindDescriptorsCommand{
		Descriptors: &nodes[0].materialData,
	}
	cmdBuf <- &DrawInstancedCommand{Mesh: nodes[0].mesh, InstanceCount: len(nodes), ModelMatrices: unsafe.Pointer(&instanceData)}
}
