package core

import (
	"math"
	"unsafe"

	"github.com/fcvarela/gosg/gpu"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
)

// Shadower is an interface which wraps logic to implement shadowing of a light
type Shadower interface {
	ShadowTexture() *Texture
	NumCascades() int
	Render(*Light, *Camera)
}

// ShadowMap implements depth-only shadow mapping with PCF and configurable cascades.
type ShadowMap struct {
	size         uint32
	numCascades  int
	lambda       float64
	cameras      []*Camera
	depthTexture gpu.Texture   // single 2D array texture with N layers
	arrayView    gpu.TextureView // view as 2d_array for shader sampling
	layerViews   []gpu.TextureView // per-layer views for framebuffer attachments
	texture      *Texture     // wrapper for bind group creation
	cascadeAABBs [maxCascades]*AABB
	cascadeZCuts [maxCascades]float64
}

const maxCascades = 10

// NewShadowMap returns a new ShadowMap with the given number of cascades.
func NewShadowMap(size uint32, cascades int) *ShadowMap {
	sm := &ShadowMap{
		size:        size,
		numCascades: cascades,
		lambda:      0.5,
		cameras:     make([]*Camera, cascades),
		layerViews:  make([]gpu.TextureView, cascades),
	}

	// Create a single 2D array depth texture with N layers
	sm.depthTexture = renderer.device.CreateTexture(gpu.TextureDescriptor{
		Size:      gpu.Extent3D{Width: size, Height: size, DepthOrArrayLayers: uint32(cascades)},
		Format:    gpu.TextureFormatDepth32Float,
		Usage:     gpu.TextureUsageTextureBinding | gpu.TextureUsageRenderAttachment,
		Dimension: gpu.TextureDimension2D,
		MipLevels: 1,
	})

	// Create array view for shader sampling
	sm.arrayView = sm.depthTexture.CreateViewArray(uint32(cascades))

	// Create per-layer views for framebuffer attachments + comparison sampler
	compSampler := renderer.device.CreateSampler(gpu.SamplerDescriptor{
		AddressModeU: gpu.AddressModeClampToEdge,
		AddressModeV: gpu.AddressModeClampToEdge,
		AddressModeW: gpu.AddressModeClampToEdge,
		MinFilter:    gpu.FilterModeNearest,
		MagFilter:    gpu.FilterModeNearest,
		Compare:      gpu.CompareFunctionLessEqual,
	})

	sm.texture = &Texture{
		id:      allocateTextureID(),
		texture: sm.depthTexture,
		view:    sm.arrayView,
		sampler: compSampler,
		descriptor: TextureDescriptor{
			Width: size, Height: size,
			Format: TextureFormatDEPTH, SizedFormat: TextureSizedFormatDEPTH32F,
		},
	}

	for i := 0; i < cascades; i++ {
		sm.layerViews[i] = sm.depthTexture.CreateViewLayer(uint32(i))

		fb := &Framebuffer{colorAttachments: make(map[int]*Texture)}
		fb.depthAttachment = &Texture{
			id:   allocateTextureID(),
			view: sm.layerViews[i],
			descriptor: TextureDescriptor{
				Width: size, Height: size,
				Format: TextureFormatDEPTH, SizedFormat: TextureSizedFormatDEPTH32F,
			},
		}

		c := NewCamera("ShadowCamera", OrthographicProjection)
		c.SetFramebuffer(fb)
		c.SetViewport(mgl32.Vec4{0, 0, float32(size), float32(size)})
		c.SetAutoReshape(false)
		c.SetRenderTechnique(nil)
		c.SetClearMode(ClearDepth)
		c.SetClearDepth(1.0)
		sm.cameras[i] = c
	}

	return sm
}

// ShadowTexture returns the shadow map array texture for binding.
func (s *ShadowMap) ShadowTexture() *Texture {
	return s.texture
}

// NumCascades returns the number of cascades.
func (s *ShadowMap) NumCascades() int {
	return s.numCascades
}

