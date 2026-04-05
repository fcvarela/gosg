package core

import (
	"math"

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
	Active map[MouseButton]bool
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
	Mods     map[Key]bool
	Active   map[Key]bool
	Released map[Key]bool
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

// InputManager wraps global input state.
type InputManager struct {
	state InputState
}

// InputComponent is an interface which returns NodeCommands from nodes.
type InputComponent interface {
	Run(node *Node) []NodeCommand
}

var (
	inputManager *InputManager
)

func init() {
	inputManager = &InputManager{}
	inputManager.state.Keys.Active = make(map[Key]bool)
	inputManager.state.Keys.Released = make(map[Key]bool)
	inputManager.state.Mouse.Buttons.Active = make(map[MouseButton]bool)
}

// GetInputManager returns the manager.
func GetInputManager() *InputManager {
	return inputManager
}

// State returns the manager's input state.
func (i *InputManager) State() *InputState {
	return &i.state
}

// reset resets all input state and marks substates as invalid.
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

// HandleKeyEvent is called by the window system to register key events.
func (i *InputManager) HandleKeyEvent(key Key, pressed bool) {
	i.state.Keys.Valid = true

	if pressed {
		i.state.Keys.Active[key] = true
		i.state.Keys.Released[key] = false
	} else {
		i.state.Keys.Active[key] = false
		i.state.Keys.Released[key] = true
	}
}

// HandleMouseButton is called by the window system to register mouse button events.
func (i *InputManager) HandleMouseButton(button MouseButton, pressed bool) {
	i.state.Mouse.Buttons.Valid = true
	i.state.Mouse.Buttons.Active[button] = pressed
}

// HandleMouseScroll is called by the window system to register mouse scroll events.
func (i *InputManager) HandleMouseScroll(x, y float64) {
	if GetPlatform() == PlatformLinux || GetPlatform() == PlatformWindows {
		y = -y
	}

	i.state.Mouse.Valid = true
	i.state.Mouse.Scroll.Valid = true
	i.state.Mouse.Scroll.X = x
	i.state.Mouse.Scroll.Y = y
}

// HandleMouseMove is called by the window system to register mouse move events.
func (i *InputManager) HandleMouseMove(x, y, relX, relY float64) {
	i.state.Mouse.Valid = true
	i.state.Mouse.Position.Valid = true

	i.state.Mouse.Position.X = x
	i.state.Mouse.Position.Y = y
	i.state.Mouse.Position.DistX = relX
	i.state.Mouse.Position.DistY = relY

	windowManager.cursorPosition = mgl64.Vec2{
		math.Min(math.Max(0.0, windowManager.cursorPosition.X()+relX), float64(windowManager.cfg.Width)),
		math.Min(math.Max(0.0, windowManager.cursorPosition.Y()+relY), float64(windowManager.cfg.Height)),
	}
}
