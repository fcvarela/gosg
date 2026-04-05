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

	normalTexture := core.GetRenderer().NewTexture(core.TextureDescriptor{
		Width:   uint32(core.GetWindowManager().WindowSize().X()),
		Height:  uint32(core.GetWindowManager().WindowSize().Y()),
		Mipmaps: false, Target: core.TextureTarget2D,
		Format: core.TextureFormatRGBA, SizedFormat: core.TextureSizedFormatRGBA16F,
		ComponentType: core.TextureComponentTypeFLOAT,
		Filter:        core.TextureFilterNearest, WrapMode: core.TextureWrapModeClampEdge,
	}, nil)
	fb.SetColorAttachment(1, normalTexture)

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

func makeSSGISubscene(sourceFB *core.Framebuffer, sourceCamera *core.Camera) (*core.Node, *core.Camera, *core.Framebuffer) {
	windowSize := core.GetWindowManager().WindowSize()

	ssgiCamera := core.NewCamera("SSGICamera", core.OrthographicProjection)
	ssgiCamera.SetAutoReshape(false)
	ssgiCamera.SetViewport(mgl32.Vec4{0.0, 0.0, windowSize.X(), windowSize.Y()})
	ssgiCamera.SetClipDistance(mgl64.Vec2{0.0, 1.0})
	ssgiCamera.SetClearColor(mgl32.Vec4{0.0, 0.0, 0.0, 0.0})
	ssgiCamera.SetClearMode(core.ClearColor)
	ssgiCamera.SetRenderOrder(1)
	// Override constants with geometry camera's perspective projection for depth reconstruction
	geoCam := sourceCamera
	ssgiCamera.SetRenderTechnique(func(cam *core.Camera, buckets map[*core.Pipeline][]*core.Node) {
		cam.SetConstants(geoCam.ProjectionMatrix(), geoCam.ViewMatrix(), nil)
		core.PostProcessRenderTechnique(cam, buckets)
	})
	ssgiCamera.Reshape(core.GetWindowManager().WindowSize())

	// SSGI output framebuffer
	ssgiFB := core.GetRenderer().NewFramebuffer()
	ssgiColorTex := core.GetRenderer().NewTexture(core.TextureDescriptor{
		Width:   uint32(windowSize.X()),
		Height:  uint32(windowSize.Y()),
		Mipmaps: false, Target: core.TextureTarget2D,
		Format: core.TextureFormatRGBA, SizedFormat: core.TextureSizedFormatRGBA16F,
		ComponentType: core.TextureComponentTypeFLOAT,
		Filter:        core.TextureFilterLinear, WrapMode: core.TextureWrapModeClampEdge,
	}, nil)
	ssgiFB.SetColorAttachment(0, ssgiColorTex)
	ssgiCamera.SetFramebuffer(ssgiFB)

	ssgiNode := core.NewNode("SSGIQuad")
	ssgiNode.SetCullComponent(new(core.AlwaysPassCuller))
	ssgiNode.SetMesh(core.NewScreenQuadMesh(windowSize.X(), windowSize.Y()))
	ssgiNode.SetPipeline(core.GetResourceManager().Pipeline("ssgi"))

	// Bind geometry pass outputs
	ssgiNode.Material().SetTexture("colorTex", sourceFB.ColorAttachment(0))
	ssgiNode.Material().SetTexture("normalTex", sourceFB.ColorAttachment(1))
	// Depth needs non-comparison sampler for reading raw values
	depthReadTex := core.GetRenderer().NewDepthReadTexture(sourceFB.DepthAttachment())
	ssgiNode.Material().SetTexture("depthTex", depthReadTex)

	ssgiCamera.SetScene(ssgiNode)
	return ssgiNode, ssgiCamera, ssgiFB
}

func makeFXAASubscene(sourceFB *core.Framebuffer, ssgiFB *core.Framebuffer) (*core.Node, *core.Camera) {
	windowSize := core.GetWindowManager().WindowSize()

	geometryCamera := core.NewCamera("FXAACamera", core.OrthographicProjection)
	geometryCamera.SetAutoReshape(false)
	geometryCamera.SetViewport(mgl32.Vec4{0.0, 0.0, windowSize.X(), windowSize.Y()})
	geometryCamera.SetClipDistance(mgl64.Vec2{0.0, 1.0})
	geometryCamera.SetClearColor(mgl32.Vec4{0.1, 0.0, 0.0, 1.0})
	geometryCamera.SetClearMode(core.ClearColor | core.ClearDepth)
	geometryCamera.SetRenderOrder(2)
	geometryCamera.SetRenderTechnique(core.PostProcessRenderTechnique)
	geometryCamera.Reshape(core.GetWindowManager().WindowSize())

	screenQuadNode := core.NewNode("ScreenQuad")
	screenQuadNode.SetCullComponent(new(core.AlwaysPassCuller))
	screenQuadNode.SetMesh(core.NewScreenQuadMesh(windowSize.X(), windowSize.Y()))
	screenQuadNode.SetPipeline(core.GetResourceManager().Pipeline("fxaa"))
	screenQuadNode.Material().SetTexture("colorTexture", sourceFB.ColorAttachment(0))
	screenQuadNode.Material().SetTexture("ssgiTexture", ssgiFB.ColorAttachment(0))

	geometryCamera.SetScene(screenQuadNode)
	return screenQuadNode, geometryCamera
}

func makeDemoScene() *core.Scene {
	s := core.NewScene("Demo1")
	s.SetRoot(core.NewNode("ROOT"))

	geoRoot, geoCamera := makeGeometrySubscene()
	s.AddCamera(geoRoot, geoCamera)

	ssgiRoot, ssgiCamera, ssgiFB := makeSSGISubscene(geoCamera.Framebuffer(), geoCamera)
	s.AddCamera(ssgiRoot, ssgiCamera)

	hdrRoot, hdrCamera := makeFXAASubscene(geoCamera.Framebuffer(), ssgiFB)
	s.AddCamera(hdrRoot, hdrCamera)

	s.SetActive(true)

	s.Root().AddChild(geoRoot)
	s.Root().AddChild(ssgiRoot)
	s.Root().AddChild(hdrRoot)

	return s
}
