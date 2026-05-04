package core

import (
	"encoding/json"
	"fmt"

	"github.com/fcvarela/gosg/gpu"
	"github.com/golang/glog"
)

func unmarshalEnumJSON(b []byte, m map[string]int, defaultVal int) (int, error) {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		if v, ok := m[s]; ok {
			return v, nil
		}
		return defaultVal, fmt.Errorf("unknown enum value: %s", s)
	}
	var i int
	if err := json.Unmarshal(b, &i); err != nil {
		return defaultVal, fmt.Errorf("failed to unmarshal enum: %w", err)
	}
	return i, nil
}

// CullFace specifies which face to cull.
type CullFace int

const (
	CullBack  CullFace = 0
	CullFront CullFace = 1
	CullBoth  CullFace = 2
)

var cullFaceMap = map[string]CullFace{
	"CULL_BACK": CullBack, "CULL_FRONT": CullFront, "CULL_BOTH": CullBoth,
}

func (c *CullFace) UnmarshalJSON(b []byte) error {
	v, err := unmarshalEnumJSON(b, map[string]int{
		"CULL_BACK": int(CullBack), "CULL_FRONT": int(CullFront), "CULL_BOTH": int(CullBoth),
	}, int(CullBack))
	if err != nil {
		return err
	}
	*c = CullFace(v)
	return nil
}

// BlendMode specifies a blend factor.
type BlendMode int

const (
	BlendSrcAlpha         BlendMode = 0
	BlendOneMinusSrcAlpha BlendMode = 1
	BlendOne              BlendMode = 2
)

var blendModeMap = map[string]BlendMode{
	"BLEND_SRC_ALPHA": BlendSrcAlpha, "BLEND_ONE_MINUS_SRC_ALPHA": BlendOneMinusSrcAlpha, "BLEND_ONE": BlendOne,
}

func (b *BlendMode) UnmarshalJSON(data []byte) error {
	v, err := unmarshalEnumJSON(data, map[string]int{
		"BLEND_SRC_ALPHA": int(BlendSrcAlpha), "BLEND_ONE_MINUS_SRC_ALPHA": int(BlendOneMinusSrcAlpha), "BLEND_ONE": int(BlendOne),
	}, int(BlendSrcAlpha))
	if err != nil {
		return err
	}
	*b = BlendMode(v)
	return nil
}

// BlendEquation specifies a blend equation.
type BlendEquation int

const (
	BlendFuncAdd BlendEquation = 0
	BlendFuncMax BlendEquation = 1
)

var blendEqMap = map[string]BlendEquation{
	"BLEND_FUNC_ADD": BlendFuncAdd, "BLEND_FUNC_MAX": BlendFuncMax,
}

func (e *BlendEquation) UnmarshalJSON(data []byte) error {
	v, err := unmarshalEnumJSON(data, map[string]int{
		"BLEND_FUNC_ADD": int(BlendFuncAdd), "BLEND_FUNC_MAX": int(BlendFuncMax),
	}, int(BlendFuncAdd))
	if err != nil {
		return err
	}
	*e = BlendEquation(v)
	return nil
}

// DepthFunc specifies a depth comparison function.
type DepthFunc int

const (
	DepthLessEqual DepthFunc = 0
	DepthLess      DepthFunc = 1
	DepthEqual     DepthFunc = 2
)

var depthFuncMap = map[string]DepthFunc{
	"DEPTH_LESS_EQUAL": DepthLessEqual, "DEPTH_LESS": DepthLess, "DEPTH_EQUAL": DepthEqual,
}

func (d *DepthFunc) UnmarshalJSON(data []byte) error {
	v, err := unmarshalEnumJSON(data, map[string]int{
		"DEPTH_LESS_EQUAL": int(DepthLessEqual), "DEPTH_LESS": int(DepthLess), "DEPTH_EQUAL": int(DepthEqual),
	}, int(DepthLessEqual))
	if err != nil {
		return err
	}
	*d = DepthFunc(v)
	return nil
}

// Pipeline describes the GPU pipeline configuration for rendering.
type Pipeline struct {
	Name          string        `json:"name,omitempty"`
	ProgramName   string        `json:"programName,omitempty"`
	Topology      string        `json:"topology,omitempty"` // "triangles" (default), "lines", "points"
	Culling       bool          `json:"culling,omitempty"`
	CullFace      CullFace      `json:"cullFace,omitempty"`
	Blending      bool          `json:"blending,omitempty"`
	BlendSrcMode  BlendMode     `json:"blendSrcMode,omitempty"`
	BlendDstMode  BlendMode     `json:"blendDstMode,omitempty"`
	BlendEquation BlendEquation `json:"blendEquation,omitempty"`
	DepthTest     bool          `json:"depthTest,omitempty"`
	DepthWrite    bool          `json:"depthWrite,omitempty"`
	DepthFunc     DepthFunc     `json:"depthFunc,omitempty"`
	ColorWrite    bool          `json:"colorWrite,omitempty"`
	ScissorTest   bool          `json:"scissorTest,omitempty"`
}

