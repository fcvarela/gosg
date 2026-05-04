package core

/*
#cgo pkg-config: sdl3
#include <SDL3/SDL.h>
#include <SDL3/SDL_metal.h>
*/
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/golang/glog"
)

// WindowConfig is used by client applications to request a specific video mode.
type WindowConfig struct {
	Name              string
	MonitorIndex      int
	Width, Height, Hz int
	Fullscreen        bool
	Vsync             int
}

// WindowManager exposes windowing to client applications.
type WindowManager struct {
	window         *C.SDL_Window
	metalView      C.SDL_MetalView
	cfg            WindowConfig
	pixelWidth     int
	pixelHeight    int
	cursorPosition mgl64.Vec2
	shouldClose    bool
}

var (
	windowManager *WindowManager
	windowInitErr error
)

func init() {
	windowManager = &WindowManager{}
}

// InitWindowManager initializes SDL. Call before using WindowManager.
func InitWindowManager() error {
	if windowInitErr != nil {
		return windowInitErr
	}
	if C.SDL_Init(C.SDL_INIT_VIDEO|C.SDL_INIT_EVENTS) == false {
		windowInitErr = fmt.Errorf("SDL_Init failed: %s", C.GoString(C.SDL_GetError()))
		return windowInitErr
	}
	return nil
}

// GetWindowManager returns the global WindowManager.
func GetWindowManager() *WindowManager {
	return windowManager
}

// SetWindowConfig sets the config for WindowManager created windows.
func (w *WindowManager) SetWindowConfig(cfg WindowConfig) {
	w.cfg = cfg
}

// MakeWindow creates the window using the previously passed config.
func (w *WindowManager) MakeWindow() error {
	glog.Info("Creating SDL3 window")

	flags := C.SDL_WINDOW_METAL | C.SDL_WINDOW_HIGH_PIXEL_DENSITY
	if w.cfg.Fullscreen {
		flags |= C.SDL_WINDOW_FULLSCREEN

		// Use native display resolution
		displayID := C.SDL_GetPrimaryDisplay()
		mode := C.SDL_GetDesktopDisplayMode(displayID)
		if mode != nil {
			w.cfg.Width = int(mode.w)
			w.cfg.Height = int(mode.h)
		}
	}

	w.window = C.SDL_CreateWindow(
		C.CString(w.cfg.Name),
		C.int(w.cfg.Width), C.int(w.cfg.Height),
		C.Uint64(flags),
	)
	if w.window == nil {
		return fmt.Errorf("SDL_CreateWindow failed: %s", C.GoString(C.SDL_GetError()))
	}

	// Hide cursor and enable relative mouse mode
	C.SDL_SetWindowRelativeMouseMode(w.window, true)

	// Create Metal view for wgpu surface
	w.metalView = C.SDL_Metal_CreateView(w.window)
	if w.metalView == nil {
		return fmt.Errorf("SDL_Metal_CreateView failed: %s", C.GoString(C.SDL_GetError()))
	}

	// In fullscreen mode, pump events to let macOS finish the fullscreen
	// transition (notch animation, etc.) before querying the final size.
	if w.cfg.Fullscreen {
		C.SDL_SyncWindow(w.window)
	}

	// Query the actual pixel dimensions after the window is fully set up
	var pw, ph C.int
	C.SDL_GetWindowSizeInPixels(w.window, &pw, &ph)
	w.pixelWidth = int(pw)
	w.pixelHeight = int(ph)

	// Update point-space dimensions from actual window size
	var pointW, pointH C.int
	C.SDL_GetWindowSize(w.window, &pointW, &pointH)
	if pointW > 0 && pointH > 0 {
		w.cfg.Width = int(pointW)
		w.cfg.Height = int(pointH)
	}

	return InitRenderer(w.GetMetalLayer(), uint32(w.pixelWidth), uint32(w.pixelHeight))
}

// GetMetalLayer returns the CAMetalLayer pointer for wgpu surface creation.
func (w *WindowManager) GetMetalLayer() unsafe.Pointer {
	return C.SDL_Metal_GetLayer(w.metalView)
}

// WindowSize returns the window size in pixels (for framebuffers/viewports).
func (w *WindowManager) WindowSize() mgl32.Vec2 {
	return mgl32.Vec2{float32(w.pixelWidth), float32(w.pixelHeight)}
}

// WindowSizePoints returns the window size in points (for UI layout).
func (w *WindowManager) WindowSizePoints() mgl32.Vec2 {
	return mgl32.Vec2{float32(w.cfg.Width), float32(w.cfg.Height)}
}

// PixelDensity returns the ratio of pixels to points (e.g., 2.0 on Retina).
func (w *WindowManager) PixelDensity() float32 {
	if w.cfg.Width == 0 {
		return 1.0
	}
	return float32(w.pixelWidth) / float32(w.cfg.Width)
}
func (w *WindowManager) ShouldClose() bool {
	return w.shouldClose
}

// PollEvents processes all pending SDL events and dispatches to InputManager.
func (w *WindowManager) PollEvents() {
	var event C.SDL_Event
	for C.SDL_PollEvent(&event) != false {
		eventType := *(*C.Uint32)(unsafe.Pointer(&event))
		switch eventType {
		case C.SDL_EVENT_QUIT:
			w.shouldClose = true
		case C.SDL_EVENT_WINDOW_CLOSE_REQUESTED:
			w.shouldClose = true
		case C.SDL_EVENT_KEY_DOWN:
			ke := (*C.SDL_KeyboardEvent)(unsafe.Pointer(&event))
			inputManager.HandleKeyEvent(Key(ke.scancode), true)
		case C.SDL_EVENT_KEY_UP:
			ke := (*C.SDL_KeyboardEvent)(unsafe.Pointer(&event))
			inputManager.HandleKeyEvent(Key(ke.scancode), false)
		case C.SDL_EVENT_MOUSE_BUTTON_DOWN:
			me := (*C.SDL_MouseButtonEvent)(unsafe.Pointer(&event))
			inputManager.HandleMouseButton(MouseButton(me.button), true)
		case C.SDL_EVENT_MOUSE_BUTTON_UP:
			me := (*C.SDL_MouseButtonEvent)(unsafe.Pointer(&event))
			inputManager.HandleMouseButton(MouseButton(me.button), false)
		case C.SDL_EVENT_MOUSE_MOTION:
			me := (*C.SDL_MouseMotionEvent)(unsafe.Pointer(&event))
			inputManager.HandleMouseMove(float64(me.x), float64(me.y), float64(me.xrel), float64(me.yrel))
		case C.SDL_EVENT_MOUSE_WHEEL:
			me := (*C.SDL_MouseWheelEvent)(unsafe.Pointer(&event))
			inputManager.HandleMouseScroll(float64(me.x), float64(me.y))
		}
	}
}

// CursorPosition reports the current cursor position in window coordinates.
func (w *WindowManager) CursorPosition() (float64, float64) {
	return w.cursorPosition.X(), w.cursorPosition.Y()
}

func (w *WindowManager) closeWindow() {
	glog.Info("Stopping")
	if renderer != nil {
		renderer.Shutdown()
	}
	if w.metalView != nil {
		C.SDL_Metal_DestroyView(w.metalView)
	}
	if w.window != nil {
		C.SDL_DestroyWindow(w.window)
	}
	C.SDL_Quit()
}

// sdlGetTimeSeconds returns SDL time in seconds (used by TimerManager).
func sdlGetTimeSeconds() float64 {
	return float64(C.SDL_GetTicks()) / 1000.0
}
