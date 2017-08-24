// Package openal provides an implementation of a core.AudioSystem by wrapping OpenAL
package openal

import (
	"github.com/fcvarela/gosg/core"
	"github.com/golang/glog"
)

// AudioSystem implements the core.AudioSystem interface by wrapping OpenAL
type AudioSystem struct {
	//Device  *al.Device
	//Context *al.Context
}

func init() {
	core.SetAudioSystem(&AudioSystem{})
}

// Start implements the core.AudioSystem interface
func (a *AudioSystem) Start() {
	glog.Info("Starting")
}

// Step implements the core.AudioSystem interface
func (a *AudioSystem) Step() {

}

// Stop implements the core.AudioSystem interface
func (a *AudioSystem) Stop() {
	glog.Info("Stopping")
}
