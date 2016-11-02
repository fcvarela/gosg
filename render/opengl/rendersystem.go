package opengl

import (
	"github.com/fcvarela/gosg/core"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/golang/glog"
)

type lambdaRenderCommand struct {
	core.RenderCommand
	fn func() error
}

// RenderSystem implements the core.RenderSystem interface
type RenderSystem struct {
	commandQueue chan core.RenderCommand
	renderLog    string
}

func init() {
	core.SetRenderSystem(New())
}

var (
	renderSystem *RenderSystem
)

// New returns a new RenderSystem
func New() *RenderSystem {
	renderSystem = &RenderSystem{
		commandQueue: make(chan (core.RenderCommand), 1000),
		renderLog:    "",
	}

	return renderSystem
}

// CommandQueue returns the RenderSystem commandQueue
func (r *RenderSystem) CommandQueue() chan core.RenderCommand {
	return r.commandQueue
}

// Run implements the core.RenderSystem interface
func (r *RenderSystem) Run(cmd core.RenderCommand, sync bool) error {
	if sync {
		return r.dispatchCommand(cmd)
	}

	// default
	r.commandQueue <- cmd
	return nil
}

// Flush implements the core.RenderSystem interface
func (r *RenderSystem) Flush() {
	for {
		var cmd = <-r.commandQueue
		if cmd == nil {
			break
		}
		r.dispatchCommand(cmd)
	}
}

// RenderLog implements the core.RenderSystem interface
func (r *RenderSystem) RenderLog() string {
	return r.renderLog
}

func (r *RenderSystem) dispatchCommand(cmd core.RenderCommand) error {
	switch t := cmd.(type) {
	case *core.DrawInstancedCommand:
		return r.drawInstanced(t)
	case *core.DrawCommand:
		return r.draw(t)
	case *core.BindDescriptorsCommand:
		return r.bindDescriptors(t)
	case *core.BindStateCommand:
		return r.bindState(t)
	case *core.SwapBuffersCommand:
		return r.swapBuffers(t)
	case *core.ClearCommand:
		return r.clear(t)
	case *core.BindUniformBufferCommand:
		currentProgram.setUniformBufferByName(t.Name, t.UniformBuffer.(*UniformBuffer))
		return nil
	case *core.SetFramebufferCommand:
		return r.setFramebuffer(t)
	case *core.SetViewportCommand:
		return r.setViewport(t)
	case *makeTextureCommand:
		return r.newTexture(t)
	case *setTextureFilterCommand:
		return t.texture.(*Texture).setFilter(t.filter)
	case *setTextureWrapModeCommand:
		return t.texture.(*Texture).setWrapMode(t.wrapMode)
	case *makeUniformBufferCommand:
		return r.makeUniformBuffer(t)
	case *makeFramebufferCommand:
		return r.makeFramebuffer(t)
	case *makeProgramCommand:
		return r.makeProgram(t)
	case *lambdaRenderCommand:
		return t.fn()
	default:
		glog.Errorf("Unsupported command type: %T", t)
	}
	return nil
}

func (r *RenderSystem) swapBuffers(cmd *core.SwapBuffersCommand) error {
	cmd.Window.SwapBuffers()
	return nil
}

func (r *RenderSystem) setFramebuffer(cmd *core.SetFramebufferCommand) error {
	if cmd.Framebuffer != nil {
		bindFramebuffer(cmd.Framebuffer.(*Framebuffer).fbo)
	} else {
		bindFramebuffer(0)
	}

	return nil
}

func (r *RenderSystem) setViewport(cmd *core.SetViewportCommand) error {
	v := cmd.Viewport
	gl.Viewport(int32(v[0]), int32(v[1]), int32(v[2]), int32(v[3]))
	return nil
}

func (r *RenderSystem) clear(cmd *core.ClearCommand) error {
	var clearargs uint32
	if cmd.ClearMode&core.ClearColor > 0 {
		clearargs = clearargs | gl.COLOR_BUFFER_BIT
		gl.ClearColor(cmd.ClearColor[0], cmd.ClearColor[1], cmd.ClearColor[2], cmd.ClearColor[3])
	}

	if cmd.ClearMode&core.ClearDepth > 0 {
		clearargs = clearargs | gl.DEPTH_BUFFER_BIT
		gl.ClearDepth(cmd.ClearDepth)
	}

	if clearargs != 0 {
		bindState(clearState, false)
		gl.Clear(clearargs)
	}

	return nil
}

func (r *RenderSystem) bindState(cmd *core.BindStateCommand) error {
	bindState(cmd.State, false)
	return nil
}

func (r *RenderSystem) bindDescriptors(cmd *core.BindDescriptorsCommand) error {
	bindTextures(currentProgram, cmd.Descriptors)
	bindUniformBuffers(currentProgram, cmd.Descriptors)
	bindUniforms(currentProgram, cmd.Descriptors)
	return nil
}

func (r *RenderSystem) draw(cmd *core.DrawCommand) error {
	cmd.Mesh.Draw()
	return nil
}

func (r *RenderSystem) drawInstanced(cmd *core.DrawInstancedCommand) error {
	cmd.Mesh.SetInstanceCount(cmd.InstanceCount)
	cmd.Mesh.SetModelMatrices(cmd.ModelMatrices)
	cmd.Mesh.Draw()

	return nil
}
