package core

// SceneFile is the top-level YAML structure for a scene file.
type SceneFile struct {
	Name  string      `yaml:"name"`
	Nodes []SceneNode `yaml:"nodes"`
}

// SceneNode describes a node in the scene YAML.
type SceneNode struct {
	Name       string      `yaml:"name"`
	Model      string      `yaml:"model,omitempty"`
	Position   [3]float64  `yaml:"position,omitempty"`
	Rotation   [4]float64  `yaml:"rotation,omitempty"`   // [angle, axisX, axisY, axisZ]
	Scale      [3]float64  `yaml:"scale,omitempty"`
	Pipeline   string      `yaml:"pipeline,omitempty"`
	Cull       string      `yaml:"cull,omitempty"`
	ScreenQuad bool        `yaml:"screenQuad,omitempty"`
	Light      *LightDef   `yaml:"light,omitempty"`
	Camera     *CameraDef  `yaml:"camera,omitempty"`
	Textures   map[string]string `yaml:"textures,omitempty"`
	Children   []SceneNode `yaml:"children,omitempty"`
}

// CameraDef describes a camera in the scene YAML.
type CameraDef struct {
	Name         string         `yaml:"name"`
	Projection   string         `yaml:"projection"`          // "perspective" or "orthographic"
	FOV          float64        `yaml:"fov,omitempty"`
	AutoReshape  *bool          `yaml:"autoReshape,omitempty"`
	AutoFrustum  *bool          `yaml:"autoFrustum,omitempty"`
	ClearColor   [4]float32     `yaml:"clearColor,omitempty"`
	ClearMode    []string       `yaml:"clearMode,omitempty"` // ["color", "depth"]
	ClipDistance  [2]float64     `yaml:"clipDistance,omitempty"`
	RenderOrder  uint8          `yaml:"renderOrder,omitempty"`
	Position     [3]float64     `yaml:"position,omitempty"`
	Input        string         `yaml:"input,omitempty"`
	Technique    string         `yaml:"technique,omitempty"`
	Framebuffer  *FramebufferDef `yaml:"framebuffer,omitempty"`
}

// LightDef describes a light in the scene YAML.
type LightDef struct {
	Color      [3]float32 `yaml:"color"`
	ShadowBias float32    `yaml:"shadowBias,omitempty"`
	Shadow     *ShadowDef `yaml:"shadow,omitempty"`
}

// ShadowDef describes shadow map parameters.
type ShadowDef struct {
	Size     uint32 `yaml:"size"`
	Cascades int    `yaml:"cascades"`
}

// FramebufferDef describes a framebuffer in the scene YAML.
type FramebufferDef struct {
	Color0 *AttachmentDef `yaml:"color0,omitempty"`
	Depth  *AttachmentDef `yaml:"depth,omitempty"`
}

// AttachmentDef describes a framebuffer attachment.
type AttachmentDef struct {
	Format string `yaml:"format"` // "rgba16f", "depth32f", "rgba8", etc.
	Filter string `yaml:"filter"` // "linear", "nearest", "mipmapLinear"
	Wrap   string `yaml:"wrap"`   // "clampEdge", "repeat"
}
