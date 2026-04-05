package core

import "encoding/json"

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
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		if v, ok := cullFaceMap[s]; ok {
			*c = v
			return nil
		}
	}
	var i int
	if err := json.Unmarshal(b, &i); err == nil {
		*c = CullFace(i)
		return nil
	}
	*c = CullBack
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
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		if v, ok := blendModeMap[s]; ok {
			*b = v
			return nil
		}
	}
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		*b = BlendMode(i)
		return nil
	}
	*b = BlendSrcAlpha
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
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		if v, ok := blendEqMap[s]; ok {
			*e = v
			return nil
		}
	}
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		*e = BlendEquation(i)
		return nil
	}
	*e = BlendFuncAdd
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
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		if v, ok := depthFuncMap[s]; ok {
			*d = v
			return nil
		}
	}
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		*d = DepthFunc(i)
		return nil
	}
	*d = DepthLessEqual
	return nil
}

// Pipeline describes the GPU pipeline state for rendering.
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
