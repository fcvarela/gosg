package core

import (
	"runtime"

	"github.com/fcvarela/gosg/protos"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/golang/glog"
)

// NodeCommand is an interface which wraps logic for running a command against a node.
type NodeCommand interface {
	Run(node *Node)
}

// BoundsCallbackFn is called if set to allow the user to specify logic to customise
// a node's bounds. The bounds will have been grown to include any mesh or children. It's/
// up to the user to decide whether to return the same bounds or a new AABB.
type BoundsCallbackFn func(n *Node) *AABB

// Node represents a scenegraph node.
type Node struct {
	name string

	// graph stuff
	active   bool
	children []*Node
	parent   *Node

	// transform and bounds in object space
	transform      mgl64.Mat4
	bounds         *AABB
	boundsCallback BoundsCallbackFn

	// flags for transform and bounds update
	dirtyTransform bool
	dirtyBounds    bool

	// same in world space. never used here but other components
	// shouldn't have to compute when needed.
	worldTransform        mgl64.Mat4
	inverseWorldTransform mgl64.Mat4
	worldBounds           *AABB

	// state management
	state        *protos.State
	materialData Descriptors

	// geometry, lighting & physics
	mesh      Mesh
	light     *Light
	rigidBody RigidBody

	// possibly custom stuff
	lightExtractor   LightExtractor
	inputComponent   InputComponent
	updateComponent  Updater
	cullComponent    Culler
	physicsComponent PhysicsComponent
}

// NodesByMaterial is used to sort nodes according to material.
type NodesByMaterial []*Node

// Len implements the sort.Interface interface
func (a NodesByMaterial) Len() int {
	return len(a)
}

// Swap implements the sort.Interface interface.
func (a NodesByMaterial) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// Less implements the sort.Interface interface.
func (a NodesByMaterial) Less(i, j int) bool {
	// we want to sort by: program, textures, bufferobjects
	// this maps to: node.materialName, node.materialData.textures, node.materialData.bufferObjects

	// sort by materialName first
	if a[i].state.Name < a[j].state.Name {
		return true
	} else if a[i].state.Name > a[j].state.Name {
		return false
	}

	// if we got here they share the same material, check program
	if a[i].state.ProgramName < a[j].state.ProgramName {
		return true
	} else if a[i].state.ProgramName > a[j].state.ProgramName {
		return false
	}

	// if we got here, they share the same program hence use the same sampler names, look for those textures
	for textureName := range a[i].materialData.textures {
		textureI, textureJ := a[i].materialData.textures[textureName], a[j].materialData.textures[textureName]
		if textureI.Lt(textureJ) {
			return true
		} else if textureI.Gt(textureJ) {
			return false
		}
	}

	// if we got here, they use the same textures for the same samplers. Check mesh. Do not call this with any nodes
	// not containing meshes
	meshI, meshJ := a[i].mesh, a[j].mesh
	if meshI.Lt(meshJ) {
		return true
	} else if meshI.Gt(meshJ) {
		return false
	}

	/*
		// if we got here, they use the same textures for the same samplers. Check uniform buffers
		for uniformBufferName := range a[i].materialData.uniformBuffers {
			uniformBufferI, uniformBufferJ := a[i].materialData.uniformBuffers[uniformBufferName], a[j].materialData.uniformBuffers[uniformBufferName]
			if uniformBufferI.Lt(uniformBufferJ) {
				return true
			} else if uniformBufferI.Gt(uniformBufferJ) {
				return false
			}
		}
	*/

	// if we got here, they share everything, return false (not less :))
	return false
}

// NodesByCameraDistanceNearToFar is used to sort nodes according to camera distance from near to far.
type NodesByCameraDistanceNearToFar struct {
	Nodes   []*Node
	RefNode *Node
}

// Len implements the sort.Interface interface.
func (a NodesByCameraDistanceNearToFar) Len() int {
	return len(a.Nodes)
}

// Swap implements the sort.Interface interface.
func (a NodesByCameraDistanceNearToFar) Swap(i, j int) {
	a.Nodes[i], a.Nodes[j] = a.Nodes[j], a.Nodes[i]
}

// Less implements the sort.Interface interface.
func (a NodesByCameraDistanceNearToFar) Less(i, j int) bool {
	pos1 := a.Nodes[i].WorldPosition()
	pos2 := a.Nodes[j].WorldPosition()

	dist1 := pos1.Sub(a.RefNode.WorldPosition()).Len()
	dist2 := pos2.Sub(a.RefNode.WorldPosition()).Len()

	return dist1 < dist2
}

// NodesByName is used to sort nodes by alphabetic name order.
type NodesByName struct {
	Nodes []*Node
}

// Len implements the sort.Interface interface.
func (a NodesByName) Len() int {
	return len(a.Nodes)
}

// Swap implements the sort.Interface interface.
func (a NodesByName) Swap(i, j int) {
	a.Nodes[i], a.Nodes[j] = a.Nodes[j], a.Nodes[i]
}

