package metal

import "github.com/fcvarela/gosg/core"

// RenderSystem implements the core.RenderSystem interface
type RenderSystem struct {
}

func init() {
	core.SetRenderSystem(New())
}

// New returns a new instance of a metal rendersystem
func New() core.RenderSystem {
	return &RenderSystem{}
}
