// Package dearimgui provides an implementation of core.IMGUISystem by wrapping the C++ library DearIMGUI
package dearimgui

// #include "gosg_imgui.h"
// #cgo windows LDFLAGS: -Wl,--allow-multiple-definition -limm32
import "C"
import (
	"unsafe"

	"github.com/fcvarela/gosg/core"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/golang/glog"
)

func init() {
	core.SetIMGUISystem(&IMGUISystem{})
}

// IMGUISystem provides an implementation of the core.IMGUISystem interface by wrapping the C++ library DearIMGUI.
type IMGUISystem struct {
	texture core.Texture
}

type textureData struct {
	width   int
	height  int
	payload []byte
}

var (
	displaySize mgl32.Vec2
)

func (i *IMGUISystem) getTextureData() textureData {
	var width, height C.int
	payload := unsafe.Pointer(C.get_texture_data(&width, &height))

	bufSize := int(width) * int(height) * 4
	return textureData{int(width), int(height), C.GoBytes(payload, C.int(bufSize))}
}

// Start implements the core.IMGUISystem interface
func (i *IMGUISystem) Start() {
	tdata := i.getTextureData()
	textureDescriptor := core.TextureDescriptor{
		Width:         uint32(tdata.width),
		Height:        uint32(tdata.height),
		Mipmaps:       false,
		Target:        core.TextureTarget2D,
		Format:        core.TextureFormatRGBA,
		SizedFormat:   core.TextureSizedFormatRGBA8,
		ComponentType: core.TextureComponentTypeUNSIGNEDBYTE,
		Filter:        core.TextureFilterLinear,
		WrapMode:      core.TextureWrapModeClampEdge,
	}
	i.texture = core.GetRenderSystem().NewTexture(textureDescriptor, tdata.payload)
	if i.texture == nil {
		glog.Fatal("Cannot set nil texture")
	}
	C.set_texture_id(i.texture.Handle())
}

// Stop implements the core.IMGUISystem interface
func (i *IMGUISystem) Stop() {

}

// Begin implements the core.IMGUISystem interface
func (i *IMGUISystem) Begin(name string, flags core.WindowFlags) bool {
	return int(C.begin(C.CString(name), C.int(flags))) == 1
}

// End implements the core.IMGUISystem interface
func (i *IMGUISystem) End() {
	C.end()
}

// CollapsingHeader implements the core.IMGUISystem interface
func (i *IMGUISystem) CollapsingHeader(name string) bool {
	return int(C.collapsing_header(C.CString(name))) == 1
}

// PlotHistogram implements the core.IMGUISystem interface
func (i *IMGUISystem) PlotHistogram(name string, values []float32, minScale, maxScale float32, size mgl32.Vec2) {
	C.plot_histogram(C.CString(name), (*C.float)(unsafe.Pointer(&values[0])), C.int(len(values)), C.float(minScale), C.float(maxScale), (*C.float)(unsafe.Pointer(&size[0])))

}

// Image implements the core.IMGUISystem interface
func (i IMGUISystem) Image(texture core.Texture, size mgl32.Vec2) {
	if texture == nil {
		glog.Fatal("Cannot draw nil texture")
	}
	C.image(texture.Handle(), (*C.float)(unsafe.Pointer(&size[0])))
}

// Text implements core.IMGUISystem interface
func (i IMGUISystem) Text(data string) {
	C.text(C.CString(data))
}

// SetNextWindowPos implements the core.IMGUISystem interface
func (i *IMGUISystem) SetNextWindowPos(pos mgl32.Vec2) {
	C.set_next_window_pos(C.float(pos[0]), C.float(pos[1]))
}

// SetNextWindowSize implements the core.IMGUISystem interface
func (i *IMGUISystem) SetNextWindowSize(size mgl32.Vec2) {
	C.set_next_window_size(C.float(size[0]), C.float(size[1]))
}

