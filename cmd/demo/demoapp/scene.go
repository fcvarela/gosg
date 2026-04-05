package demoapp

import (
	"github.com/fcvarela/gosg/core"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
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

func makeGeometrySubscene() (*core.Node, *core.Camera) {
	geometryCamera := core.NewCamera("GeometryPassCamera", core.PerspectiveProjection)
	geometryCamera.SetAutoReshape(true)
	geometryCamera.SetAutoFrustum(true)
	geometryCamera.SetVerticalFieldOfView(60.0)
	geometryCamera.SetClearColor(mgl32.Vec4{0.4, 0.6, 0.9, 1.0})
	geometryCamera.SetClearMode(core.ClearColor | core.ClearDepth)
	geometryCamera.SetClipDistance(mgl64.Vec2{1.0, 250.0})
	geometryCamera.Node().SetInputComponent(core.NewMouseCameraInputComponent())
	geometryCamera.SetRenderOrder(0)

	fb := core.GetRenderer().NewFramebuffer()
	geometryCamera.SetFramebuffer(fb)

	colorTexture := core.GetRenderer().NewTexture(core.TextureDescriptor{
		Width:   uint32(core.GetWindowManager().WindowSize().X()),
		Height:  uint32(core.GetWindowManager().WindowSize().Y()),
		Mipmaps: false, Target: core.TextureTarget2D,
		Format: core.TextureFormatRGBA, SizedFormat: core.TextureSizedFormatRGBA16F,
		ComponentType: core.TextureComponentTypeFLOAT,
		Filter:        core.TextureFilterLinear, WrapMode: core.TextureWrapModeClampEdge,
	}, nil)

	depthTexture := core.GetRenderer().NewTexture(core.TextureDescriptor{
		Width:   uint32(core.GetWindowManager().WindowSize().X()),
		Height:  uint32(core.GetWindowManager().WindowSize().Y()),
		Mipmaps: false, Target: core.TextureTarget2D,
		Format: core.TextureFormatDEPTH, SizedFormat: core.TextureSizedFormatDEPTH32F,
		ComponentType: core.TextureComponentTypeFLOAT,
		Filter:        core.TextureFilterNearest, WrapMode: core.TextureWrapModeClampEdge,
	}, nil)

	fb.SetColorAttachment(0, colorTexture)
	fb.SetDepthAttachment(depthTexture)

	geometryNode := core.NewNode("GeometryRoot")

	for i := -12; i < 12; i++ {
		for j := -12; j < 12; j++ {
			randomVec := mgl64.Vec3{float64(i) * 9.96 * 2.0, float64(j) * 9.96 * 2.0, 0.0}
			f16 := core.GetResourceManager().Model("f16.model")
			f16.Translate(randomVec)
			f16.Rotate(float64(i), randomVec)
			geometryNode.AddChild(f16)
		}
	}

	shadowMap1 := core.NewShadowMap(2048, 3)
	lightNode1 := core.NewNode("Light1")
	lightNode1.Translate(mgl64.Vec3{+1000.0, 0.0, +1000.0})
	light1 := &core.Light{
		Block: core.LightBlock{
			Position: mgl32.Vec4{0.0, 0.0, 0.0, 1.0},
			Color:    mgl32.Vec4{1.0, 1.0, 1.0, 1.0},
		},
		Shadower:   shadowMap1,
		ShadowBias: 0.01,
	}
	lightNode1.SetLight(light1)

	geometryNode.AddChild(lightNode1)
	geometryCamera.Node().Translate(mgl64.Vec3{0.0, 0.0, 50.0})
	geometryCamera.SetScene(geometryNode)

	return geometryNode, geometryCamera
}

func makeFXAASubscene(sourceFB *core.Framebuffer) (*core.Node, *core.Camera) {
	windowSize := core.GetWindowManager().WindowSize()

	fxaaCamera := core.NewCamera("FXAACamera", core.OrthographicProjection)
	fxaaCamera.SetAutoReshape(false)
	fxaaCamera.SetViewport(mgl32.Vec4{0.0, 0.0, windowSize.X(), windowSize.Y()})
	fxaaCamera.SetClipDistance(mgl64.Vec2{0.0, 1.0})
	fxaaCamera.SetClearColor(mgl32.Vec4{0.1, 0.0, 0.0, 1.0})
	fxaaCamera.SetClearMode(core.ClearColor | core.ClearDepth)
	fxaaCamera.SetRenderOrder(1)
	fxaaCamera.SetRenderTechnique(core.PostProcessRenderTechnique)
	fxaaCamera.Reshape(core.GetWindowManager().WindowSize())

	screenQuadNode := core.NewNode("ScreenQuad")
	screenQuadNode.SetCullComponent(new(core.AlwaysPassCuller))
	screenQuadNode.SetMesh(core.NewScreenQuadMesh(windowSize.X(), windowSize.Y()))
	screenQuadNode.SetPipeline(core.GetResourceManager().Pipeline("fxaa"))
	screenQuadNode.Material().SetTexture("colorTexture", sourceFB.ColorAttachment(0))

	fxaaCamera.SetScene(screenQuadNode)
	return screenQuadNode, fxaaCamera
}

func makeDemoScene() *core.Scene {
	s := core.NewScene("Demo1")
	s.SetRoot(core.NewNode("ROOT"))

	geoRoot, geoCamera := makeGeometrySubscene()
	s.AddCamera(geoRoot, geoCamera)

	hdrRoot, hdrCamera := makeFXAASubscene(geoCamera.Framebuffer())
	s.AddCamera(hdrRoot, hdrCamera)

	s.SetActive(true)

	s.Root().AddChild(geoRoot)
	s.Root().AddChild(hdrRoot)

	return s
}
