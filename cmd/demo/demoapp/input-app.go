package demoapp

import (
	"github.com/fcvarela/gosg/core"
	"github.com/go-gl/glfw/v3.2/glfw"
)

// ApplicationInputComponent implements InputComponent
type applicationInputComponent struct{}

// Run checks the InputSystem for actionable state and returns commands
func (ic *applicationInputComponent) Run() (commands []core.ClientApplicationCommand) {
	// check for quit key, append to command list
	state := *core.GetInputManager().State()

	if state.Keys.Active[glfw.KeyEscape] {
		commands = append(commands, new(clientApplicationQuitCommand))
	}

	// key-up, after down
	if state.Keys.Released[glfw.KeyE] {
		commands = append(commands, new(clientApplicationToggleDebugMenuCommand))
	}

	return commands
}
