package opengl

import (
	"github.com/fcvarela/gosg/core"
	"github.com/fcvarela/gosg/protos"
	"github.com/go-gl/gl/v3.3-core/gl"
)

var (
	clearState          *protos.State
	currentState        *protos.State
	textureUnitBindings = map[uint32]uint32{}
)

func bindState(state *protos.State, force bool) {
	if state.DepthTest != currentState.DepthTest || force {
		if state.DepthTest {
			gl.Enable(gl.DEPTH_TEST)
		} else {
			gl.Disable(gl.DEPTH_TEST)
		}
	}

	if state.DepthFunc != currentState.DepthFunc || force {
		switch state.DepthFunc {
		case protos.State_DEPTH_LESS:
			gl.DepthFunc(gl.LESS)
		case protos.State_DEPTH_LESS_EQUAL:
			gl.DepthFunc(gl.LEQUAL)
		case protos.State_DEPTH_EQUAL:
			gl.DepthFunc(gl.EQUAL)
		}
	}

	if state.ScissorTest != currentState.ScissorTest || force {
		if state.ScissorTest {
			gl.Enable(gl.SCISSOR_TEST)
		} else {
			gl.Disable(gl.SCISSOR_TEST)
		}
	}

	if state.Blending != currentState.Blending || force {
		if state.Blending {
			gl.Enable(gl.BLEND)
		} else {
			gl.Disable(gl.BLEND)
		}
	}

	if state.BlendSrcMode != currentState.BlendSrcMode || state.BlendDstMode != currentState.BlendDstMode || force {
		srcMode := uint32(0)
		dstMode := uint32(0)

		switch state.BlendSrcMode {
		case protos.State_BLEND_ONE:
			srcMode = gl.ONE
		case protos.State_BLEND_ONE_MINUS_SRC_ALPHA:
			srcMode = gl.ONE_MINUS_SRC_ALPHA
		case protos.State_BLEND_SRC_ALPHA:
			srcMode = gl.SRC_ALPHA
		}

		switch state.BlendDstMode {
		case protos.State_BLEND_ONE:
			dstMode = gl.ONE
		case protos.State_BLEND_ONE_MINUS_SRC_ALPHA:
			dstMode = gl.ONE_MINUS_SRC_ALPHA
		case protos.State_BLEND_SRC_ALPHA:
			dstMode = gl.SRC_ALPHA
		}
		gl.BlendFunc(srcMode, dstMode)
	}

	if state.BlendEquation != currentState.BlendEquation || force {
		switch state.BlendEquation {
		case protos.State_BLEND_FUNC_ADD:
			gl.BlendEquation(gl.FUNC_ADD)
		case protos.State_BLEND_FUNC_MAX:
			gl.BlendEquation(gl.MAX)
		}
	}

	if state.DepthWrite != currentState.DepthWrite || force {
		gl.DepthMask(state.DepthWrite)
	}

	if state.ColorWrite != currentState.ColorWrite || force {
		gl.ColorMask(state.ColorWrite, state.ColorWrite, state.ColorWrite, state.ColorWrite)
	}

	if state.Culling != currentState.Culling || force {
		if state.Culling {
			gl.Enable(gl.CULL_FACE)
		} else {
			gl.Disable(gl.CULL_FACE)
		}
	}

	if state.CullFace != currentState.CullFace || force {
		switch state.CullFace {
		case protos.State_CULL_BACK:
			gl.CullFace(gl.BACK)
		case protos.State_CULL_FRONT:
			gl.CullFace(gl.FRONT)
		case protos.State_CULL_BOTH:
			gl.CullFace(gl.FRONT_AND_BACK)
		}
	}

	if (state.ProgramName != currentState.ProgramName || force) && state.ProgramName != "" {
		core.GetResourceManager().Program(state.ProgramName).(*Program).bind()
	}

	currentState = state
}

// CanBatch determines whether or not two states can be batched
func (r *RenderSystem) CanBatch(a *core.Descriptors, b *core.Descriptors) bool {
	for name, tb := range b.Textures() {
		ta, ok := a.Textures()[name]
		if !ok {
			return false
		}

		if ta.(*Texture).id != tb.(*Texture).id {
			return false
		}
	}

	return true
}

func bindTextures(p *Program, md *core.Descriptors) {
	for name, texture := range md.Textures() {
		p.setTexture(name, texture.(*Texture))
	}
}

func bindUniformBuffers(p *Program, md *core.Descriptors) {
	for name, uniformBuffer := range md.UniformBuffers() {
		p.setUniformBufferByName(name, uniformBuffer.(*UniformBuffer))
	}
}

func bindUniforms(p *Program, md *core.Descriptors) {
	for name, uniform := range md.Uniforms() {
		p.setUniform(name, uniform.(*Uniform))
	}
}