// computeCascades computes PSSM cascade split frustums with tight fitting.
func (s *ShadowMap) computeCascades(camera *Camera) {
	near := camera.ClipDistance()[0]
	far := camera.ClipDistance()[1]

	splits := [maxCascades + 1]float64{}
	splits[0] = near
	for i := 1; i <= s.numCascades; i++ {
		t := float64(i) / float64(s.numCascades)
		logSplit := near * math.Pow(far/near, t)
		uniSplit := near + (far-near)*t
		splits[i] = s.lambda*logSplit + (1-s.lambda)*uniSplit
	}

	var sceneBounds *AABB
	if len(camera.VisibleOpaqueNodes()) > 0 {
		sceneBounds = NewAABB()
		for _, n := range camera.VisibleOpaqueNodes() {
			if n.worldBounds != nil {
				sceneBounds.ExtendWithBox(n.worldBounds)
			}
		}
	}

	aspect := camera.Aspect()
	fovRad := mgl64.DegToRad(camera.VerticalFOV())
	viewMatrix := camera.ViewMatrix()

	for cascade := 0; cascade < s.numCascades; cascade++ {
		proj := PerspectiveWebGPU(fovRad, aspect, splits[cascade], splits[cascade+1])
		invVP := proj.Mul4(viewMatrix).Inv()

		corners := [8]mgl64.Vec3{
			{-1, -1, 0}, {+1, -1, 0}, {-1, +1, 0}, {+1, +1, 0},
			{-1, -1, 1}, {+1, -1, 1}, {-1, +1, 1}, {+1, +1, 1},
		}

		cascadeAABB := NewAABB()
		for _, c := range corners {
			cascadeAABB.ExtendWithPoint(mgl64.TransformCoordinate(c, invVP))
		}

		if sceneBounds != nil {
			cascadeAABB = cascadeAABB.Intersection(sceneBounds)
		}

		s.cascadeAABBs[cascade] = cascadeAABB
		s.cascadeZCuts[cascade] = splits[cascade+1]
	}
}

func (s *ShadowMap) renderCascade(cascade int, light *Light, camera *Camera) {
	shadowCam := s.cameras[cascade]

	lightPos64 := mgl64.Vec3{float64(light.Block.Position.X()), float64(light.Block.Position.Y()), float64(light.Block.Position.Z())}
	shadowCam.viewMatrix = mgl64.LookAtV(lightPos64, mgl64.Vec3{0, 0, 0}, mgl64.Vec3{0, 1, 0})

	nodesBoundsLight := s.cascadeAABBs[cascade].Transformed(shadowCam.viewMatrix)

	worldUnitsPerTexel := nodesBoundsLight.Max().Sub(nodesBoundsLight.Min()).Mul(1.0 / float64(s.size))
	projMinX := math.Floor(nodesBoundsLight.Min().X()/worldUnitsPerTexel.X()) * worldUnitsPerTexel.X()
	projMaxX := math.Floor(nodesBoundsLight.Max().X()/worldUnitsPerTexel.X()) * worldUnitsPerTexel.X()
	projMinY := math.Floor(nodesBoundsLight.Min().Y()/worldUnitsPerTexel.Y()) * worldUnitsPerTexel.Y()
	projMaxY := math.Floor(nodesBoundsLight.Max().Y()/worldUnitsPerTexel.Y()) * worldUnitsPerTexel.Y()

	shadowCam.projectionMatrix = OrthoWebGPU(
		projMinX, projMaxX, projMinY, projMaxY,
		-nodesBoundsLight.Max().Z(), -nodesBoundsLight.Min().Z())

	vpmatrix := shadowCam.projectionMatrix.Mul4(shadowCam.viewMatrix)
	biasvpmatrix := mgl64.Mat4FromCols(
		mgl64.Vec4{0.5, 0.0, 0.0, 0.0},
		mgl64.Vec4{0.0, -0.5, 0.0, 0.0},
		mgl64.Vec4{0.0, 0.0, 1.0, 0.0},
		mgl64.Vec4{0.5, 0.5, 0.0, 1.0}).Mul4(vpmatrix)

	light.Block.ZCuts[cascade] = mgl32.Vec4{float32(s.cascadeZCuts[cascade]), 0, 0, 0}
	light.Block.VPMatrix[cascade] = Mat4DoubleToFloat(biasvpmatrix)

	shadowCam.constants.SetData(shadowCam.projectionMatrix, shadowCam.viewMatrix, nil)

	desc := shadowCam.MakeRenderPassDescriptor(false, true)
	pass := renderer.BeginRenderPass(desc)

	for pipeline, nodeBucket := range camera.pipelineBuckets {
		if pipeline.Blending {
			continue
		}
		// Bind the shadow array texture for all nodes
		for _, n := range nodeBucket {
			n.material.SetTexture("shadowTex", s.texture)
		}
		pass.SetPipeline(resourceManager.Pipeline("shadow"))
		pass.SetCameraConstants(shadowCam.constants.buffer)
		RenderBatchedNodes(pass, shadowCam, nodeBucket)
	}
	pass.End()
}

// Render implements the Shadower interface
func (s *ShadowMap) Render(light *Light, cam *Camera) {
	s.computeCascades(cam)
	for c := 0; c < s.numCascades; c++ {
		s.renderCascade(c, light, cam)
		renderer.Flush()
	}
}

// InstanceData and rendering helpers use this
var _ = unsafe.Pointer(nil)
