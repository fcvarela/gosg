package demoapp

import (
	"github.com/fcvarela/gosg/core"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
)

func getDemoSceneShadowTextures(s *core.Scene) []core.Texture {
	geoRoot := s.Root().Children()[0]
	lightNode := geoRoot.Children()[len(geoRoot.Children())-2]
	return lightNode.Light().Shadower.Textures()
}

func makeGeometrySubscene() (*core.Node, *core.Camera) {
	// geometry camera
	geometryCamera := core.NewCamera("GeometryPassCamera", core.PerspectiveProjection)
	geometryCamera.SetAutoReshape(true)
	geometryCamera.SetAutoFrustum(true)
	geometryCamera.SetVerticalFieldOfView(60.0)
	geometryCamera.SetClearColor(mgl32.Vec4{0.4, 0.6, 0.9, 1.0})
	geometryCamera.SetClearMode(core.ClearColor | core.ClearDepth)
	geometryCamera.SetClipDistance(mgl64.Vec2{1.0, 250.0})
	geometryCamera.Node().SetInputComponent(core.NewMouseCameraInputComponent())
	geometryCamera.SetRenderOrder(0)

	// create a framebuffer for this camera
	fb := core.GetRenderSystem().NewFramebuffer()
	geometryCamera.SetFramebuffer(fb)

	// attach its textures
	colorTexture := core.GetRenderSystem().NewTexture(core.TextureDescriptor{
		Width:         uint32(core.GetWindowManager().WindowSize().X()),
		Height:        uint32(core.GetWindowManager().WindowSize().Y()),
		Mipmaps:       false,
		Target:        core.TextureTarget2D,
		Format:        core.TextureFormatRGBA,
		SizedFormat:   core.TextureSizedFormatRGBA32F,
		ComponentType: core.TextureComponentTypeFLOAT,
		Filter:        core.TextureFilterLinear,
		WrapMode:      core.TextureWrapModeClampEdge,
	}, nil)

	depthTexture := core.GetRenderSystem().NewTexture(core.TextureDescriptor{
		Width:         uint32(core.GetWindowManager().WindowSize().X()),
		Height:        uint32(core.GetWindowManager().WindowSize().Y()),
		Mipmaps:       false,
		Target:        core.TextureTarget2D,
		Format:        core.TextureFormatDEPTH,
		SizedFormat:   core.TextureSizedFormatDEPTH32F,
		ComponentType: core.TextureComponentTypeFLOAT,
		Filter:        core.TextureFilterNearest,
		WrapMode:      core.TextureWrapModeClampEdge,
	}, nil)

	fb.SetColorAttachment(0, colorTexture)
	fb.SetDepthAttachment(depthTexture)

	geometryNode := core.NewNode("GeometryRoot")

	for i := -5; i < 5; i++ {
		for j := -5; j < 5; j++ {
			randomVec := mgl64.Vec3{float64(i) * 9.96 * 2.0, float64(j) * 9.96 * 2.0, 0.0}

			// load model
			f16 := core.GetResourceManager().Model("f16.model")
			f16.Translate(randomVec)
			f16.Rotate(float64(i), randomVec)

			// set aabb on subnodes
			//for _, c := range f16.Children() {
			//	c.State().AABB = true
			//}

			// add model to scenegraph
			geometryNode.AddChild(f16)
		}
	}

	// attach a light
	shadowMap1 := core.NewShadowMap(2048)
	lightNode1 := core.NewNode("Light1")
	lightNode1.Translate(mgl64.Vec3{+1000.0, 0.0, +1000.0})
	light1 := &core.Light{
		Block: core.LightBlock{
			Position: mgl32.Vec4{0.0, 0.0, 0.0, 1.0},
			Color:    mgl32.Vec4{1.0, 1.0, 1.0, 1.0},
		},
		Shadower: shadowMap1,
	}
	lightNode1.SetLight(light1)

	geometryNode.AddChild(lightNode1)
	geometryCamera.Node().Translate(mgl64.Vec3{0.0, 0.0, 50.0})
	geometryCamera.SetScene(geometryNode)

	return geometryNode, geometryCamera
}

func makeFXAASubscene(sourceFB core.Framebuffer) (*core.Node, *core.Camera) {
	windowSize := core.GetWindowManager().WindowSize()

	// geometry camera
	geometryCamera := core.NewCamera("FXAACamera", core.OrthographicProjection)
	geometryCamera.SetAutoReshape(false)
	geometryCamera.SetViewport(mgl32.Vec4{0.0, 0.0, windowSize.X(), windowSize.Y()})
	geometryCamera.SetClipDistance(mgl64.Vec2{0.0, 1.0})
	geometryCamera.SetClearColor(mgl32.Vec4{0.1, 0.0, 0.0, 1.0})
	geometryCamera.SetClearMode(core.ClearColor | core.ClearDepth)
	geometryCamera.SetRenderOrder(1)
	geometryCamera.Reshape(core.GetWindowManager().WindowSize())

	screenQuadNode := core.NewNode("ScreenQuad")
	screenQuadNode.SetCullComponent(new(core.AlwaysPassCuller))
	screenQuadNode.SetMesh(core.NewScreenQuadMesh(windowSize.X(), windowSize.Y()))
	screenQuadNode.SetState(core.GetResourceManager().State("fxaa"))
	screenQuadNode.MaterialData().SetTexture("colorTexture", sourceFB.ColorAttachment(0))

	geometryCamera.SetScene(screenQuadNode)
	return screenQuadNode, geometryCamera
}

func makeDemoScene() *core.Scene {
	// main scene
	s := core.NewScene("Demo1")
	s.SetRoot(core.NewNode("ROOT"))

	geoRoot, geoCamera := makeGeometrySubscene()

	// add geometry camera to geometry root node
	s.AddCamera(geoRoot, geoCamera)

	// add hdr + tonemapping camera
	hdrRoot, hdrCamera := makeFXAASubscene(geoCamera.Framebuffer())

	// add hdr camera to hdr root node
	s.AddCamera(hdrRoot, hdrCamera)

	// visibility and cursor mode
	s.SetActive(true)

	// add both scenes
	s.Root().AddChild(geoRoot)
	s.Root().AddChild(hdrRoot)

	return s
}
