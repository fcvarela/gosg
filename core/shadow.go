package core

import (
	"math"

	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
)

// Shadower is an interface which wraps logic to implement shadowing of a light
type Shadower interface {
	// Textures returns the shadow textures used by this shadower
	Textures() []Texture

	// Render calls the shadower render implementation by assing a light and a scene camera.
	Render(*Light, *Camera, chan RenderCommand)
}

// ShadowMap is a utility implementation of the Shadower interface which renders shadows by using a cascading shadow map.
type ShadowMap struct {
	size     uint32
	cameras  []*Camera
	textures []Texture
}

const (
	numCascades = 3
	maxCascades = 10
)

// NewShadowMap returns a new ShadowMap
func NewShadowMap(size uint32) *ShadowMap {
	shadowMap := &ShadowMap{size, make([]*Camera, numCascades), make([]Texture, numCascades)}
	for i := 0; i < numCascades; i++ {
		// create a framebuffer for the cascade
		framebuffer := renderSystem.NewFramebuffer()

		// create a texture to write to
		texture := renderSystem.NewTexture(TextureDescriptor{
			Width:         size,
			Height:        size,
			Mipmaps:       false,
			Target:        TextureTarget2D,
			Format:        TextureFormatRG,
			SizedFormat:   TextureSizedFormatRG32F,
			ComponentType: TextureComponentTypeFLOAT,
			Filter:        TextureFilterLinear,
			WrapMode:      TextureWrapModeRepeat,
		}, nil)

		// set it as the framebuffer color attachment
		framebuffer.SetColorAttachment(0, texture)

		// create a camera and set its framebuffer
		c := NewCamera("ShadowCamera", OrthographicProjection)
		c.SetFramebuffer(framebuffer)
		c.SetViewport(mgl32.Vec4{0.0, 0.0, float32(size), float32(size)})
		c.SetAutoReshape(false)
		c.SetRenderTechnique(nil)
		shadowMap.cameras[i] = c
		shadowMap.textures[i] = framebuffer.ColorAttachment(0)
	}
	return shadowMap
}

// Textures implements the Shadower interface
func (s *ShadowMap) Textures() []Texture {
	return s.textures
}

func (s *ShadowMap) renderCascade(cascade int, light *Light, camera *Camera, cmdBuf chan RenderCommand) {
	/*
		1-find all objects that are inside the current camera frustum
		2-find minimal aa bounding box that encloses them all
		3-transform corners of that bounding box to the light's space (using light's view matrix)
		4-find aa bounding box in light's space of the transformed (now obb) bounding box
		5-this aa bounding box is your directional light's orthographic frustum.
	*/

	// compute lightcamera viewmatrix
	lightPos64 := mgl64.Vec3{float64(light.Block.Position.X()), float64(light.Block.Position.Y()), float64(light.Block.Position.Z())}
	s.cameras[cascade].viewMatrix = mgl64.LookAtV(lightPos64, mgl64.Vec3{0.0, 0.0, 0.0}, mgl64.Vec3{0.0, 1.0, 0.0})

	// 3-transform corners of that bounding box to the light's space (using light's view matrix)
	// 4-find aa bounding box in light's space of the transformed (now obb) bounding box
	nodesBoundsLight := camera.cascadingAABBS[cascade].Transformed(s.cameras[cascade].viewMatrix)

	// 5-this aa bounding box is your directional light's orthographic frustum. except we want integer increments
	worldUnitsPerTexel := nodesBoundsLight.Max().Sub(nodesBoundsLight.Min()).Mul(1.0 / float64(s.size))
	projMinX := math.Floor(nodesBoundsLight.Min().X()/worldUnitsPerTexel.X()) * worldUnitsPerTexel.X()
	projMaxX := math.Floor(nodesBoundsLight.Max().X()/worldUnitsPerTexel.X()) * worldUnitsPerTexel.X()
	projMinY := math.Floor(nodesBoundsLight.Min().Y()/worldUnitsPerTexel.Y()) * worldUnitsPerTexel.Y()
	projMaxY := math.Floor(nodesBoundsLight.Max().Y()/worldUnitsPerTexel.Y()) * worldUnitsPerTexel.Y()

	s.cameras[cascade].projectionMatrix = mgl64.Ortho(
		projMinX, projMaxX,
		projMinY, projMaxY,
		-nodesBoundsLight.Max().Z(),
		-nodesBoundsLight.Min().Z())

	vpmatrix := s.cameras[cascade].projectionMatrix.Mul4(s.cameras[cascade].viewMatrix)
	biasvpmatrix := mgl64.Mat4FromCols(
		mgl64.Vec4{0.5, 0.0, 0.0, 0.0},
		mgl64.Vec4{0.0, 0.5, 0.0, 0.0},
		mgl64.Vec4{0.0, 0.0, 0.5, 0.0},
		mgl64.Vec4{0.5, 0.5, 0.5, 1.0}).Mul4(vpmatrix)

	// set light block
	light.Block.ZCuts[cascade] = mgl32.Vec4{float32(camera.cascadingZCuts[cascade]), 0.0, 0.0, 0.0}
	light.Block.VPMatrix[cascade] = Mat4DoubleToFloat(biasvpmatrix)

	// set camera constants
	s.cameras[cascade].constants.SetData(s.cameras[cascade].projectionMatrix, s.cameras[cascade].viewMatrix, nil)

	// create a single stage now
	var shadowCam = s.cameras[cascade]
	cmdBuf <- &SetFramebufferCommand{shadowCam.framebuffer}
	cmdBuf <- &SetViewportCommand{shadowCam.viewport}
	cmdBuf <- &ClearCommand{shadowCam.clearMode, shadowCam.clearColor, shadowCam.clearDepth}

	// create pass per bucket, opaque is default
	for state, nodeBucket := range camera.stateBuckets {
		if state.Blending == true {
			continue
		}

		for _, n := range nodeBucket {
			n.materialData.SetTexture(fmt.Sprintf("shadowTex%d", cascade), s.textures[cascade])
		}

		cmdBuf <- &BindStateCommand{resourceManager.State("shadow")}
		cmdBuf <- &BindUniformBufferCommand{"cameraConstants", shadowCam.constants.buffer}
		RenderBatchedNodes(nodeBucket, cmdBuf)
	}
}

// Render implements the Shadower interface
func (s *ShadowMap) Render(light *Light, cam *Camera, cmdBuf chan RenderCommand) {
	for c := 0; c < numCascades; c++ {
		s.renderCascade(c, light, cam, cmdBuf)
	}

	return
}
