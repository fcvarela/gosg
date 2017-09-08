package core

import (
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/golang/glog"
)

// WindowConfig is used by client applications to request a specific video mode
// from a monitor by calling InitWindow and passing it as an argument.
type WindowConfig struct {
	Name              string
	Monitor           *glfw.Monitor
	Width, Height, Hz int
	Fullscreen        bool
	Vsync             int
}

// WindowManager exposes windowing to client applications
type WindowManager struct {
	window         *glfw.Window
	cfg            WindowConfig
	cursorPosition mgl64.Vec2
}

var (
	windowManager *WindowManager
)

func init() {
	windowManager = &WindowManager{}

	err := glfw.Init()
	if err != nil {
		glog.Fatal(err)
	}
}

// GetWindowManager Returns the global WindowManager
func GetWindowManager() *WindowManager {
	return windowManager
}

// SetWindowConfig Sets the config for WindowManager created windows
func (w *WindowManager) SetWindowConfig(cfg WindowConfig) {
	w.cfg = cfg
}

// MakeWindow created the window manager's window using the previously passed config
func (w *WindowManager) MakeWindow() {
	glog.Info("Preparing to create window, will ask renderSystem for any tips")
	renderSystem.PreMakeWindow()

	var err error
	if w.cfg.Fullscreen {
		w.window, err = glfw.CreateWindow(w.cfg.Width, w.cfg.Height, w.cfg.Name, w.cfg.Monitor, nil)
	} else {
		w.window, err = glfw.CreateWindow(w.cfg.Width, w.cfg.Height, w.cfg.Name, nil, nil)
	}
	if err != nil {
		glog.Fatal(err)
	}

	w.window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	w.installCallbacks()

	renderSystem.PostMakeWindow(w.cfg, w.window)
}

// WindowSize the window size
func (w *WindowManager) WindowSize() mgl32.Vec2 {
	return mgl32.Vec2{float32(w.cfg.Width), float32(w.cfg.Height)}
}

// ShouldClose whether a stop signal is being processed
func (w *WindowManager) ShouldClose() bool {
	return w.window.ShouldClose()
}

func (w *WindowManager) installCallbacks() {
	w.window.SetKeyCallback(inputManager.KeyCallback)
	w.window.SetMouseButtonCallback(inputManager.MouseButtonCallback)
	w.window.SetCursorPosCallback(inputManager.MouseMoveCallback)
	w.window.SetScrollCallback(inputManager.MouseScrollCallback)
}

func (w *WindowManager) closeWindow() {
	glog.Info("Stopping")
	glfw.Terminate()
}

// CursorPosition reports the current cursor position in window coordinates
func (w *WindowManager) CursorPosition() (float64, float64) {
	return w.cursorPosition.X(), w.cursorPosition.Y()
}
