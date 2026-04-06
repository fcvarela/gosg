package core

import (
	"strings"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/golang/glog"
	"gopkg.in/yaml.v3"
)

// LoadSceneFromYAML parses YAML scene data and returns a Scene.
func LoadSceneFromYAML(data []byte) *Scene {
	var sf SceneFile
	if err := yaml.Unmarshal(data, &sf); err != nil {
		glog.Fatalf("Failed to parse scene YAML: %v", err)
	}

	scene := NewScene(sf.Name)
	root := NewNode("ROOT")
	scene.SetRoot(root)

	// Track cameras and framebuffers for deferred texture references
	cameraMap := make(map[string]*Camera)
	var deferred []deferredTexRef

	// Pass 1: Build node tree
	for i := range sf.Nodes {
		node, cams, refs := buildNode(&sf.Nodes[i])
		root.AddChild(node)
		for _, ce := range cams {
			cameraMap[ce.camera.Name()] = ce.camera
			scene.AddCamera(ce.parent, ce.camera)
		}
		deferred = append(deferred, refs...)
	}

	// Pass 2: Resolve deferred texture references ($CameraName.framebuffer.color0)
	for _, ref := range deferred {
		tex := resolveTextureRef(ref.ref, cameraMap)
		if tex != nil {
			ref.node.Material().SetTexture(ref.texName, tex)
		} else {
			glog.Warningf("Scene: unresolved texture reference %q for node %q", ref.ref, ref.node.Name())
		}
	}

	scene.SetActive(true)
	return scene
}

type deferredTexRef struct {
	node    *Node
	texName string
	ref     string
}

type cameraEntry struct {
	parent *Node
	camera *Camera
}

func buildNode(sn *SceneNode) (*Node, []cameraEntry, []deferredTexRef) {
	node := NewNode(sn.Name)
	var deferred []deferredTexRef
	var cameras []cameraEntry

	// Apply transform
	if sn.Position != [3]float64{} {
		node.Translate(mgl64.Vec3(sn.Position))
	}
	if sn.Rotation != [4]float64{} {
		node.Rotate(sn.Rotation[0], mgl64.Vec3{sn.Rotation[1], sn.Rotation[2], sn.Rotation[3]})
	}
	if sn.Scale != [3]float64{} {
		node.Scale(mgl64.Vec3(sn.Scale))
	}

	// Load model
	if sn.Model != "" {
		model := resourceManager.Model(sn.Model)
		for _, c := range model.Children() {
			node.AddChild(c)
		}
	}

	// Pipeline
	if sn.Pipeline != "" {
		node.SetPipeline(resourceManager.Pipeline(sn.Pipeline))
	}

	// Cull component
	if sn.Cull != "" {
		if culler := LookupCullComponent(sn.Cull); culler != nil {
			node.SetCullComponent(culler)
		}
	}

	// Screen quad
	if sn.ScreenQuad {
		windowSize := GetWindowManager().WindowSize()
		node.SetMesh(NewScreenQuadMesh(windowSize.X(), windowSize.Y()))
	}

	// Light
	if sn.Light != nil {
		light := &Light{
			Block: LightBlock{
				Position: mgl32.Vec4{0, 0, 0, 1},
				Color:    mgl32.Vec4{sn.Light.Color[0], sn.Light.Color[1], sn.Light.Color[2], 1},
			},
			ShadowBias: sn.Light.ShadowBias,
		}
		if sn.Light.Shadow != nil {
			light.Shadower = NewShadowMap(sn.Light.Shadow.Size, sn.Light.Shadow.Cascades)
		}
		node.SetLight(light)
	}

	// Textures (may contain deferred references)
	for texName, texRef := range sn.Textures {
		if strings.HasPrefix(texRef, "$") {
			deferred = append(deferred, deferredTexRef{node: node, texName: texName, ref: texRef})
		}
	}

	// Build children
	for i := range sn.Children {
		child, childCams, childRefs := buildNode(&sn.Children[i])
		node.AddChild(child)
		deferred = append(deferred, childRefs...)
		cameras = append(cameras, childCams...)
	}

	// Camera (defined on the node that owns it)
	if sn.Camera != nil {
		cam := buildCamera(sn.Camera, node)
		cameras = append(cameras, cameraEntry{parent: node, camera: cam})
	}

	return node, cameras, deferred
}

