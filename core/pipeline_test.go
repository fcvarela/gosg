package core

import (
	"testing"
)

func TestParsePipeline(t *testing.T) {
	data := []byte(`{
		"name": "test-pipeline",
		"programName": "test-program",
		"culling": true,
		"cullFace": "CULL_BACK",
		"blending": false,
		"depthTest": true,
		"depthWrite": true,
		"depthFunc": "DEPTH_LESS",
		"colorWrite": true
	}`)

	p, err := ParsePipeline(data)
	if err != nil {
		t.Fatalf("ParsePipeline failed: %v", err)
	}

	if p.Name != "test-pipeline" {
		t.Errorf("name = %q, want %q", p.Name, "test-pipeline")
	}
	if p.ProgramName != "test-program" {
		t.Errorf("programName = %q, want %q", p.ProgramName, "test-program")
	}
	if !p.Culling {
		t.Error("culling should be true")
	}
	if p.CullFace != CullBack {
		t.Errorf("cullFace = %v, want %v", p.CullFace, CullBack)
	}
	if p.Blending {
		t.Error("blending should be false")
	}
	if !p.DepthTest {
		t.Error("depthTest should be true")
	}
	if !p.DepthWrite {
		t.Error("depthWrite should be true")
	}
	if p.DepthFunc != DepthLess {
		t.Errorf("depthFunc = %v, want %v", p.DepthFunc, DepthLess)
	}
}

func TestParsePipelineInvalidJSON(t *testing.T) {
	_, err := ParsePipeline([]byte(`{invalid}`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestCullFaceUnmarshalJSON(t *testing.T) {
	tests := []struct {
		input string
		want  CullFace
	}{
		{`"CULL_BACK"`, CullBack},
		{`"CULL_FRONT"`, CullFront},
		{`"CULL_BOTH"`, CullBoth},
		{`0`, CullBack},
		{`1`, CullFront},
		{`2`, CullBoth},
	}

	for _, tt := range tests {
		var cf CullFace
		err := cf.UnmarshalJSON([]byte(tt.input))
		if err != nil {
			t.Errorf("UnmarshalJSON(%s) error: %v", tt.input, err)
			continue
		}
		if cf != tt.want {
			t.Errorf("UnmarshalJSON(%s) = %v, want %v", tt.input, cf, tt.want)
		}
	}
}

func TestCullFaceUnmarshalJSONInvalid(t *testing.T) {
	var cf CullFace
	err := cf.UnmarshalJSON([]byte(`"INVALID"`))
	if err == nil {
		t.Error("expected error for invalid cull face")
	}
}

func TestBlendModeUnmarshalJSON(t *testing.T) {
	tests := []struct {
		input string
		want  BlendMode
	}{
		{`"BLEND_SRC_ALPHA"`, BlendSrcAlpha},
		{`"BLEND_ONE_MINUS_SRC_ALPHA"`, BlendOneMinusSrcAlpha},
		{`"BLEND_ONE"`, BlendOne},
	}

	for _, tt := range tests {
		var bm BlendMode
		err := bm.UnmarshalJSON([]byte(tt.input))
		if err != nil {
			t.Errorf("UnmarshalJSON(%s) error: %v", tt.input, err)
			continue
		}
		if bm != tt.want {
			t.Errorf("UnmarshalJSON(%s) = %v, want %v", tt.input, bm, tt.want)
		}
	}
}

func TestDepthFuncUnmarshalJSON(t *testing.T) {
	tests := []struct {
		input string
		want  DepthFunc
	}{
		{`"DEPTH_LESS_EQUAL"`, DepthLessEqual},
		{`"DEPTH_LESS"`, DepthLess},
		{`"DEPTH_EQUAL"`, DepthEqual},
	}

	for _, tt := range tests {
		var df DepthFunc
		err := df.UnmarshalJSON([]byte(tt.input))
		if err != nil {
			t.Errorf("UnmarshalJSON(%s) error: %v", tt.input, err)
			continue
		}
		if df != tt.want {
			t.Errorf("UnmarshalJSON(%s) = %v, want %v", tt.input, df, tt.want)
		}
	}
}

func TestBlendEquationUnmarshalJSON(t *testing.T) {
	tests := []struct {
		input string
		want  BlendEquation
	}{
		{`"BLEND_FUNC_ADD"`, BlendFuncAdd},
		{`"BLEND_FUNC_MAX"`, BlendFuncMax},
	}

	for _, tt := range tests {
		var be BlendEquation
		err := be.UnmarshalJSON([]byte(tt.input))
		if err != nil {
			t.Errorf("UnmarshalJSON(%s) error: %v", tt.input, err)
			continue
		}
		if be != tt.want {
			t.Errorf("UnmarshalJSON(%s) = %v, want %v", tt.input, be, tt.want)
		}
	}
}
