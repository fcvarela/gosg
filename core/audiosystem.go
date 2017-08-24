package core

import "log"

// AudioSystem is the interface that wraps all audio processing logic.
type AudioSystem interface {
	// Start starts the audio system. This is where implementations should detect
	// hardware devices and initialize them.
	Start()

	// Step is called by the application runloop update.
	Step()

	// Stop is called by the application after its main loop returns. Implementations
	// should call their termination/cleanup logic here.
	Stop()
}

var (
	audioSystem AudioSystem
)

// SetAudioSystem should be called by implementations on their init function. It registers
// the implementation as the active audio system.
func SetAudioSystem(a AudioSystem) {
	if audioSystem != nil {
		log.Fatal("Can't replace previously registered audio system. Please make sure you're not importing twice.")
	}
	audioSystem = a
}

// GetAudioSystem returns the registered window system
func GetAudioSystem() AudioSystem {
	return audioSystem
}
