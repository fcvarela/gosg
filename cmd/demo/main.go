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
	_ "github.com/fcvarela/gosg/resource/filesystem"
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

	core.GetWindowManager().SetWindowConfig(core.WindowConfig{
		Name:       "Demo",
		Width:      1024,
		Height:     768,
		Fullscreen: true,
		Vsync:      1,
	})

	// start main loop, pass the appcontroller init function
	app.Start(demoapp.NewClientApplication)
}
