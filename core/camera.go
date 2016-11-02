package core

import (
	"math"
	"runtime"
	"unsafe"

	"github.com/fcvarela/gosg/protos"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/golang/glog"
)

// ProjectionType is used to express projection types (ie: perspective, orthographic).
type ProjectionType uint8

// ClearMode is used to express the type of framebuffer clearing which will be executed
// before the rendering logic is called.
type ClearMode uint8

const (
	// ClearColor is used to signal whether a framebuffer's color buffer should be cleared.
	ClearColor ClearMode = 1 << 0

	// ClearDepth is used to signal whether a framebuffer's color buffer should be cleared.
	ClearDepth ClearMode = 1 << 1
)

const (
	// PerspectiveProjection represents a perspective projection.
	PerspectiveProjection ProjectionType = iota

	// OrthographicProjection represents an orthographic projection.
	OrthographicProjection
)

// RenderFn is a function which injects render commands into a buffer
type RenderFn func(*Camera, map[*protos.State][]*Node, chan RenderCommand)

// A Camera represents a scenegraph camera object. It wraps data which holds
// its transforms (projection and view matrices), clear information, whether
// it should perform auto reshape on window resizes, a pointer to its node on
// the scenegraph, as well as clipping distances (near, far planes) and render
// ordering, target and techniques.
type Camera struct {
	name               string
	autoReshape        bool
	autoFrustum        bool
	projectionType     ProjectionType
	clearColor         mgl32.Vec4
	clearDepth         float64
	clearMode          ClearMode
	node               *Node
	scene              *Node
	viewMatrix         mgl64.Mat4
	projectionMatrix   mgl64.Mat4
	viewport           mgl32.Vec4
	vertFOV            float64
	clipDistance       mgl64.Vec2
	dirty              bool
	renderOrder        uint8
	framebuffer        Framebuffer
	frustum            [6]mgl64.Vec4
	cascadingAABBS     [maxCascades]*AABB
	cascadingZCuts     [maxCascades]float64
	constants          *CameraConstants
	renderTechnique    RenderFn
	stateBuckets       map[*protos.State][]*Node
	visibleOpaqueNodes []*Node
}

// CamerasByRenderOrder is used to sort cameras by the render order field.
type CamerasByRenderOrder []*Camera

// Len implements the sort.Interface interface.
func (a CamerasByRenderOrder) Len() int {
	return len(a)
}

// Swap implements the sort.Interface interface.
func (a CamerasByRenderOrder) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// Less implements the sort.Interface interface.
func (a CamerasByRenderOrder) Less(i, j int) bool {
	return a[i].renderOrder < a[j].renderOrder
}

func deleteCamera(c *Camera) {
	glog.Info("Camera finalizer start: ", c.name)

	c.node = nil
	c.scene = nil

	glog.Info("Camera finalizer finish: ", c.name)
}

// NewCamera creates and returns a Camera with the given name and projection type.
func NewCamera(name string, projType ProjectionType) *Camera {
	cam := Camera{}
	cam.name = name
	cam.clearColor = mgl32.Vec4{0.0, 0.0, 0.0, 0.0}
	cam.clearDepth = 1.0
	cam.clearMode = ClearColor | ClearDepth
	cam.SetProjectionType(projType)
	cam.node = NewNode(name)
	cam.node.bounds = nil
	cam.constants = NewCameraConstants()
	cam.renderTechnique = DefaultRenderTechnique
	cam.stateBuckets = make(map[*protos.State][]*Node)
	cam.visibleOpaqueNodes = make([]*Node, 0)

	runtime.SetFinalizer(&cam, deleteCamera)
	return &cam
}

// Node returns the camera's scenegraph Node
func (c *Camera) Node() *Node {
	return c.node
}

// SetAutoReshape sets whether the camera should reshape its viewport and transforms
// when the window is resized.
func (c *Camera) SetAutoReshape(autoReshape bool) {
	c.autoReshape = autoReshape
}

// SetAutoFrustum sets whether the camera should auto set its near/far clip planes
func (c *Camera) SetAutoFrustum(af bool) {
	c.autoFrustum = af
}

// Scene returns the camera's scene root. This is the node that it will start culling
// traversals on. To render an entire scene you would set this to the scene's root node. To
// render subtrees you would set this to the subtree root node. This allows you to split
// your scene into subgroups and have cameras render them separately.
func (c *Camera) Scene() *Node {
	return c.scene
}

// SetScene sets the camera's scene root. See Scene() for more information on what this is.
func (c *Camera) SetScene(s *Node) {
	c.scene = s
}

// Name returns the camera's name.
func (c *Camera) Name() string {
	return c.name
}

