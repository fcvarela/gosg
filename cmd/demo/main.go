package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"runtime"

	_ "github.com/fcvarela/gosg/audio/openal"
	"github.com/fcvarela/gosg/cmd/demo/demoapp"
	"github.com/fcvarela/gosg/core"
	_ "github.com/fcvarela/gosg/imgui/dearimgui"
	_ "github.com/fcvarela/gosg/physics/bullet"
	_ "github.com/fcvarela/gosg/render/opengl"
	_ "github.com/fcvarela/gosg/resource/filesystem"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang/glog"
)

func init() {
	// expose profiler
	go func() {
		glog.Info(http.ListenAndServe("localhost:6060", nil))
	}()

	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("true")
	glog.Infof("Starting on %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

func main() {
	app := new(core.Application)

	// initialize the window, maybe show an OS-native dialogue here?
	monitors := glfw.GetMonitors()
	vms := monitors[0].GetVideoModes()
	vm := vms[len(vms)-1]

	core.GetWindowManager().SetWindowConfig(core.WindowConfig{
		Name:       "Demo",
		Monitor:    monitors[0],
		Width:      vm.Width / 2,
		Height:     vm.Height / 2,
		Fullscreen: false,
		Vsync:      1,
	})

	// start main loop, pass the appcontroller init function
	app.Start(demoapp.NewClientApplication)
}
