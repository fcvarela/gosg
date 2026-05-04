package core

import (
	"github.com/go-gl/mathgl/mgl64"
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
	pipeline *Pipeline
	material *Material
	sortKey  uint64

	// geometry, lighting & physics
	mesh      *Mesh
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
	return a[i].sortKey < a[j].sortKey
}

// NodesByCameraDistanceNearToFar is used to sort nodes according to camera distance from near to far.
type NodesByCameraDistanceNearToFar struct {
	Nodes      []*Node
	RefNode    *Node
	refPos     mgl64.Vec3
	computed   bool
	nodePos    []mgl64.Vec3
	nodeDist   []float64
}

// Len implements the sort.Interface interface.
func (a NodesByCameraDistanceNearToFar) Len() int {
	return len(a.Nodes)
}

// Swap implements the sort.Interface interface.
func (a NodesByCameraDistanceNearToFar) Swap(i, j int) {
	a.Nodes[i], a.Nodes[j] = a.Nodes[j], a.Nodes[i]
	a.nodePos[i], a.nodePos[j] = a.nodePos[j], a.nodePos[i]
	a.nodeDist[i], a.nodeDist[j] = a.nodeDist[j], a.nodeDist[i]
}

// Less implements the sort.Interface interface.
func (a NodesByCameraDistanceNearToFar) Less(i, j int) bool {
	if !a.computed {
		a.refPos = a.RefNode.WorldPosition()
		a.nodePos = make([]mgl64.Vec3, len(a.Nodes))
		a.nodeDist = make([]float64, len(a.Nodes))
		for i, n := range a.Nodes {
			a.nodePos[i] = n.WorldPosition()
			a.nodeDist[i] = a.nodePos[i].Sub(a.refPos).Len()
		}
		a.computed = true
	}
	return a.nodeDist[i] < a.nodeDist[j]
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

// NewNode returns a new node named `name`.
func NewNode(name string) *Node {
	mat := NewMaterial()
	n := Node{
		name:           name,
		transform:      mgl64.Ident4(),
		worldTransform: mgl64.Ident4(),
		active:         true,
		bounds:         NewAABB(),
		worldBounds:    NewAABB(),
		material:       &mat,
		children:       make([]*Node, 0),
		dirtyBounds:    true,
		dirtyTransform: true,

		lightExtractor:   new(DefaultLightExtractor),
		cullComponent:    new(DefaultCuller),
		physicsComponent: new(DefaultPhysicsComponent),
	}
	n.inverseWorldTransform = n.transform.Inv()

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
func (n *Node) Mesh() *Mesh {
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
func (n *Node) SetMesh(m *Mesh) {
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

// SetPipeline sets the node's pipeline
func (n *Node) SetPipeline(s *Pipeline) {
	n.pipeline = s
}

// Pipeline returns the node's pipeline
func (n *Node) Pipeline() *Pipeline {
	return n.pipeline
}

// Material returns the node's material.
func (n *Node) Material() *Material {
	return n.material
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
	n.setDirtyBounds()
}

// AddChild adds a child to the node
func (n *Node) AddChild(c *Node) {
	n.children = append(n.children, c)
	c.parent = n
	n.setDirtyBounds()
}

// RemoveChild removes a node's child. Returns true if the child was found and removed.
func (n *Node) RemoveChild(c *Node) bool {
	for i := range n.children {
		if n.children[i] == c {
			n.setDirtyBounds()
			n.children = append(n.children[:i], n.children[i+1:]...)
			c.parent = nil
			return true
		}
	}
	return false
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
	mat := NewMaterial()
	nc := Node{
		name:           n.name,
		active:         n.active,
		transform:      n.transform,
		worldTransform: n.worldTransform,
		inverseWorldTransform: n.inverseWorldTransform,
		pipeline:       n.pipeline,
		material:       &mat,
		mesh:           n.mesh,
		light:          n.light,
		rigidBody:      n.rigidBody,
		lightExtractor: n.lightExtractor,
		updateComponent: n.updateComponent,
		cullComponent:  n.cullComponent,
		inputComponent: n.inputComponent,
		physicsComponent: n.physicsComponent,
		boundsCallback: n.boundsCallback,
		bounds:         NewAABB(),
		worldBounds:    NewAABB(),
		dirtyTransform: true,
		dirtyBounds:    true,
		children:       make([]*Node, 0),
	}

	// deep copy material data
	for k, v := range n.material.uniforms {
		nc.material.uniforms[k] = v.Copy()
	}
	for k, v := range n.material.textures {
		nc.material.SetTexture(k, v)
	}
	nc.material.instanceData = n.material.instanceData

	for _, c := range n.children {
		nc.AddChild(c.Copy())
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