// ParsePipeline parses a Pipeline from JSON bytes.
func ParsePipeline(data []byte) (*Pipeline, error) {
	var s Pipeline
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// pipelineKey uniquely identifies a render pipeline configuration.
type pipelineKey struct {
	programName        string
	topology           string
	depthTest          bool
	depthWrite         bool
	depthFunc          DepthFunc
	blending           bool
	blendSrcMode       BlendMode
	blendDstMode       BlendMode
	blendEquation      BlendEquation
	culling            bool
	cullFace           CullFace
	colorWrite         bool
	colorTargetFormats [4]gpu.TextureFormat // up to 4 MRT targets
	numColorTargets    int
	depthFormat        gpu.TextureFormat
}

type pipelineCache struct {
	cache map[pipelineKey]gpu.RenderPipeline
}

func newPipelineCache() *pipelineCache {
	return &pipelineCache{
		cache: make(map[pipelineKey]gpu.RenderPipeline),
	}
}

func (pc *pipelineCache) release() {
	for _, p := range pc.cache {
		p.Release()
	}
	pc.cache = make(map[pipelineKey]gpu.RenderPipeline)
}

func (pc *pipelineCache) getOrCreate(p *Pipeline, program *Program, colorFormats []gpu.TextureFormat, depthFormat gpu.TextureFormat) gpu.RenderPipeline {
	var fmtArr [4]gpu.TextureFormat
	copy(fmtArr[:], colorFormats)

	key := pipelineKey{
		programName:        p.ProgramName,
		topology:           p.Topology,
		depthTest:          p.DepthTest,
		depthWrite:         p.DepthWrite,
		depthFunc:          p.DepthFunc,
		blending:           p.Blending,
		blendSrcMode:       p.BlendSrcMode,
		blendDstMode:       p.BlendDstMode,
		blendEquation:      p.BlendEquation,
		culling:            p.Culling,
		cullFace:           p.CullFace,
		colorWrite:         p.ColorWrite,
		colorTargetFormats: fmtArr,
		numColorTargets:    len(colorFormats),
		depthFormat:        depthFormat,
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
		Primitive:    stateTopology(p.Topology),
		FrontFace:    gpu.FrontFaceCCW,
	}

	// Vertex buffer layouts — 4 slots matching the mesh
	desc.Buffers = standardVertexBufferLayouts()

	// Cull mode
	if p.Culling {
		switch p.CullFace {
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

	// Color targets (multiple for MRT)
	for _, colorFormat := range colorFormats {
		if colorFormat == gpu.TextureFormatUndefined {
			continue
		}
		writeMask := gpu.ColorWriteMaskAll
		if !p.ColorWrite {
			writeMask = gpu.ColorWriteMaskNone
		}

		target := gpu.ColorTargetState{
			Format:    colorFormat,
			WriteMask: writeMask,
		}

		if p.Blending {
			srcFactor := mapBlendFactor(p.BlendSrcMode)
			dstFactor := mapBlendFactor(p.BlendDstMode)
			blendOp := gpu.BlendOperationAdd
			if p.BlendEquation == BlendFuncMax {
				blendOp = gpu.BlendOperationMax
			}
			target.Blend = &gpu.BlendState{
				Color: gpu.BlendComponent{SrcFactor: srcFactor, DstFactor: dstFactor, Operation: blendOp},
				Alpha: gpu.BlendComponent{SrcFactor: srcFactor, DstFactor: dstFactor, Operation: blendOp},
			}
		}

		desc.Targets = append(desc.Targets, target)
	}

	// Depth stencil
	if depthFormat != gpu.TextureFormatUndefined {
		depthCompare := gpu.CompareFunctionAlways
		if p.DepthTest {
			switch p.DepthFunc {
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
			DepthWriteEnabled: p.DepthWrite,
			DepthCompare:      depthCompare,
		}
	}

	pipeline := renderer.device.CreateRenderPipeline(desc)
	pc.cache[key] = pipeline
	glog.Infof("Created pipeline: %s (program: %s)", p.Name, p.ProgramName)
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
