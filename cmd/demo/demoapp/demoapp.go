package demoapp

import (
	"github.com/fcvarela/gosg/core"
	"github.com/golang/glog"
)

type demoApp struct {
	done           bool
	inputComponent core.ClientApplicationInputComponent
}

// NewClientApplication returns an initalized demApp which implements the core.ClientApplication interface
func NewClientApplication() core.ClientApplication {
	// initialize and set app (so we don't have to guess systems type)
	c := demoApp{}

	// use the default inputcomponent do deal with ESC->quit.
	c.inputComponent = new(applicationInputComponent)

	// push the main scene into the scenemanager
	core.GetSceneManager().PushScene(makeDemoScene())

	// return
	return &c
}

// InputComponent implements the core.ClientApplication interface
func (c *demoApp) InputComponent() core.ClientApplicationInputComponent {
	return c.inputComponent
}

// Stop implements the core.ClientApplication interface
func (c *demoApp) Stop() {
	c.done = true
	glog.Info("Stopping")
}

// Done implements the core.ClientApplication interface
func (c *demoApp) Done() bool {
	return c.done
}
