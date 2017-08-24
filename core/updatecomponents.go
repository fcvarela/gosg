package core

// Updater is an interface that wraps updating a node.
type Updater interface {
	// Run updates a scenegraph node.
	Run(*Node)
}
