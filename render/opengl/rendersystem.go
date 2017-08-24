package opengl

import (
	"github.com/fcvarela/gosg/core"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/golang/glog"
)

// RenderSystem implements the core.RenderSystem interface
type RenderSystem struct {
	renderLog string
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
		renderLog: "",
	}

	return renderSystem
}

// RenderLog implements the core.RenderSystem interface
func (r *RenderSystem) RenderLog() string {
	return r.renderLog
}

// Dispatch implements the core.RenderSystem interface
func (r *RenderSystem) Dispatch(cmd core.RenderCommand) {
	switch t := cmd.(type) {
	case *core.DrawInstancedCommand:
		r.drawInstanced(t)
	case *core.DrawCommand:
		r.draw(t)
	case *core.DrawIMGUICommand:
		r.drawIMGUI(t)
	case *core.BindDescriptorsCommand:
		r.bindDescriptors(t)
	case *core.BindStateCommand:
		r.bindState(t)
	case *core.ClearCommand:
		r.clear(t)
	case *core.BindUniformBufferCommand:
		currentProgram.setUniformBufferByName(t.Name, t.UniformBuffer.(*UniformBuffer))
	case *core.SetFramebufferCommand:
		r.setFramebuffer(t)
	case *core.SetViewportCommand:
		r.setViewport(t)
	default:
		glog.Errorf("Unsupported command type: %T", t)
	}
}

func (r *RenderSystem) setFramebuffer(cmd *core.SetFramebufferCommand) {
	if cmd.Framebuffer != nil {
		bindFramebuffer(cmd.Framebuffer.(*Framebuffer).fbo)
	} else {
		bindFramebuffer(0)
	}
}

func (r *RenderSystem) setViewport(cmd *core.SetViewportCommand) {
	v := cmd.Viewport
	gl.Viewport(int32(v[0]), int32(v[1]), int32(v[2]), int32(v[3]))
}

func (r *RenderSystem) clear(cmd *core.ClearCommand) {
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
}

func (r *RenderSystem) bindState(cmd *core.BindStateCommand) {
	bindState(cmd.State, false)
}

func (r *RenderSystem) bindDescriptors(cmd *core.BindDescriptorsCommand) {
	bindTextures(currentProgram, cmd.Descriptors)
	bindUniformBuffers(currentProgram, cmd.Descriptors)
	bindUniforms(currentProgram, cmd.Descriptors)
}

func (r *RenderSystem) draw(cmd *core.DrawCommand) {
	cmd.Mesh.Draw()
}

func (r *RenderSystem) drawInstanced(cmd *core.DrawInstancedCommand) {
	cmd.Mesh.DrawInstanced(cmd.InstanceCount, cmd.InstanceData)
}
