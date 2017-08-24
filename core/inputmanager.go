package core

import (
	"math"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl64"
)

// MousePositionState holds the mouse position information.
type MousePositionState struct {
	Valid bool
	X     float64
	Y     float64
	DistX float64
	DistY float64
}

// MouseButtonState holds the mouse button state.
type MouseButtonState struct {
	Valid  bool
	Active map[glfw.MouseButton]bool
	Action int
}

// MouseScrollState holds the mouse scroll state.
type MouseScrollState struct {
	Valid bool
	X     float64
	Y     float64
}

// MouseState holds mouse input state.
type MouseState struct {
	Valid    bool
	Position MousePositionState
	Scroll   MouseScrollState
	Buttons  MouseButtonState
}

// KeyState holds key input state.
type KeyState struct {
	Valid    bool
	Mods     map[glfw.Key]bool
	Active   map[glfw.Key]bool
	Released map[glfw.Key]bool
}

// InputState wraps mouse and keys input state.
type InputState struct {
	Mouse MouseState
	Keys  KeyState
}

// SetMouseValid sets the mouse state as valid. It will not be processed unless this is set.
func (i *InputState) SetMouseValid(valid bool) {
	i.Mouse.Valid = valid
}

// SetKeysValid sets the key state as valid. It will not be processed unless this is set.
func (i *InputState) SetKeysValid(valid bool) {
	i.Keys.Valid = valid
}

// InputManager wraps global input state. WindowSystem implementations use the manager to expose
// input state to the system.
type InputManager struct {
	state InputState
}

// InputComponent is an interface which returns NodeCommands from nodes. Each node may have its own
// input component which checks the manager for input and determines what commands should be output.
type InputComponent interface {
	// Run returns commands from a given node to itself.
	Run(node *Node) []NodeCommand
}

var (
	inputManager *InputManager
)

func init() {
	inputManager = &InputManager{}
	inputManager.state.Keys.Active = make(map[glfw.Key]bool)
	inputManager.state.Keys.Released = make(map[glfw.Key]bool)
	inputManager.state.Mouse.Buttons.Active = make(map[glfw.MouseButton]bool)
}

// GetInputManager returns the manager.
func GetInputManager() *InputManager {
	return inputManager
}

// State returns the manager's input state.
func (i *InputManager) State() *InputState {
	return &i.state
}

// Reset resets all input state and marks substates as invalid.
func (i *InputManager) reset() {
	for j := range i.state.Keys.Released {
		i.state.Keys.Released[j] = false
	}

	i.state.Mouse.Valid = false
	i.state.Mouse.Position.Valid = false
	i.state.Mouse.Buttons.Valid = false
	i.state.Mouse.Position.DistX = 0.0
	i.state.Mouse.Position.DistY = 0.0
	i.state.Mouse.Scroll = MouseScrollState{false, 0.0, 0.0}
}

// KeyCallback is called by windowsystems to register key events.
func (i *InputManager) KeyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	i.state.Keys.Valid = true

	if action == glfw.Press || action == glfw.Repeat {
		i.state.Keys.Active[key] = true
		i.state.Keys.Released[key] = false
	} else if action == glfw.Release {
		i.state.Keys.Active[key] = false
		i.state.Keys.Released[key] = true
	}
}

// MouseButtonCallback is called by windowsystems to register mouse button events.
func (i *InputManager) MouseButtonCallback(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	i.state.Mouse.Buttons.Valid = true

	if action == glfw.Press {
		i.state.Mouse.Buttons.Active[button] = true
	} else {
		i.state.Mouse.Buttons.Active[button] = false
	}
}

// MouseScrollCallback is called by windowsystems to register mouse scroll events.
func (i *InputManager) MouseScrollCallback(window *glfw.Window, x, y float64) {
	if GetPlatform() == PlatformLinux || GetPlatform() == PlatformWindows {
		y = -y
	}

	i.state.Mouse.Valid = true
	i.state.Mouse.Scroll.Valid = true

	i.state.Mouse.Scroll.X = x
	i.state.Mouse.Scroll.Y = y
}

// MouseMoveCallback is called by windowsystems to register mouse move events.
func (i *InputManager) MouseMoveCallback(window *glfw.Window, x, y float64) {
	i.state.Mouse.Valid = true
	i.state.Mouse.Position.Valid = true

	i.state.Mouse.Position.DistX = x - i.state.Mouse.Position.X
	i.state.Mouse.Position.DistY = y - i.state.Mouse.Position.Y

	i.state.Mouse.Position.X = x
	i.state.Mouse.Position.Y = y

	// we also keep track of deltas for cursor position in window
	windowManager.cursorPosition = mgl64.Vec2{
		math.Min(math.Max(0.0, windowManager.cursorPosition.X()+i.state.Mouse.Position.DistX), float64(windowManager.cfg.Width)),
		math.Min(math.Max(0.0, windowManager.cursorPosition.Y()+i.state.Mouse.Position.DistY), float64(windowManager.cfg.Height)),
	}
}
