// Package physics contains subpackages which implement the core.PhysicsSystem interface. You can only set one per
// client application, you can only do it once, and if you wish to provide an implementation, your init method must
// register the type implementing the interface by calling core.SetPhysicsSystem(i).
package physics
