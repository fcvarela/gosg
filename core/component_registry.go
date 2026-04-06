package core

var (
	inputComponentFactories  = make(map[string]func() InputComponent)
	cullComponentFactories   = make(map[string]func() Culler)
	renderTechniqueRegistry  = make(map[string]CameraRenderFn)
)

// RegisterInputComponent registers a named input component factory.
func RegisterInputComponent(name string, factory func() InputComponent) {
	inputComponentFactories[name] = factory
}

// RegisterCullComponent registers a named cull component factory.
func RegisterCullComponent(name string, factory func() Culler) {
	cullComponentFactories[name] = factory
}

// RegisterRenderTechnique registers a named render technique.
func RegisterRenderTechnique(name string, fn CameraRenderFn) {
	renderTechniqueRegistry[name] = fn
}

// LookupInputComponent returns a new instance of the named input component, or nil.
func LookupInputComponent(name string) InputComponent {
	if f, ok := inputComponentFactories[name]; ok {
		return f()
	}
	return nil
}

// LookupCullComponent returns a new instance of the named cull component, or nil.
func LookupCullComponent(name string) Culler {
	if f, ok := cullComponentFactories[name]; ok {
		return f()
	}
	return nil
}

// LookupRenderTechnique returns the named render technique, or nil.
func LookupRenderTechnique(name string) CameraRenderFn {
	if fn, ok := renderTechniqueRegistry[name]; ok {
		return fn
	}
	return nil
}

func init() {
	RegisterRenderTechnique("default", DefaultRenderTechnique)
	RegisterRenderTechnique("postprocess", PostProcessRenderTechnique)
	RegisterRenderTechnique("debug", DebugRenderTechnique)
	RegisterRenderTechnique("aabb", AABBRenderTechnique)

	RegisterCullComponent("default", func() Culler { return new(DefaultCuller) })
	RegisterCullComponent("alwaysPass", func() Culler { return new(AlwaysPassCuller) })

	RegisterInputComponent("mouseCameraInput", func() InputComponent { return NewMouseCameraInputComponent() })
}
