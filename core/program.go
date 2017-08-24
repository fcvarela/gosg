package core

// Program is an interface which wraps a GPU program (OpenGL, etc). This will be removed soon and programs
// will be accessed by name handle via the resource system or abstracted in opaque material definitions.
type Program interface {
	Name() string
}
