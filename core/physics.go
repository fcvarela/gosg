package core

import (
	"log"

	"github.com/go-gl/mathgl/mgl64"
)

// PhysicsSystem is an interface which wraps all physics related logic.
type PhysicsSystem interface {
	// Start is called by the application at startup time. Implementations should perform bootstapping here.
	Start()

	// Stop is called by the application at shutdown time. Implementations should perform cleanup here.
	Stop()

	// Update is called at every cycle of the application runloop with a list of nodes and a time delta from the
	// previous iteration. Implementations will want to perform all their computation here.
	Update(dt float64, nodes []*Node)

	// SetGravity sets the global gravity vector. This is for testing purposes and will be removed.
	SetGravity(g mgl64.Vec3)

	// AddRigidBody adds a rigid body to the physics world.
	AddRigidBody(RigidBody)

	// RemoveRigidBody removed a rigid body from the physics world.
	RemoveRigidBody(RigidBody)

	// CreateRigidBody creates a new rigid body which can be attached to a scenegraph node.
	CreateRigidBody(mass float32, collisionShape CollisionShape) RigidBody

	// DeleteRigidBody deletes a rigid body
	DeleteRigidBody(RigidBody)

	// NewStaticPlaneShape returns a collision shape.
	NewStaticPlaneShape(normal mgl64.Vec3, constant float64) CollisionShape

	// NewSphereShape returns a collision shape.
	NewSphereShape(radius float64) CollisionShape

	// NewBoxShape returns a collision shape.
	NewBoxShape(mgl64.Vec3) CollisionShape

	// NewCapsuleShape returns a collision shape.
	NewCapsuleShape(radius float64, height float64) CollisionShape

	// NewConeShape returns a collision shape.
	NewConeShape(radius float64, height float64) CollisionShape

	// NewCylinderShape returns a collision shape.
	NewCylinderShape(radius float64, height float64) CollisionShape

	// NewCompoundSphereShape returns a collision shape.
	NewCompoundShape() CollisionShape

	// NewConvexHullShape returns a collision shape.
	NewConvexHullShape() CollisionShape

	// NewStaticTriangleMeshShape returns a collision shape.
	NewStaticTriangleMeshShape(Mesh) CollisionShape

	// DeleteShape deletes a collision shape.
	DeleteShape(CollisionShape)
}

// CollisionShape is an interface which wraps information used to compute object collisions.
type CollisionShape interface {
	// AddChildShape adds a child collision shape to this shape.
	AddChildShape(childshape CollisionShape, position mgl64.Vec3, orientation mgl64.Quat)

	// AddVertex adds a single vertex. Used for convex hull shapes.
	AddVertex(mgl64.Vec3)
}

// RigidBody is an interface which wraps a physics rigid body. It contains
// position, orientation, momentum and collision shape information.
type RigidBody interface {
	// GetTransform returns the rigid body world transform.
	GetTransform() mgl64.Mat4

	// SetTransform sets the rigid body world transform.
	SetTransform(mgl64.Mat4)

	// ApplyImpulse applies `impulse` on the rigid body at its position `localPosition`.
	ApplyImpulse(impulse mgl64.Vec3, localPosition mgl64.Vec3)
}

var (
	physicsSystem PhysicsSystem
)

// SetPhysicsSystem is meant to be called from PhysicsSystem implementations on their init method
func SetPhysicsSystem(ps PhysicsSystem) {
	if physicsSystem != nil {
		log.Fatal("Can't replace previously registered physics system. Please make sure you're not importing twice")
	}
	physicsSystem = ps
}

// GetPhysicsSystem returns the renderSystem, thereby exposing it to any package importing core.
func GetPhysicsSystem() PhysicsSystem {
	return physicsSystem
}

// PhysicsComponent is an interface which wraps physics handling logic for a scenegraph node
type PhysicsComponent interface {
	// Run is called on each node which should determine whether it should be added to the simulation step or not.
	Run(node *Node, nodeBucket *[]*Node)
}

// DefaultPhysicsComponent is a utility physics component which adds all nodes containing a rigid body
// to the bucket.
type DefaultPhysicsComponent struct{}

// Run implements the PhysicsComponent interface
func (p *DefaultPhysicsComponent) Run(n *Node, nodeBucket *[]*Node) {
	if n.rigidBody != nil {
		*nodeBucket = append(*nodeBucket, n)
	}

	for _, c := range n.children {
		c.physicsComponent.Run(c, nodeBucket)
	}
}

// NewDefaultPhysicsComponent returns a new DefaultPhysicsComponent.
func NewDefaultPhysicsComponent(active bool) *DefaultPhysicsComponent {
	pc := DefaultPhysicsComponent{}
	return &pc
}
