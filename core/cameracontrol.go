package core

import (
	"math"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl64"
)

// MouseCameraMoveCommand is a utility command for simple camera movement.
type MouseCameraMoveCommand struct {
	direction mgl64.Vec3
}

// Run implements the NodeCommand interface
func (mc MouseCameraMoveCommand) Run(node *Node) {
	node.Translate(mc.direction)
}

// MouseCameraRotateCommand is a utility command for simple camera movement.
type MouseCameraRotateCommand struct {
	angle float64
	axis  mgl64.Vec3
}

// Run implements the NodeCommand interface
func (rc MouseCameraRotateCommand) Run(node *Node) {
	node.Rotate(rc.angle, rc.axis)
}

// MouseCameraInputComponent is a utility inputcomponent for simple camera movement.
type MouseCameraInputComponent struct {
	velocityExponent float64
	velocity         float64
}

// NewMouseCameraInputComponent returns a default inputcomponent for use with camera nodes which
// uses the mouse wheel to set the camera's velocity on 10x increments (world units/second).
func NewMouseCameraInputComponent() *MouseCameraInputComponent {
	mic := new(MouseCameraInputComponent)
	return mic
}

// Run implements the InputComponent interface.
func (ic *MouseCameraInputComponent) Run(node *Node) []NodeCommand {
	state := *inputManager.State()

	var commands []NodeCommand

	// keyboard input
	movementMap := make(map[glfw.Key]mgl64.Vec3)
	movementMap[glfw.KeyW] = mgl64.Vec3{0.0, 0.0, -1.0}
	movementMap[glfw.KeyS] = mgl64.Vec3{0.0, 0.0, +1.0}
	movementMap[glfw.KeyA] = mgl64.Vec3{-1.0, 0.0, 0.0}
	movementMap[glfw.KeyD] = mgl64.Vec3{+1.0, 0.0, 0.0}
	movementMap[glfw.KeyQ] = mgl64.Vec3{0.0, +1.0, 0.0}
	movementMap[glfw.KeyZ] = mgl64.Vec3{0.0, -1.0, 0.0}

	var direction = mgl64.Vec3{0.0, 0.0, 0.0}
	for k, v := range movementMap {
		if state.Keys.Active[k] {
			direction = direction.Add(v)
		}
	}

	if state.Keys.Valid {
		if direction.Len() > 0.0 {
			dtfactor := (ic.velocity) * timerManager.Dt()
			commands = append(commands, MouseCameraMoveCommand{direction.Mul(dtfactor)})
		}
	}

	// mouse movement
	if state.Mouse.Valid {
		if state.Mouse.Position.Valid {
			pitch, yaw := -state.Mouse.Position.DistY, -state.Mouse.Position.DistX
			commands = append(commands, MouseCameraRotateCommand{5.0 * timerManager.Dt() * pitch, mgl64.Vec3{1.0, 0.0, 0.0}})
			commands = append(commands, MouseCameraRotateCommand{5.0 * timerManager.Dt() * yaw, mgl64.Vec3{0.0, 1.0, 0.0}})
		}

		// speed from scroll: doesn't generate commands, affects internal state only
		if state.Mouse.Scroll.Valid {
			ic.velocityExponent += -state.Mouse.Scroll.Y

			if ic.velocityExponent >= 0 {
				ic.velocity = math.Pow(2.0, ic.velocityExponent)
			} else {
				ic.velocity = 0.0
				ic.velocityExponent = -1.0
			}
		}
	}

	return commands
}
