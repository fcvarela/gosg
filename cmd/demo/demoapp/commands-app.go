package demoapp

import "github.com/fcvarela/gosg/core"

type clientApplicationQuitCommand struct{}

func (c *clientApplicationQuitCommand) Run(ac core.ClientApplication) {
	ac.Stop()
}

type clientApplicationToggleDebugMenuCommand struct{}

func (c *clientApplicationToggleDebugMenuCommand) Run(ac core.ClientApplication) {
	sm := core.GetSceneManager()
	frontScene := sm.FrontScene()

	switch frontScene.Name() {
	case "debugMenu":
		sm.PopScene()
	case "Demo1":
		debugMenu := new(demo1DebugMenuInputComponent)
		debugMenu.shadowTextures = getDemoSceneShadowTextures(frontScene)
		sm.PushScene(core.NewIMGUIScene("debugMenu", debugMenu))
	}
}