// Less implements the sort.Interface interface.
func (a NodesByName) Less(i, j int) bool {
	return a.Nodes[i].name < a.Nodes[j].name
}

func deleteNode(n *Node) {
	glog.Info("Node cleaning up: ", n.name)
}

// NewNode returns a new node named `name`.
func NewNode(name string) *Node {
	n := Node{}
	n.name = name
	n.transform = mgl64.Ident4()
	n.worldTransform = n.transform
	n.inverseWorldTransform = n.transform.Inv()
	n.active = true
	n.bounds = NewAABB()
	n.worldBounds = NewAABB()
	n.state = nil
	n.materialData = NewDescriptors()
	n.children = make([]*Node, 0)
	n.dirtyBounds = true
	n.dirtyTransform = true

	// user should override
	n.lightExtractor = new(DefaultLightExtractor)
	n.updateComponent = nil
	n.cullComponent = new(DefaultCuller)
	n.inputComponent = nil
	n.physicsComponent = new(DefaultPhysicsComponent)

	runtime.SetFinalizer(&n, deleteNode)

	return &n
}

// Name returns the node's name.
func (n *Node) Name() string {
	return n.name
}

// SetActive marks the node as active.
func (n *Node) SetActive(active bool) {
	n.active = active
}

// Active returns whether the node is active.
func (n *Node) Active() bool {
	return n.active
}

// SetChildrenActive marks all of the nodes's children as active.
func (n *Node) SetChildrenActive(active bool) {
	for _, c := range n.Children() {
		c.SetActive(active)
	}
}

// InverseWorldTransform returns the node's inverse world transform.
func (n *Node) InverseWorldTransform() mgl64.Mat4 {
	return n.inverseWorldTransform
}

// SetBoundsCallback sets a boundsCallback function
func (n *Node) SetBoundsCallback(f BoundsCallbackFn) {
	n.boundsCallback = f
}

// Mesh returns the node's mesh.
func (n *Node) Mesh() Mesh {
	return n.mesh
}

// RigidBody returns the node's rigid body
func (n *Node) RigidBody() RigidBody {
	return n.rigidBody
}

// Children returns the node's children.
func (n *Node) Children() []*Node {
	return n.children
}

// SetMesh sets the node's mesh.
func (n *Node) SetMesh(m Mesh) {
	n.mesh = m
	n.setDirtyBounds()
}

// SetLight set's the node's light
func (n *Node) SetLight(l *Light) {
	n.light = l
	n.setDirtyBounds()
}

// Light returns the node's light.
func (n *Node) Light() *Light {
	return n.light
}

// SetRigidBody sets the node's rigid body.
func (n *Node) SetRigidBody(r RigidBody) {
	n.rigidBody = r
}

func (n *Node) setDirtyBounds() {
	n.dirtyBounds = true
	if n.parent != nil {
		n.parent.setDirtyBounds()
	}
}

func (n *Node) setDirtyTransform() {
	n.dirtyTransform = true
	if n.parent != nil {
		n.parent.setDirtyBounds()
	}
	for _, c := range n.children {
		c.setDirtyTransform()
	}
}

// SetState sets the node's pipeline state
func (n *Node) SetState(s *protos.State) {
	n.state = s
}

// State returns the node's state
func (n *Node) State() *protos.State {
	return n.state
}

// MaterialData returns the node's state
func (n *Node) MaterialData() *Descriptors {
	return &n.materialData
}

// Transform returns the node's transform
func (n *Node) Transform() mgl64.Mat4 {
	return n.transform
}

// SetWorldTransform sets the node's world transform. It also sets the node's transform appropriately.
func (n *Node) SetWorldTransform(transform mgl64.Mat4) {
	if n.parent != nil {
		n.transform = n.parent.inverseWorldTransform.Mul4(transform)
	} else {
		n.transform = transform
		n.worldTransform = transform
	}
	n.setDirtyTransform()
	n.setDirtyBounds()
}

// WorldTransform returns the node's world transform.
func (n *Node) WorldTransform() mgl64.Mat4 {
	return n.worldTransform
}

func (n *Node) update(s *Scene, dt float64) {
	// do we have an input component
	if n.inputComponent != nil {
		cmds := n.inputComponent.Run(n)
		for _, c := range cmds {
			c.Run(n)
		}
	}

	if n.updateComponent != nil {
		n.updateComponent.Run(n)
	}

	// update our transforms
	if n.dirtyTransform {
		n.updateTransforms()
	}

	// recurse
	for _, c := range n.children {
		c.update(s, dt)
	}

	// are our bounds dirty?
	if n.dirtyBounds {
		n.updateBounds()
	}
}

// Parent returns the node's parent.
func (n *Node) Parent() *Node {
	return n.parent
}

// SetCullComponent sets the node's culler.
func (n *Node) SetCullComponent(cc Culler) {
	n.cullComponent = cc
}

// SetInputComponent sets the node's input component.
func (n *Node) SetInputComponent(ic InputComponent) {
	n.inputComponent = ic
}

// SetUpdateComponent sets the node's update component.
func (n *Node) SetUpdateComponent(uc Updater) {
	n.updateComponent = uc
}

