package core

// Culler is an interface that wraps culling of a scenegraph.
type Culler interface {
	// Run culls a scenegraph node. A scene and camera are provided for visibility/frustum checks.
	// If the policy dictates the node is to be drawn, then it should be added to the nodeBucket.
	Run(*Scene, *Camera, *Node)
}

// DefaultCuller implements a scenegraph culler. The policy for this culler is to
// mark all nodes in frustum for drawing. The node's modelMatrix state uniform is also set
// from the nodes worldtransform. This may change as we transition away from individual uniforms
// for instanced/indirect drawing.
type DefaultCuller struct{}

// Run implements the CullComponent interface
func (cc *DefaultCuller) Run(scene *Scene, camera *Camera, node *Node) {
	if !node.worldBounds.InFrustum(camera.frustum) {
		return
	}

	if !node.active {
		return
	}

	// the default implementation is to add ourselves to the bucket
	if node.mesh != nil {
		camera.stateBuckets[node.state] = append(camera.stateBuckets[node.state], node)
		if !node.state.Blending {
			camera.visibleOpaqueNodes = append(camera.visibleOpaqueNodes, node)
		}
	}

	for _, c := range node.children {
		c.cullComponent.Run(scene, camera, c)
	}
}

// AlwaysPassCuller implements a scenegraph culler by always adding the node to the bucket
type AlwaysPassCuller struct{}

// Run implements the Culler interface
func (apcc *AlwaysPassCuller) Run(scene *Scene, camera *Camera, node *Node) {
	// the default implementation is to add ourselves to the bucket
	if node.mesh != nil {
		camera.stateBuckets[node.state] = append(camera.stateBuckets[node.state], node)
	}

	for _, ch := range node.children {
		ch.cullComponent.Run(scene, camera, ch)
	}
}
