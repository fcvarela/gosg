package core

import (
	"fmt"

	"github.com/fcvarela/gosg/gpu"
	"github.com/golang/glog"
)

// pipelineKey uniquely identifies a render pipeline configuration.
type pipelineKey struct {
	programName       string
	topology          string
	depthTest         bool
	depthWrite        bool
	depthFunc         DepthFunc
	blending          bool
	blendSrcMode      BlendMode
	blendDstMode      BlendMode
	blendEquation     BlendEquation
	culling           bool
	cullFace          CullFace
	colorWrite        bool
	colorTargetFormat gpu.TextureFormat
	depthFormat       gpu.TextureFormat
}

type pipelineCache struct {
	cache map[pipelineKey]gpu.RenderPipeline
}

func newPipelineCache() *pipelineCache {
	return &pipelineCache{
		cache: make(map[pipelineKey]gpu.RenderPipeline),
	}
}

func (pc *pipelineCache) getOrCreate(state *State, program *Program, colorFormat, depthFormat gpu.TextureFormat) gpu.RenderPipeline {
	key := pipelineKey{
		programName:       state.ProgramName,
		topology:          state.Topology,
		depthTest:         state.DepthTest,
		depthWrite:        state.DepthWrite,
		depthFunc:         state.DepthFunc,
		blending:          state.Blending,
		blendSrcMode:      state.BlendSrcMode,
		blendDstMode:      state.BlendDstMode,
		blendEquation:     state.BlendEquation,
		culling:           state.Culling,
		cullFace:          state.CullFace,
		colorWrite:        state.ColorWrite,
		colorTargetFormat: colorFormat,
		depthFormat:       depthFormat,
	}

	if p, ok := pc.cache[key]; ok {
		return p
	}

	// Build the pipeline
	desc := gpu.RenderPipelineDescriptor{
		Layout:       program.pipelineLayout,
		VertexModule: program.vertexModule,
		VertexEntry:  "main",
		FragmentModule: program.fragmentModule,
		FragmentEntry:  "main",
		Primitive:    stateTopology(state.Topology),
		FrontFace:    gpu.FrontFaceCCW,
	}

	// Vertex buffer layouts — 4 slots matching the mesh
	desc.Buffers = standardVertexBufferLayouts()

	// Cull mode
	if state.Culling {
		switch state.CullFace {
		case CullBack:
			desc.CullMode = gpu.CullModeBack
		case CullFront:
			desc.CullMode = gpu.CullModeFront
		default:
			desc.CullMode = gpu.CullModeBack
		}
	} else {
		desc.CullMode = gpu.CullModeNone
	}

	// Color target (skip for depth-only passes)
	if colorFormat != gpu.TextureFormatUndefined {
		writeMask := gpu.ColorWriteMaskAll
		if !state.ColorWrite {
			writeMask = gpu.ColorWriteMaskNone
		}

		target := gpu.ColorTargetState{
			Format:    colorFormat,
			WriteMask: writeMask,
		}

		if state.Blending {
			srcFactor := mapBlendFactor(state.BlendSrcMode)
			dstFactor := mapBlendFactor(state.BlendDstMode)
			blendOp := gpu.BlendOperationAdd
			if state.BlendEquation == BlendFuncMax {
				blendOp = gpu.BlendOperationMax
			}
			target.Blend = &gpu.BlendState{
				Color: gpu.BlendComponent{SrcFactor: srcFactor, DstFactor: dstFactor, Operation: blendOp},
				Alpha: gpu.BlendComponent{SrcFactor: srcFactor, DstFactor: dstFactor, Operation: blendOp},
			}
		}

		desc.Targets = []gpu.ColorTargetState{target}
	}

	// Depth stencil
	if depthFormat != gpu.TextureFormatUndefined {
		depthCompare := gpu.CompareFunctionAlways
		if state.DepthTest {
			switch state.DepthFunc {
			case DepthLess:
				depthCompare = gpu.CompareFunctionLess
			case DepthLessEqual:
				depthCompare = gpu.CompareFunctionLessEqual
			case DepthEqual:
				depthCompare = gpu.CompareFunctionEqual
			}
		}
		desc.DepthStencil = &gpu.DepthStencilState{
			Format:            depthFormat,
			DepthWriteEnabled: state.DepthWrite,
			DepthCompare:      depthCompare,
		}
	}

	pipeline := renderer.device.CreateRenderPipeline(desc)
	pc.cache[key] = pipeline
	glog.Infof("Created pipeline for state: %s (program: %s)", state.Name, state.ProgramName)
	return pipeline
}

