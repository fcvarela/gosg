package core

import (
	"errors"
	"fmt"
	"strings"
)

// ResourceSystem is an interface which wraps all resource management logic.
type ResourceSystem interface {
	Model(string) []byte
	ModelPath(string) string
	Texture(string) []byte
	Program(string) []byte
	Pipeline(string) []byte
	ProgramData(string) []byte
	Scene(string) []byte
}

// ResourceManager wraps a resourcesystem and contains configuration about the location of each resource type.
type ResourceManager struct {
	system    ResourceSystem
	programs  map[string]*Program
	pipelines map[string]*Pipeline
	models    map[string]*Node
	textures  map[string]*Texture
}

var (
	resourceManager *ResourceManager
)

func init() {
	resourceManager = &ResourceManager{
		programs:  make(map[string]*Program),
		pipelines: make(map[string]*Pipeline),
		models:    make(map[string]*Node),
		textures:  make(map[string]*Texture),
	}
}

// GetResourceManager returns the resource manager.
func GetResourceManager() *ResourceManager {
	return resourceManager
}

// SetSystem sets the resource manager's resource system
func (r *ResourceManager) SetSystem(s ResourceSystem) error {
	if r.system != nil {
		return errors.New("can't replace previously registered resource system")
	}
	r.system = s
	return nil
}

// Model returns a scenegraph node.
func (r *ResourceManager) Model(name string) (*Node, error) {
	if r.models[name] == nil {
		var node *Node
		var err error
		if strings.HasSuffix(name, ".gltf") || strings.HasSuffix(name, ".glb") {
			node, err = LoadGLTF(name, r.system)
		} else {
			resource := r.system.Model(name)
			node, err = LoadModel(name, resource)
		}
		if err != nil {
			return nil, err
		}
		r.models[name] = node
	}
	return r.models[name].Copy(), nil
}

// Program returns a GPU program.
func (r *ResourceManager) Program(name string) *Program {
	if r.programs[name] == nil {
		resource := r.system.Program(name)
		r.programs[name] = renderer.NewProgram(name, resource)
	}
	return r.programs[name]
}

// Pipeline returns a Pipeline parsed from JSON.
func (r *ResourceManager) Pipeline(name string) (*Pipeline, error) {
	if r.pipelines[name] == nil {
		resource := r.system.Pipeline(name)
		pipeline, err := ParsePipeline(resource)
		if err != nil {
			return nil, fmt.Errorf("cannot parse pipeline %s: %w", name, err)
		}
		pipeline.Name = name
		r.pipelines[name] = pipeline
	}
	return r.pipelines[name], nil
}

// ProgramData returns source file contents for a given program or subprogram
func (r *ResourceManager) ProgramData(name string) []byte {
	return r.system.ProgramData(name)
}

// Scene loads and returns a Scene from a YAML file.
func (r *ResourceManager) Scene(name string) *Scene {
	data := r.system.Scene(name)
	return LoadSceneFromYAML(data)
}