// Viewport returns the camera's viewport.
func (c *Camera) Viewport() mgl32.Vec4 {
	return c.viewport
}

// ProjectionMatrix returns the camera's projection matrix.
func (c *Camera) ProjectionMatrix() mgl64.Mat4 {
	return c.projectionMatrix
}

// ViewMatrix returns the camera's view matrix.
func (c *Camera) ViewMatrix() mgl64.Mat4 {
	return c.viewMatrix
}

// Frustum returns the camera's frustum, which is a set of 6 planes.
func (c *Camera) Frustum() [6]mgl64.Vec4 {
	return c.frustum
}

// ClearColor returns the camera's clear color.
func (c *Camera) ClearColor() mgl32.Vec4 {
	return c.clearColor
}

// ClearDepth returns the camera's clear depth.
func (c *Camera) ClearDepth() float64 {
	return c.clearDepth
}

// ClearMode returns the camera's clear mode.
func (c *Camera) ClearMode() ClearMode {
	return c.clearMode
}

// AddNodeToRenderBuckets adds a node to the camera's renderbuckets for the next render
// loop
func (c *Camera) AddNodeToRenderBuckets(n *Node) {
	c.stateBuckets[n.state] = append(c.stateBuckets[n.state], n)
	if !n.state.Blending {
		c.visibleOpaqueNodes = append(c.visibleOpaqueNodes, n)
	}
}

// Reshape reshapes the camera's viewport and transforms according to a given window size.
func (c *Camera) Reshape(windowSize mgl32.Vec2) {
	if c.autoReshape {
		if windowSize[0] != c.viewport[2] || windowSize[1] != c.viewport[3] {
			c.SetViewport(mgl32.Vec4{0.0, 0.0, windowSize[0], windowSize[1]})
		}
	}

	c.viewMatrix = c.node.InverseWorldTransform()

	if c.dirty == true {
		if c.projectionType == PerspectiveProjection {
			c.projectionMatrix = mgl64.Perspective(mgl64.DegToRad(c.vertFOV), float64(c.viewport[2]/c.viewport[3]), c.clipDistance[0], c.clipDistance[1])
		}
		if c.projectionType == OrthographicProjection {
			c.projectionMatrix = mgl64.Ortho(float64(c.viewport[0]), float64(c.viewport[2]), float64(c.viewport[3]), float64(c.viewport[1]), c.clipDistance[0], c.clipDistance[1])
		}
		c.dirty = false
	}

	var worldVisibleRange = c.clipDistance[1] - c.clipDistance[0]
	var minRange, maxRange = c.clipDistance[0], float64(0)
	for cascade := 0; cascade < numCascades; cascade++ {
		maxRange = worldVisibleRange / (math.Pow(2, 2*float64(numCascades-cascade-1)))
		var projectionMatrix = mgl64.Perspective(mgl64.DegToRad(c.vertFOV), float64(c.viewport[2]/c.viewport[3]), minRange, maxRange)
		var frustumCorners = [8]mgl64.Vec3{
			{-1, -1, -1},
			{+1, -1, -1},
			{-1, +1, -1},
			{+1, +1, -1},
			{-1, -1, +1},
			{+1, -1, +1},
			{-1, +1, +1},
			{+1, +1, +1},
		}

		c.cascadingAABBS[cascade] = NewAABB()
		var invProjectionMatrix = projectionMatrix.Mul4(c.viewMatrix).Inv()

		for p := range frustumCorners {
			c.cascadingAABBS[cascade].ExtendWithPoint(mgl64.TransformCoordinate(frustumCorners[p], invProjectionMatrix))
		}
		c.cascadingZCuts[cascade] = maxRange
		minRange = maxRange
	}

	c.frustum = MakeFrustum(c.projectionMatrix, c.viewMatrix)
}

// MakeFrustum creates a frustum's 6 planes from a projection and view matrix
func MakeFrustum(p, v mgl64.Mat4) (f [6]mgl64.Vec4) {
	viewProj := p.Mul4(v)
	rowX := viewProj.Row(0)
	rowY := viewProj.Row(1)
	rowZ := viewProj.Row(2)
	rowW := viewProj.Row(3)

	f[0] = rowW.Add(rowX)
	f[1] = rowW.Sub(rowX)
	f[2] = rowW.Add(rowY)
	f[3] = rowW.Sub(rowY)
	f[4] = rowW.Add(rowZ)
	f[5] = rowW.Sub(rowZ)

	// normalize planes
	for i := range f {
		len := f[i].Vec3().Len()
		f[i] = f[i].Mul(1.0 / len)
	}

	return f
}