func buildCamera(cd *CameraDef, sceneNode *Node) *Camera {
	projType := PerspectiveProjection
	if cd.Projection == "orthographic" {
		projType = OrthographicProjection
	}

	cam := NewCamera(cd.Name, projType)

	if cd.AutoReshape != nil {
		cam.SetAutoReshape(*cd.AutoReshape)
	}
	if cd.AutoFrustum != nil {
		cam.SetAutoFrustum(*cd.AutoFrustum)
	}
	if cd.FOV != 0 {
		cam.SetVerticalFieldOfView(cd.FOV)
	}
	if cd.ClearColor != [4]float32{} {
		cam.SetClearColor(mgl32.Vec4(cd.ClearColor))
	}
	if len(cd.ClearMode) > 0 {
		var cm ClearMode
		for _, m := range cd.ClearMode {
			switch m {
			case "color":
				cm |= ClearColor
			case "depth":
				cm |= ClearDepth
			}
		}
		cam.SetClearMode(cm)
	}
	if cd.ClipDistance != [2]float64{} {
		cam.SetClipDistance(mgl64.Vec2(cd.ClipDistance))
	}

	cam.SetRenderOrder(cd.RenderOrder)

	if cd.Position != [3]float64{} {
		cam.Node().Translate(mgl64.Vec3(cd.Position))
	}

	// Input component
	if cd.Input != "" {
		if ic := LookupInputComponent(cd.Input); ic != nil {
			cam.Node().SetInputComponent(ic)
		}
	}

	// Render technique
	if cd.Technique != "" {
		if tech := LookupRenderTechnique(cd.Technique); tech != nil {
			cam.SetRenderTechnique(tech)
		}
	}

	// Framebuffer
	if cd.Framebuffer != nil {
		fb := GetRenderer().NewFramebuffer()
		cam.SetFramebuffer(fb)

		windowSize := GetWindowManager().WindowSize()
		w := uint32(windowSize.X())
		h := uint32(windowSize.Y())

		if cd.Framebuffer.Color0 != nil {
			tex := createAttachmentTexture(cd.Framebuffer.Color0, w, h)
			fb.SetColorAttachment(0, tex)
		}
		if cd.Framebuffer.Depth != nil {
			tex := createAttachmentTexture(cd.Framebuffer.Depth, w, h)
			fb.SetDepthAttachment(tex)
		}
	}

	// For orthographic cameras without autoReshape, set viewport explicitly
	if projType == OrthographicProjection && (cd.AutoReshape == nil || !*cd.AutoReshape) {
		windowSize := GetWindowManager().WindowSize()
		cam.SetViewport(mgl32.Vec4{0, 0, windowSize.X(), windowSize.Y()})
		cam.Reshape(windowSize)
	}

	cam.SetScene(sceneNode)
	return cam
}

func createAttachmentTexture(ad *AttachmentDef, width, height uint32) *Texture {
	desc := TextureDescriptor{
		Width:  width,
		Height: height,
		Target: TextureTarget2D,
	}

	switch ad.Format {
	case "rgba16f":
		desc.Format = TextureFormatRGBA
		desc.SizedFormat = TextureSizedFormatRGBA16F
		desc.ComponentType = TextureComponentTypeFLOAT
	case "rgba8":
		desc.Format = TextureFormatRGBA
		desc.SizedFormat = TextureSizedFormatRGBA8
		desc.ComponentType = TextureComponentTypeUNSIGNEDBYTE
	case "depth32f":
		desc.Format = TextureFormatDEPTH
		desc.SizedFormat = TextureSizedFormatDEPTH32F
		desc.ComponentType = TextureComponentTypeFLOAT
	default:
		glog.Warningf("Scene: unknown attachment format %q, defaulting to rgba8", ad.Format)
		desc.Format = TextureFormatRGBA
		desc.SizedFormat = TextureSizedFormatRGBA8
		desc.ComponentType = TextureComponentTypeUNSIGNEDBYTE
	}

	switch ad.Filter {
	case "linear":
		desc.Filter = TextureFilterLinear
	case "nearest":
		desc.Filter = TextureFilterNearest
	case "mipmapLinear":
		desc.Filter = TextureFilterMipmapLinear
		desc.Mipmaps = true
	default:
		desc.Filter = TextureFilterNearest
	}

	switch ad.Wrap {
	case "clampEdge":
		desc.WrapMode = TextureWrapModeClampEdge
	case "repeat":
		desc.WrapMode = TextureWrapModeRepeat
	default:
		desc.WrapMode = TextureWrapModeClampEdge
	}

	return GetRenderer().NewTexture(desc, nil)
}

// resolveTextureRef resolves a "$CameraName.framebuffer.color0" style reference.
func resolveTextureRef(ref string, cameraMap map[string]*Camera) *Texture {
	// Format: $CameraName.framebuffer.color0 or $CameraName.framebuffer.depth
	ref = strings.TrimPrefix(ref, "$")
	parts := strings.Split(ref, ".")
	if len(parts) != 3 || parts[1] != "framebuffer" {
		return nil
	}

	cam, ok := cameraMap[parts[0]]
	if !ok || cam.Framebuffer() == nil {
		return nil
	}

	switch parts[2] {
	case "color0":
		return cam.Framebuffer().ColorAttachment(0)
	case "depth":
		return cam.Framebuffer().DepthAttachment()
	default:
		return nil
	}
}
