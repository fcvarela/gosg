package demoapp

import (
	"github.com/fcvarela/gosg/core"
)

func getDemoSceneShadowTextures(s *core.Scene) []*core.Texture {
	light := getDemoSceneLight(s)
	return []*core.Texture{light.Shadower.ShadowTexture()}
}

func getDemoSceneLight(s *core.Scene) *core.Light {
	geoRoot := s.Root().Children()[0]
	lightNode := geoRoot.Children()[len(geoRoot.Children())-2]
	return lightNode.Light()
}

func makeDemoScene() *core.Scene {
	return core.GetResourceManager().Scene("demo.yaml")
}
