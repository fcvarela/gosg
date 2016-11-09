package core

import (
	"unsafe"

	"github.com/fcvarela/gosg/protos"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
)

// WindowFlags holds a set of properties for window layout.
type WindowFlags int32

const (
	// WindowFlagsNoTitleBar hides the window title bar/
	WindowFlagsNoTitleBar WindowFlags = 1 << iota

	// WindowFlagsNoResize disallows window resizing.
	WindowFlagsNoResize

	//WindowFlagsNoMove disallows moving the window.
	WindowFlagsNoMove

	// WindowFlagsNoScrollbar hides the scrollbar.
	WindowFlagsNoScrollbar

	// WindowFlagsNoScrollWithMouse disallows using wheel for scrolling.
	WindowFlagsNoScrollWithMouse

	// WindowFlagsNoCollapse disables window collapsing.
	WindowFlagsNoCollapse

	// WindowFlagsAlwaysAutoResize enables window auto resizing.
	WindowFlagsAlwaysAutoResize

	// WindowFlagsShowBorders shows the window border.
	WindowFlagsShowBorders

	// WindowFlagsNoSavedSettings disabled saving/using saved window settings.
	WindowFlagsNoSavedSettings

	// WindowFlagsNoInputs disables input handling.
	WindowFlagsNoInputs

	// WindowFlagsMenuBar enables a menubar on the window.
	WindowFlagsMenuBar

	// WindowFlagsHorizontalScrollbar enables horizontal scrolling.
	WindowFlagsHorizontalScrollbar

	// WindowFlagsNoFocusOnAppearing disables auto focus when displayed.
	WindowFlagsNoFocusOnAppearing

	// WindowFlagsNoBringToFrontOnFocus disables moving window to front when focused.
	WindowFlagsNoBringToFrontOnFocus
)

// IMGUISystem is the interface that wraps an immediate mode GUI component.
type IMGUISystem interface {
	// Start is called at application startup. This is where implementations should check
	// for fonts, create their textures, etc.
	Start()

	// Stop is called at application shutdown. Implementations will want to perform cleanup here.
	Stop()

	// StartFrame is called at the draw stage of the runloop.
	StartFrame(dt float64)

	// EndFrame is called at the end of the draw stage of the runloop.
	EndFrame()

	// Begin creates a new widget and returns whether it is visible.
	Begin(name string, flags WindowFlags) bool

	// End closes the current active widget.
	End()

	// DisplaySize returns the current display size.
	DisplaySize() mgl32.Vec2

	// SetDisplaySize sets the display size. This is called by the application on every
	// iteration of the runloop. Implementations should check this for window resizes.
	SetDisplaySize(mgl32.Vec2)

	// SetNextWindowPos is used to request a position for the next window that gets created.
	SetNextWindowPos(mgl32.Vec2)

	// SetNextWindowSize is used to request a size for the next window that gets created.
	SetNextWindowSize(mgl32.Vec2)

	// GetDrawData returns IMGUIDrawData which is used by the RenderSystem to present the UI on screen.
	GetDrawData() IMGUIDrawData

	// CollapsingHeader returns a widget with a collapsing header.
	CollapsingHeader(name string) bool

	//PlotHistogram draws a histogram on a widget using the passed name and values. minScale and maxScale set
	// the Y axis scale and size sets the widgets width and height.
	PlotHistogram(name string, values []float32, minScale, maxScale float32, size mgl32.Vec2)

	// Image draws a texture on a widget with the provided size.
	Image(texture Texture, size mgl32.Vec2)

	// Text displays a text box
	Text(data string)
}

// IMGUICommand represents an individual draw command for the RenderSystem to present a part of the UI.
type IMGUICommand struct {
	ElementCount int
	ClipRect     [4]float32
	TextureID    unsafe.Pointer
}

// IMGUICommandList returns a list of draw commands for the RenderSystem to present the entire UI.
type IMGUICommandList struct {
	CmdBufferSize    int
	VertexBufferSize int
	IndexBufferSize  int
	VertexPointer    unsafe.Pointer
	IndexPointer     unsafe.Pointer
	Commands         []IMGUICommand
}

// IMGUIDrawData is an interface which exposes methods for retrieving command lists from implementations.
type IMGUIDrawData interface {
	CommandListCount() int
	GetCommandList(int) *IMGUICommandList
}

var (
	imguiSystem IMGUISystem
)

// SetIMGUISystem is meant to be called from IMGUISystem implementations on their init method
func SetIMGUISystem(is IMGUISystem) {
	imguiSystem = is
}

// GetIMGUISystem returns the IMGUISystem, thereby exposing it to any package importing core.
func GetIMGUISystem() IMGUISystem {
	return imguiSystem
}

// NewIMGUIScene returns a Scene which draws a UI. Users will want to use this to display a UI on top of
// other scenes.
func NewIMGUIScene(name string, inputComponent InputComponent) *Scene {
	s := NewScene(name)
	s.SetRoot(NewNode("root"))
	s.Root().SetInputComponent(inputComponent)

	size := windowManager.WindowSize()
	camera := NewCamera("MainMenuCamera", OrthographicProjection)
	camera.SetRenderTechnique(IMGUIRenderTechnique)
	camera.SetAutoReshape(true)
	camera.SetClearMode(0)
	camera.SetViewport(mgl32.Vec4{0, 0, size.X(), size.Y()})
	camera.SetVerticalFieldOfView(60.0)
	camera.SetClipDistance(mgl64.Vec2{0.0, 1.0})
	camera.SetRenderOrder(0)
	camera.Reshape(mgl32.Vec2{size.X(), size.Y()})

	// set camera's scene
	camera.SetScene(s.root)

	// add camera to root node
	s.AddCamera(s.root, camera)

	return s
}

// IMGUIRenderTechnique does z pre-pass, diffuse pass, transparency pass
func IMGUIRenderTechnique(camera *Camera, materialBuckets map[*protos.State][]*Node, cmdBuf chan RenderCommand) {
	cmdBuf <- &SetFramebufferCommand{camera.framebuffer}
	cmdBuf <- &SetViewportCommand{camera.viewport}
	cmdBuf <- &ClearCommand{camera.clearMode, camera.clearColor, camera.clearDepth}
	cmdBuf <- &BindUniformBufferCommand{"cameraConstants", camera.constants.buffer}

	var imguiState = resourceManager.State("imgui")
	cmdBuf <- &BindStateCommand{imguiState}
	cmdBuf <- &DrawIMGUICommand{}
}
