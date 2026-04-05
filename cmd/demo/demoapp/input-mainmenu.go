package demoapp

import (
	"fmt"

	"github.com/fcvarela/gosg/core"
	"github.com/go-gl/mathgl/mgl32"
)

type demo1DebugMenuInputComponent struct {
	shadowTextures []*core.Texture
	mainScene      *core.Scene
	light          *core.Light
}

func (u *demo1DebugMenuInputComponent) Run(n *core.Node) []core.NodeCommand {
	timerManager := core.GetTimerManager()
	imguiSystem := core.GetIMGUISystem()

	imguiSystem.StartFrame(timerManager.Dt())

	// Metrics window
	imguiSystem.SetNextWindowPos(mgl32.Vec2{10.0, 10.0})

	imguiSystem.SetNextWindowSize(mgl32.Vec2{320.0, 640})
	if imguiSystem.Begin("Inspector", core.WindowFlagsNoCollapse|core.WindowFlagsNoResize|core.WindowFlagsNoMove) {
		if imguiSystem.CollapsingHeader("Frame Times") {
			frameHistogram := timerManager.Histogram()
			fpsLabel := fmt.Sprintf("%.0f FPS", timerManager.AvgFPS())
			imguiSystem.PlotHistogram(fpsLabel, frameHistogram.Values, 0.0, frameHistogram.Max, mgl32.Vec2{0.0, 60.0})
		}

		if imguiSystem.CollapsingHeader("Shadows") {
			if u.light != nil {
				imguiSystem.SliderFloat("Shadow Bias", &u.light.ShadowBias, 0.0, 0.1)
			}
		}

		if imguiSystem.CollapsingHeader("Render Stats") {
			stats := core.GetRenderer().Stats()
			imguiSystem.Text(fmt.Sprintf("Render Passes:    %d", stats.RenderPasses))
			imguiSystem.Text(fmt.Sprintf("Pipeline Switches: %d", stats.PipelineSwitches))
			imguiSystem.Text(fmt.Sprintf("Draw Calls:       %d", stats.DrawCalls))
			imguiSystem.Text(fmt.Sprintf("Batches:          %d", stats.Batches))
			imguiSystem.Text(fmt.Sprintf("Instances Drawn:  %d", stats.InstancesDrawn))
			imguiSystem.Text(fmt.Sprintf("Flushes:          %d", stats.Flushes))
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