// StartFrame implements the core.IMGUISystem interface
func (i *IMGUISystem) StartFrame(dt float64) {
	state := core.GetInputManager().State()
	size := core.GetWindowManager().WindowSize()
	i.SetDisplaySize(size)
	i.SetMousePosition(core.GetWindowManager().CursorPosition())
	i.SetMouseButtons(
		state.Mouse.Buttons.Active[glfw.MouseButton1],
		state.Mouse.Buttons.Active[glfw.MouseButton2],
		state.Mouse.Buttons.Active[glfw.MouseButton3])
	i.SetMouseScrollPosition(state.Mouse.Scroll.X, state.Mouse.Scroll.Y)

	C.set_dt(C.double(dt))
	C.frame_new()
}

// DisplaySize implements the core.IMGUISystem interface
func (i *IMGUISystem) DisplaySize() mgl32.Vec2 {
	return displaySize
}

// SetDisplaySize implements the core.IMGUISystem interface
func (i *IMGUISystem) SetDisplaySize(s mgl32.Vec2) {
	displaySize = s
	C.set_display_size(C.float(s[0]), C.float(s[1]))
}

// SetMousePosition implements the core.IMGUISystem interface
func (i *IMGUISystem) SetMousePosition(x, y float64) {
	C.set_mouse_position(C.double(x), C.double(y))
}

// SetMouseButtons implements the core.IMGUISystem interface
func (i *IMGUISystem) SetMouseButtons(b0, b1, b2 bool) {
	var ib0, ib1, ib2 int

	if b0 {
		ib0 = 1
	}

	if b1 {
		ib1 = 1
	}

	if b2 {
		ib2 = 1
	}

	C.set_mouse_buttons(C.int(ib0), C.int(ib1), C.int(ib2))
}

// SetMouseScrollPosition implements the core.IMGUISystem interface
func (i *IMGUISystem) SetMouseScrollPosition(xoffset, yoffset float64) {
	C.set_mouse_scroll_position(C.double(xoffset), C.double(yoffset))
}

// EndFrame implements the core.IMGUISystem interface
func (i *IMGUISystem) EndFrame() {
	C.render()

	state := core.GetInputManager().State()
	state.SetMouseValid(false)
	state.SetKeysValid(false)
}

// WantsCaptureMouse implements the core.IMGUISystem interface
func (i *IMGUISystem) WantsCaptureMouse() bool {
	return int(C.wants_capture_mouse()) == 1
}

// WantsCaptureKeyboard implements the core.IMGUISystem interface
func (i *IMGUISystem) WantsCaptureKeyboard() bool {
	return int(C.wants_capture_keyboard()) == 1
}

// DrawData implements the core.DrawData interface
type DrawData struct {
	drawData unsafe.Pointer
}

// GetDrawData implements the core.IMGUISystem interface
func (i *IMGUISystem) GetDrawData() core.IMGUIDrawData {
	return &DrawData{C.get_draw_data()}
}

// CommandListCount implements the core.IMGUISystem interface
func (d *DrawData) CommandListCount() int {
	return int(C.get_cmdlist_count(d.drawData))
}

// GetCommandList implements the core.IMGUISystem interface
func (d *DrawData) GetCommandList(index int) *core.IMGUICommandList {
	cCmdList := C.get_cmdlist(d.drawData, C.int(index))

	cmdList := &core.IMGUICommandList{
		CmdBufferSize:    int(cCmdList.commandBufferSize),
		VertexBufferSize: int(cCmdList.vertexBufferSize),
		IndexBufferSize:  int(cCmdList.indexBufferSize),
		VertexPointer:    unsafe.Pointer(cCmdList.vertexPointer),
		IndexPointer:     unsafe.Pointer(cCmdList.indexPointer),
		Commands:         make([]core.IMGUICommand, int(cCmdList.commandBufferSize)),
	}

	for c := 0; c < cmdList.CmdBufferSize; c++ {
		cmd := core.IMGUICommand{}

		userTexturePtr := C.get_cmdlist_cmd(d.drawData, C.int(index), C.int(c),
			(*C.int)(unsafe.Pointer(&cmd.ElementCount)),
			(*C.float)(unsafe.Pointer(&cmd.ClipRect[0])))

		cmd.TextureID = userTexturePtr
		cmdList.Commands[c] = cmd
	}
	return cmdList
}
