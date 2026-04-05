package core

import (
	"encoding/json"

	"github.com/fcvarela/gosg/gpu"
	"github.com/golang/glog"
)

// Program holds a GPU shader program with shader modules and bind group layouts.
type Program struct {
	name             string
	vertexModule     gpu.ShaderModule
	fragmentModule   gpu.ShaderModule
	bindGroupLayouts []gpu.BindGroupLayout
	pipelineLayout   gpu.PipelineLayout
	spec             programSpec
}

// programSpec is parsed from .wgpu.json files.
type programSpec struct {
	Shaders          map[string]string       `json:"shaders"`
	BindGroupLayouts []bindGroupLayoutSpec   `json:"bindGroupLayouts"`
	TextureBindings  map[string]textureBindingSpec `json:"textureBindings"`
}

type bindGroupLayoutSpec struct {
	Entries []bindGroupEntrySpec `json:"entries"`
}

type bindGroupEntrySpec struct {
	Binding    uint32   `json:"binding"`
	Visibility []string `json:"visibility"`
	Buffer     *struct {
		Type string `json:"type"`
	} `json:"buffer,omitempty"`
	Texture *struct {
		SampleType    string `json:"sampleType"`
		ViewDimension string `json:"viewDimension"`
	} `json:"texture,omitempty"`
	Sampler *struct {
		Type string `json:"type"`
	} `json:"sampler,omitempty"`
}

type textureBindingSpec struct {
	Group          uint32 `json:"group"`
	TextureBinding uint32 `json:"textureBinding"`
	SamplerBinding uint32 `json:"samplerBinding"`
}

// Name returns the program's name.
func (p *Program) Name() string {
	return p.name
}

func parseVisibility(vis []string) gpu.ShaderStage {
	var stage gpu.ShaderStage
	for _, v := range vis {
		switch v {
		case "vertex":
			stage |= gpu.ShaderStageVertex
		case "fragment":
			stage |= gpu.ShaderStageFragment
		}
	}
	return stage
}

func loadProgram(name string, data []byte) *Program {
	var spec programSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		glog.Fatalf("Error reading program spec %s: %v", name, err)
	}

	p := &Program{
		name: name,
		spec: spec,
	}

	// Load and compile shader modules
	if vsFile, ok := spec.Shaders["vertex"]; ok {
		vsSource := resourceManager.ProgramData(vsFile)
		p.vertexModule = renderer.device.CreateShaderModuleWGSL(string(vsSource))
	}
	if fsFile, ok := spec.Shaders["fragment"]; ok {
		fsSource := resourceManager.ProgramData(fsFile)
		p.fragmentModule = renderer.device.CreateShaderModuleWGSL(string(fsSource))
	}

	// Create bind group layouts
	p.bindGroupLayouts = make([]gpu.BindGroupLayout, len(spec.BindGroupLayouts))
	for i, bglSpec := range spec.BindGroupLayouts {
		entries := make([]gpu.BindGroupLayoutEntry, len(bglSpec.Entries))
		for j, e := range bglSpec.Entries {
			entries[j].Binding = e.Binding
			entries[j].Visibility = parseVisibility(e.Visibility)

			if e.Buffer != nil {
				entries[j].Buffer = &gpu.BufferBindingLayout{
					Type: gpu.BufferBindingTypeUniform,
				}
			}
			if e.Texture != nil {
				sampleType := gpu.TextureSampleTypeFloat
				if e.Texture.SampleType == "depth" {
					sampleType = gpu.TextureSampleTypeDepth
				} else if e.Texture.SampleType == "unfilterable-float" {
					sampleType = gpu.TextureSampleTypeUnfilterableFloat
				}
				viewDim := gpu.TextureViewDimension2D
				if e.Texture.ViewDimension == "2d-array" {
					viewDim = gpu.TextureViewDimension2DArray
				}
				entries[j].Texture = &gpu.TextureBindingLayout{
					SampleType:    sampleType,
					ViewDimension: viewDim,
				}
			}
			if e.Sampler != nil {
				samplerType := gpu.SamplerBindingTypeFiltering
				if e.Sampler.Type == "non-filtering" {
					samplerType = gpu.SamplerBindingTypeNonFiltering
				} else if e.Sampler.Type == "comparison" {
					samplerType = gpu.SamplerBindingTypeComparison
				}
				entries[j].Sampler = &gpu.SamplerBindingLayout{
					Type: samplerType,
				}
			}
		}
		p.bindGroupLayouts[i] = renderer.device.CreateBindGroupLayout(entries)
	}

	// Create pipeline layout
	p.pipelineLayout = renderer.device.CreatePipelineLayout(p.bindGroupLayouts)

	glog.Infof("Loaded program: %s", name)
	return p
}
