// Package audio contains subpackages which implement the core.AudioSystem interface. You can only set one per client
// application, you can only do it once, and if you wish to provide an implementation, your init method must register
// the type implementing the interface by calling core.SetAudioSystem(i).
package audio