// SetProjectionMatrix sets the camera's projection matrix.
func (c *Camera) SetProjectionMatrix(m mgl64.Mat4) {
	c.projectionMatrix = m
}

// SetViewMatrix sets the camera's view matrix.
func (c *Camera) SetViewMatrix(m mgl64.Mat4) {
	c.viewMatrix = m
}

// SetProjectionType sets the camera's projection type.
func (c *Camera) SetProjectionType(projType ProjectionType) {
	c.dirty = true
	c.projectionType = projType
}

// SetViewport sets the camera's viewport.
func (c *Camera) SetViewport(vp mgl32.Vec4) {
	c.dirty = true
	c.viewport = vp
}

// SetVerticalFieldOfView sets the camera's vertical field of view. This is ignored
// for orthographic projections.
func (c *Camera) SetVerticalFieldOfView(vfov float64) {
	c.dirty = true
	c.vertFOV = vfov
}

// SetClipDistance sets the camera's near and far clipping planes.
func (c *Camera) SetClipDistance(cd mgl64.Vec2) {
	c.dirty = true
	c.clipDistance = cd
}

// SetClearColor sets the camera's clear color.
func (c *Camera) SetClearColor(cc mgl32.Vec4) {
	c.clearColor = cc
}

// SetClearDepth sets the camera's clear depth.
func (c *Camera) SetClearDepth(d float64) {
	c.clearDepth = d
}

// SetClearMode sets the camera's clear mode.
func (c *Camera) SetClearMode(cm ClearMode) {
	c.clearMode = cm
}

// Framebuffer returns the camera's render target.
func (c *Camera) Framebuffer() Framebuffer {
	return c.framebuffer
}

// SetFramebuffer sets the camera's render target.
func (c *Camera) SetFramebuffer(rt Framebuffer) {
	c.framebuffer = rt
}

// Constants returns the camera's constants. This contains constant buffers for vertex shaders, etc.
func (c *Camera) Constants() *CameraConstants {
	return c.constants
}

// RenderOrder order returns the camera's render order.
func (c *Camera) RenderOrder() uint8 {
	return c.renderOrder
}

// SetRenderOrder sets the camera's render order.
func (c *Camera) SetRenderOrder(o uint8) {
	c.renderOrder = o
}

// SetRenderTechnique sets the camera's render technique.
func (c *Camera) SetRenderTechnique(r RenderFn) {
	c.renderTechnique = r
}

type innerConstants struct {
	ViewMatrix           mgl32.Mat4
	ProjectionMatrix     mgl32.Mat4
	ViewProjectionMatrix mgl32.Mat4
	LightCount           mgl32.Vec4
	LightBlocks          [16]LightBlock
}

// CameraConstants holds a uniform buffer passed to all programs which contains global camera transforms
// and a list of active lights.
type CameraConstants struct {
	inner  innerConstants
	buffer UniformBuffer
}

var (
	lightBlockLen = (maxCascades*16 + maxCascades*4 + 4 + 4)
	// 16 lights * 16 floats per light + 4 floats lightcount, all mult4 (sizeof float)
	sceneBlockLen = (3*16 + 16*lightBlockLen + 4) * 4
)

// NewCameraConstants returns a new CameraConstants. The UniformBuffer is returned by the rendersystem.
func NewCameraConstants() *CameraConstants {
	return &CameraConstants{innerConstants{}, renderSystem.NewUniformBuffer()}
}

// SetData sets matrices and light information for the entire scene
func (sb *CameraConstants) SetData(pMatrix, vMatrix mgl64.Mat4, l []*Light) {
	// matrices
	sb.inner.ViewMatrix = Mat4DoubleToFloat(vMatrix)
	sb.inner.ProjectionMatrix = Mat4DoubleToFloat(pMatrix)
	sb.inner.ViewProjectionMatrix = Mat4DoubleToFloat(pMatrix.Mul4(vMatrix))

	// lights
	sb.inner.LightCount = mgl32.Vec4{0.0, 0.0, 0.0, 0.0}
	for i := 0; i < len(l) && i < len(sb.inner.LightBlocks); i++ {
		sb.inner.LightCount = sb.inner.LightCount.Add(mgl32.Vec4{1.0, 1.0, 1.0, 1.0})
		sb.inner.LightBlocks[i] = l[i].Block
	}

	// copy to buffer
	inner := sb.inner

	// set the constant buffer pointer and length
	sb.buffer.Set(unsafe.Pointer(&inner), sceneBlockLen)
}

// UniformBuffer returns the camera constants uniform buffer
func (sb *CameraConstants) UniformBuffer() UniformBuffer {
	return sb.buffer
}