// CullComponent returns the node's culler
func (n *Node) CullComponent() Culler {
	return n.cullComponent
}

// UpdateComponent returns the node's update component.
func (n *Node) UpdateComponent() Updater {
	return n.updateComponent
}

// InputComponent returns the node's input component.
func (n *Node) InputComponent() InputComponent {
	return n.inputComponent
}

// PhysicsComponent returns the node's physics component.
func (n *Node) PhysicsComponent() PhysicsComponent {
	return n.physicsComponent
}

// WE NEED SEPARATE LOCAL AND WORLD BOUNDS
func (n *Node) updateBounds() {
	if n.bounds == nil || n.light != nil {
		return
	}

	// reset
	n.bounds.Reset()
	n.worldBounds.Reset()

	// add our mesh
	if n.mesh != nil {
		n.bounds.ExtendWithBox(n.mesh.Bounds())
	} else {
		n.bounds.ExtendWithPoint(mgl64.Vec3{0.0, 0.0, 0.0})
	}

	// and our children bounds
	for _, c := range n.children {
		// fixme: we need light components, transform components and bounds components
		if c.bounds != nil && c.light == nil {
			n.bounds.ExtendWithBox(c.bounds.Transformed(c.transform))
		}
	}

	if n.boundsCallback != nil {
		n.worldBounds = n.boundsCallback(n)
		n.bounds = n.worldBounds.Transformed(n.inverseWorldTransform)
		if n.parent != nil {
			n.parent.setDirtyBounds()
		}
	} else {
		// transform bounds w/ worldtransform
		n.worldBounds = n.bounds.Transformed(n.worldTransform)
	}
	n.dirtyBounds = false
}

func (n *Node) updateTransforms() {
	if n.parent != nil {
		n.worldTransform = n.parent.worldTransform.Mul4(n.transform)
	} else {
		n.worldTransform = n.transform
	}
	n.inverseWorldTransform = n.worldTransform.Inv()
	n.dirtyTransform = false
}

// Rotate rotates the node by `eulerAngle` degrees around `axis`.
func (n *Node) Rotate(eulerAngle float64, axis mgl64.Vec3) {
	rotationMatrix := mgl64.QuatRotate(mgl64.DegToRad(eulerAngle), axis).Normalize().Mat4()
	n.transform = n.transform.Mul4(rotationMatrix)
	n.setDirtyTransform()
	n.setDirtyBounds()
}

// Translate translates a node.
func (n *Node) Translate(vec mgl64.Vec3) {
	n.transform = n.transform.Mul4(mgl64.Translate3D(vec.X(), vec.Y(), vec.Z()))
	n.setDirtyTransform()
	n.setDirtyBounds()
}

// Scale scales a node.
func (n *Node) Scale(s mgl64.Vec3) {
	n.transform = n.transform.Mul4(mgl64.Scale3D(s[0], s[1], s[2]))
	n.setDirtyTransform()
	n.updateBounds()
}

// AddChild adds a child to the node
func (n *Node) AddChild(c *Node) {
	n.children = append(n.children, c)
	c.parent = n
	n.setDirtyBounds()
}

// RemoveChild removes a node's child.
func (n *Node) RemoveChild(c *Node) {
	for i := range n.children {
		if n.children[i] == c {
			c.parent.setDirtyBounds()
			n.children = append(n.children[:i], n.children[i+1:]...)
			c.parent = nil
		}
	}
}

// RemoveChildren removes all of a node's children.
func (n *Node) RemoveChildren() {
	for _, c := range n.children {
		c.parent = nil
		c.mesh = nil
		c.inputComponent = nil
		c.cullComponent = nil
		c.physicsComponent = nil
		c.RemoveChildren()
	}
	n.children = make([]*Node, 0)
}

// Copy deep copies a node.
func (n *Node) Copy() *Node {
	nc := *n
	nc.children = make([]*Node, 0)

	for _, c := range n.children {
		cc := *c
		cc.materialData.uniforms = make(map[string]Uniform)
		for k, v := range c.materialData.uniforms {
			cc.materialData.Uniform(k).Set(v.Value())
		}
		for k, v := range c.materialData.textures {
			cc.materialData.SetTexture(k, v)
		}
		nc.AddChild(&cc)
	}

	return &nc
}

// WorldPosition returns the node's world position
func (n *Node) WorldPosition() mgl64.Vec3 {
	return mgl64.Vec3{n.worldTransform[12], n.worldTransform[13], n.worldTransform[14]}
}

// Bounds returns the node's bounds in object-space
func (n *Node) Bounds() *AABB {
	return n.bounds
}

// WorldBounds returns the node's bounds in world-space
func (n *Node) WorldBounds() *AABB {
	return n.worldBounds
}

// WorldDistance returns the distance between this node and the passed node, computed
// using node centerpoints/world position
func (n *Node) WorldDistance(n2 *Node) float64 {
	return n2.WorldPosition().Sub(n.WorldPosition()).Len()
}
