package core

// Key represents a keyboard key.
type Key int

// MouseButton represents a mouse button.
type MouseButton int

// Key constants matching SDL3 scancodes — only the keys we use.
const (
	KeyUnknown Key = -1
	KeyA       Key = 4
	KeyD       Key = 7
	KeyE       Key = 8
	KeyQ       Key = 20
	KeyS       Key = 22
	KeyW       Key = 26
	KeyZ       Key = 29
	KeyEscape  Key = 41
	KeySpace   Key = 44
)

// Mouse button constants matching SDL3.
const (
	MouseButton1 MouseButton = 1 // left
	MouseButton2 MouseButton = 2 // middle
	MouseButton3 MouseButton = 3 // right
)
