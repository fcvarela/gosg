package core

import (
	"math"

	"github.com/fcvarela/gosg/gpu"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/golang/glog"
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
	depthTexture gpu.Texture     // single 2D array texture with N layers
	arrayView    gpu.TextureView // view as 2d_array for shader sampling
	layerViews   []gpu.TextureView // per-layer views for framebuffer attachments
	texture      *Texture       // wrapper for bind group creation
	cascadeCenters [maxCascades]mgl64.Vec3
	cascadeRadii   [maxCascades]float64
	cascadeZCuts   [maxCascades]float64
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

// computeCascades computes PSSM cascade bounding spheres for stable shadow mapping.
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

	aspect := camera.Aspect()
	fovRad := mgl64.DegToRad(camera.VerticalFOV())
	viewMatrix := camera.ViewMatrix()

	for cascade := 0; cascade < s.numCascades; cascade++ {
		proj := PerspectiveWebGPU(fovRad, aspect, splits[cascade], splits[cascade+1])
		invVP := proj.Mul4(viewMatrix).Inv()

		// Unproject 8 NDC corners to world space
		ndcCorners := [8]mgl64.Vec3{
			{-1, -1, 0}, {+1, -1, 0}, {-1, +1, 0}, {+1, +1, 0},
			{-1, -1, 1}, {+1, -1, 1}, {-1, +1, 1}, {+1, +1, 1},
		}

		var worldCorners [8]mgl64.Vec3
		for i, c := range ndcCorners {
			worldCorners[i] = mgl64.TransformCoordinate(c, invVP)
		}

		// Bounding sphere: center = average of corners
		center := mgl64.Vec3{}
		for _, wc := range worldCorners {
			center = center.Add(wc)
		}
		center = center.Mul(1.0 / 8.0)

		// Radius = max distance from center to any corner
		radius := 0.0
		for _, wc := range worldCorners {
			d := wc.Sub(center).Len()
			if d > radius {
				radius = d
			}
		}

		// Round radius up to texel boundary for size stability
		texelsPerUnit := float64(s.size) / (2.0 * radius)
		radius = math.Ceil(radius*texelsPerUnit) / texelsPerUnit

		s.cascadeCenters[cascade] = center
		s.cascadeRadii[cascade] = radius
		s.cascadeZCuts[cascade] = splits[cascade+1]
	}
}

func (s *ShadowMap) renderCascade(cascade int, light *Light, camera *Camera) {
	shadowCam := s.cameras[cascade]
	center := s.cascadeCenters[cascade]
	radius := s.cascadeRadii[cascade]

	// Light view matrix centered on the cascade sphere
	lightPos64 := mgl64.Vec3{float64(light.Block.Position.X()), float64(light.Block.Position.Y()), float64(light.Block.Position.Z())}
	lightDir := lightPos64.Normalize()
	shadowCam.viewMatrix = mgl64.LookAtV(
		center.Add(lightDir.Mul(radius)),
		center,
		mgl64.Vec3{0, 1, 0},
	)

	// Compute tight ortho bounds by intersecting cascade frustum with scene AABB in light space.
	// This ensures the shadow map covers only the geometry, not the full camera frustum.
	sceneBounds := camera.Scene().WorldBounds()
	lightSceneBounds := sceneBounds.Transformed(shadowCam.viewMatrix)

	// Frustum-derived bounds in light space
	centerLight := mgl64.TransformCoordinate(center, shadowCam.viewMatrix)
	frustLeft := centerLight[0] - radius
	frustRight := centerLight[0] + radius
	frustBottom := centerLight[1] - radius
	frustTop := centerLight[1] + radius

	// Intersect XY with scene bounds in light space
	left := math.Max(frustLeft, lightSceneBounds.min[0])
	right := math.Min(frustRight, lightSceneBounds.max[0])
	bottom := math.Max(frustBottom, lightSceneBounds.min[1])
	top := math.Min(frustTop, lightSceneBounds.max[1])

	// If intersection is degenerate, fall back to frustum bounds
	if left >= right || bottom >= top {
		left = frustLeft
		right = frustRight
		bottom = frustBottom
		top = frustTop
	}

	// For Z, use the scene's full depth extent in light space.
	// Near=0 is at the shadow camera; we need to cover from the camera to the farthest scene geometry.
	// lightSceneBounds.min[2] is the most negative Z (farthest from camera in view space).
	zFar := math.Max(2.0*radius, -lightSceneBounds.min[2]+1.0)

	// Snap to texel grid for stability (use the tighter XY extent for texel size)
	extentX := right - left
	extentY := top - bottom
	extent := math.Max(extentX, extentY)
	worldUnitsPerTexel := extent / float64(s.size)
	if worldUnitsPerTexel > 0 {
		left = math.Floor(left/worldUnitsPerTexel) * worldUnitsPerTexel
		right = math.Ceil(right/worldUnitsPerTexel) * worldUnitsPerTexel
		bottom = math.Floor(bottom/worldUnitsPerTexel) * worldUnitsPerTexel
		top = math.Ceil(top/worldUnitsPerTexel) * worldUnitsPerTexel
	}

	shadowCam.projectionMatrix = OrthoWebGPU(left, right, bottom, top, 0.0, zFar)

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
	if pass == nil {
		return
	}

	shadowPipeline, err := resourceManager.Pipeline("shadow")
	if err != nil {
		glog.Warningf("failed to load shadow pipeline: %v", err)
		pass.End()
		return
	}

	for pipeline, nodeBucket := range camera.pipelineBuckets {
		if pipeline.Blending {
			continue
		}
		// Bind the shadow array texture for all nodes
		for _, n := range nodeBucket {
			n.material.SetTexture("shadowTex", s.texture)
		}
		pass.SetPipeline(shadowPipeline)
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