func mapBlendFactor(mode BlendMode) gpu.BlendFactor {
	switch mode {
	case BlendOne:
		return gpu.BlendFactorOne
	case BlendSrcAlpha:
		return gpu.BlendFactorSrcAlpha
	case BlendOneMinusSrcAlpha:
		return gpu.BlendFactorOneMinusSrcAlpha
	default:
		return gpu.BlendFactorOne
	}
}

// standardVertexBufferLayouts returns the 4 vertex buffer slot layouts used by all mesh rendering.
func standardVertexBufferLayouts() []gpu.VertexBufferLayout {
	return []gpu.VertexBufferLayout{
		// Slot 0: positions (float32x3)
		{
			ArrayStride: 12,
			StepMode:    gpu.VertexStepModeVertex,
			Attributes: []gpu.VertexAttribute{
				{Format: gpu.VertexFormatFloat32x3, Offset: 0, ShaderLocation: 0},
			},
		},
		// Slot 1: normals (float32x3)
		{
			ArrayStride: 12,
			StepMode:    gpu.VertexStepModeVertex,
			Attributes: []gpu.VertexAttribute{
				{Format: gpu.VertexFormatFloat32x3, Offset: 0, ShaderLocation: 1},
			},
		},
		// Slot 2: texcoords (float32x3)
		{
			ArrayStride: 12,
			StepMode:    gpu.VertexStepModeVertex,
			Attributes: []gpu.VertexAttribute{
				{Format: gpu.VertexFormatFloat32x3, Offset: 0, ShaderLocation: 2},
			},
		},
		// Slot 3: instance data (InstanceDataLen bytes per instance)
		{
			ArrayStride: uint64(InstanceDataLen),
			StepMode:    gpu.VertexStepModeInstance,
			Attributes: []gpu.VertexAttribute{
				// Model matrix (4 x vec4f)
				{Format: gpu.VertexFormatFloat32x4, Offset: 0, ShaderLocation: 3},
				{Format: gpu.VertexFormatFloat32x4, Offset: 16, ShaderLocation: 4},
				{Format: gpu.VertexFormatFloat32x4, Offset: 32, ShaderLocation: 5},
				{Format: gpu.VertexFormatFloat32x4, Offset: 48, ShaderLocation: 6},
				// MVP matrix (4 x vec4f)
				{Format: gpu.VertexFormatFloat32x4, Offset: 64, ShaderLocation: 7},
				{Format: gpu.VertexFormatFloat32x4, Offset: 80, ShaderLocation: 8},
				{Format: gpu.VertexFormatFloat32x4, Offset: 96, ShaderLocation: 9},
				{Format: gpu.VertexFormatFloat32x4, Offset: 112, ShaderLocation: 10},
				// Custom data (4 x vec4f)
				{Format: gpu.VertexFormatFloat32x4, Offset: 128, ShaderLocation: 11},
				{Format: gpu.VertexFormatFloat32x4, Offset: 144, ShaderLocation: 12},
				{Format: gpu.VertexFormatFloat32x4, Offset: 160, ShaderLocation: 13},
				{Format: gpu.VertexFormatFloat32x4, Offset: 176, ShaderLocation: 14},
			},
		},
	}
}

func stateTopology(t string) gpu.PrimitiveTopology {
	switch t {
	case "lines":
		return gpu.PrimitiveTopologyLineList
	case "points":
		return gpu.PrimitiveTopologyPointList
	default:
		return gpu.PrimitiveTopologyTriangleList
	}
}

func init() {
	_ = fmt.Sprintf // keep fmt imported for potential debugging
}
