package core

import (
	"log"

	"github.com/golang/glog"
)

// ResourceSystem is an interface which wraps all resource management logic.
type ResourceSystem interface {
	Model(string) []byte
	Texture(string) []byte
	Program(string) []byte
	Pipeline(string) []byte
	ProgramData(string) []byte
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
func (r *ResourceManager) SetSystem(s ResourceSystem) {
	if r.system != nil {
		log.Fatal("Can't replace previously registered resource system.")
	}
	r.system = s
}

// Model returns a scenegraph node.
func (r *ResourceManager) Model(name string) *Node {
	if r.models[name] == nil {
		resource := r.system.Model(name)
		r.models[name] = LoadModel(name, resource)
	}
	return r.models[name].Copy()
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
func (r *ResourceManager) Pipeline(name string) *Pipeline {
	if r.pipelines[name] == nil {
		resource := r.system.Pipeline(name)
		pipeline, err := ParsePipeline(resource)
		if err != nil {
			glog.Fatalf("Cannot parse pipeline %s: %v", name, err)
		}
		pipeline.Name = name
		r.pipelines[name] = pipeline
	}
	return r.pipelines[name]
}

// ProgramData returns source file contents for a given program or subprogram
func (r *ResourceManager) ProgramData(name string) []byte {
	return r.system.ProgramData(name)
}
