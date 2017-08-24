package core

import (
	"runtime"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang/glog"
)

// ClientApplication is the client app, provided by the user
type ClientApplication interface {
	InputComponent() ClientApplicationInputComponent
	Stop()
	Done() bool
}

// ClientApplicationCommand is a single runnable command
type ClientApplicationCommand interface {
	Run(ac ClientApplication)
}

// ClientApplicationInputComponent checks the input system state
// and triggers commands
type ClientApplicationInputComponent interface {
	Run() []ClientApplicationCommand
}

// Application is the top-level runnable gosg App
type Application struct {
	client ClientApplication
}

// Start starts the application runloop by calling all systems/managers Start methods,
// and calling the ClientApp constructor. On runloop termination, the Stop methods are
// called in reverse order.
func (app *Application) Start(acConstructor func() ClientApplication) {
	// we always handle input, window creating and drawing on main thread, keeps most OS happy
	runtime.LockOSThread()

	// make a window
	windowManager.MakeWindow()

	// start subsystems, all on main goroutine, main OS thread
	audioSystem.Start()
	physicsSystem.Start()
	imguiSystem.Start()

	// create the client app, same here
	app.client = acConstructor()

	// start main loop, all systems go
	app.runLoop()

	// done, stop subsystems
	imguiSystem.Stop()
	physicsSystem.Stop()
	audioSystem.Stop()

	// stop managers
	windowManager.closeWindow()
}

func (app *Application) runLoop() {
	glog.Info("Starting runloop...")

	var dt = 1.0 / 60.0
	var start = timerManager.Time()
	var end float64

	for !app.client.Done() && !windowManager.ShouldClose() {
		// run subsystem updates if not paused
		app.update(dt)

		// compute time delta
		end = timerManager.Time()
		dt = end - start

		// safeguard for extreme deltas (breakpoints, suspends)
		if dt > 1.0/10.0 {
			dt = 1.0 / 60.0
		}

		timerManager.SetDt(dt)
		timerManager.setFrameStartTime(end)

		// rotate time
		start = end
	}

	glog.Info("Runloop aborted...")
}

func (app *Application) update(dt float64) {
	// reset input
	inputManager.reset()

	// poll for events
	glfw.PollEvents()
	glfw.PollEvents()
	glfw.PollEvents()

	// update client app
	acCommands := app.client.InputComponent().Run()
	for _, command := range acCommands {
		command.Run(app.client)
	}

	// play audio
	audioSystem.Step()

	// call game object updates
	sceneManager.update(dt)

	// run the culler
	sceneManager.cull()

	// draw
	sceneManager.draw()

	// swap context buffers and poll for input
	windowManager.window.SwapBuffers()
}
