package demoapp

import (
	"fmt"

	"github.com/fcvarela/gosg/core"
	"github.com/go-gl/mathgl/mgl32"
)

type demo1DebugMenuInputComponent struct {
	shadowTextures []core.Texture
	mainScene      *core.Scene
}

func (u *demo1DebugMenuInputComponent) Run(n *core.Node) []core.NodeCommand {
	renderLog := core.GetRenderSystem().RenderLog()
	timerManager := core.GetTimerManager()
	imguiSystem := core.GetIMGUISystem()

	imguiSystem.StartFrame(timerManager.Dt())

	// Metrics window
	imguiSystem.SetNextWindowPos(mgl32.Vec2{0.0, 0.0})
	imguiSystem.SetNextWindowSize(mgl32.Vec2{320.0, core.GetWindowManager().WindowSize()[1]})
	if imguiSystem.Begin("Inspector", core.WindowFlagsNoCollapse|core.WindowFlagsNoResize|core.WindowFlagsNoMove) {
		if imguiSystem.CollapsingHeader("Frame Times") {
			frameHistogram := timerManager.Histogram()
			fpsLabel := fmt.Sprintf("%.0f FPS", timerManager.AvgFPS())
			imguiSystem.PlotHistogram(fpsLabel, frameHistogram.Values, 0.0, frameHistogram.Max, mgl32.Vec2{0.0, 60.0})
		}

		if imguiSystem.CollapsingHeader("ShadowTextures") {
			imguiSystem.Image(u.shadowTextures[0], mgl32.Vec2{128.0, 128.0})
			imguiSystem.Image(u.shadowTextures[1], mgl32.Vec2{128.0, 128.0})
			imguiSystem.Image(u.shadowTextures[2], mgl32.Vec2{128.0, 128.0})
		}

		if imguiSystem.CollapsingHeader("RenderLog") {
			imguiSystem.Text(renderLog)
		}

		if imguiSystem.CollapsingHeader("Debug Nodes") {
			u.mainScene.Camera("GeometryPassCamera").SetRenderTechnique(core.DebugRenderTechnique)
		} else {
			u.mainScene.Camera("GeometryPassCamera").SetRenderTechnique(core.DefaultRenderTechnique)
		}
	}

	imguiSystem.End()

	// Done
	imguiSystem.EndFrame()
	return nil
}
