// Package bullet implements the core.PhysicsSystem interface by wrapping the Bullet physics library.
package bullet

// #cgo pkg-config: bullet
// #cgo windows LDFLAGS: -Wl,--allow-multiple-definition
// #include "bulletglue.h"
import "C"
import (
	"github.com/fcvarela/gosg/core"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/golang/glog"
)

func init() {
	core.SetPhysicsSystem(&PhysicsSystem{})
}

// convenience
func vec3ToBullet(vec mgl64.Vec3) (out C.plVector3) {
	out[0] = C.plReal(vec.X())
	out[1] = C.plReal(vec.Y())
	out[2] = C.plReal(vec.Z())

	return out
}

func quatToBullet(quat mgl64.Quat) (out C.plQuaternion) {
	out[0] = C.plReal(quat.X())
	out[1] = C.plReal(quat.Y())
	out[2] = C.plReal(quat.Z())
	out[3] = C.plReal(quat.W)

	return out
}

func mat4ToBullet(mat mgl64.Mat4) (out [16]C.plReal) {
	for x := 0; x < 16; x++ {
		out[x] = C.plReal(mat[x])
	}
	return out
}

func mat4FromBullet(mat [16]C.plReal) (out mgl64.Mat4) {
	for x := 0; x < 16; x++ {
		out[x] = float64(mat[x])
	}
	return out
}

// PhysicsSystem implements the core.PhysicsSystem interface by wrapping the Bullet physics library.
type PhysicsSystem struct {
	sdk   C.plPhysicsSdkHandle
	world C.plDynamicsWorldHandle
}

// Start implements the core.PhysicsSystem interface
func (p *PhysicsSystem) Start() {
	glog.Info("Starting")

	// create an sdk handle
	p.sdk = C.plNewBulletSdk()

	// instance a world
	p.world = C.plCreateDynamicsWorld(p.sdk)
	C.plSetGravity(p.world, 0.0, 0.0, 0.0)
}

// Stop implements the core.PhysicsSystem interface
func (p *PhysicsSystem) Stop() {
	glog.Info("Stopping")

	C.plDeleteDynamicsWorld(p.world)
	C.plDeletePhysicsSdk(p.sdk)
}

// SetGravity implements the core.PhysicsSystem interface
func (p *PhysicsSystem) SetGravity(g mgl64.Vec3) {
	vec := vec3ToBullet(g)
	C.plSetGravity(p.world, vec[0], vec[1], vec[2])
}

// Update implements the core.PhysicsSystem interface
// fixme: remove gosg dependencies by passing a RigidBodyVec instead of NodeVec
func (p *PhysicsSystem) Update(dt float64, nodes []*core.Node) {
	for _, n := range nodes {
		n.RigidBody().SetTransform(n.WorldTransform())
	}
	C.plStepSimulation(p.world, C.plReal(dt))
	for _, n := range nodes {
		n.SetWorldTransform(n.RigidBody().GetTransform())
	}
}

// AddRigidBody implements the core.PhysicsSystem interface
func (p *PhysicsSystem) AddRigidBody(rigidBody core.RigidBody) {
	C.plAddRigidBody(p.world, rigidBody.(RigidBody).handle)
}

// RemoveRigidBody implements the core.PhysicsSystem interface
func (p *PhysicsSystem) RemoveRigidBody(rigidBody core.RigidBody) {
	C.plRemoveRigidBody(p.world, rigidBody.(RigidBody).handle)
}

// CreateRigidBody implements the core.PhysicsSystem interface
func (p *PhysicsSystem) CreateRigidBody(mass float32, shape core.CollisionShape) core.RigidBody {
	body := C.plCreateRigidBody(nil, C.float(mass), shape.(CollisionShape).handle)
	r := RigidBody{body}
	return r
}

// DeleteRigidBody implements the core.PhysicsSystem interface
func (p *PhysicsSystem) DeleteRigidBody(body core.RigidBody) {
	C.plDeleteRigidBody(body.(RigidBody).handle)
}

// NewStaticPlaneShape implements the core.PhysicsSystem interface
func (p *PhysicsSystem) NewStaticPlaneShape(normal mgl64.Vec3, constant float64) core.CollisionShape {
	vec := vec3ToBullet(normal)
	return CollisionShape{C.plNewStaticPlaneShape(&vec[0], C.float(constant))}
}

// NewSphereShape implements the core.PhysicsSystem interface
func (p *PhysicsSystem) NewSphereShape(radius float64) core.CollisionShape {
	return CollisionShape{C.plNewSphereShape(C.plReal(radius))}
}

// NewBoxShape implements the core.PhysicsSystem interface
func (p *PhysicsSystem) NewBoxShape(box mgl64.Vec3) core.CollisionShape {
	vec := vec3ToBullet(box)
	return CollisionShape{C.plNewBoxShape(vec[0], vec[1], vec[2])}
}

// NewCapsuleShape implements the core.PhysicsSystem interface
func (p *PhysicsSystem) NewCapsuleShape(radius float64, height float64) core.CollisionShape {
	return CollisionShape{C.plNewCapsuleShape(C.plReal(radius), C.plReal(height))}
}

// NewConeShape implements the core.PhysicsSystem interface
func (p *PhysicsSystem) NewConeShape(radius float64, height float64) core.CollisionShape {
	return CollisionShape{C.plNewConeShape(C.plReal(radius), C.plReal(height))}
}

// NewCylinderShape implements the core.PhysicsSystem interface
func (p *PhysicsSystem) NewCylinderShape(radius float64, height float64) core.CollisionShape {
	return CollisionShape{C.plNewCylinderShape(C.plReal(radius), C.plReal(height))}
}

// NewCompoundShape implements the core.PhysicsSystem interface
func (p *PhysicsSystem) NewCompoundShape() core.CollisionShape {
	return CollisionShape{C.plNewCompoundShape()}
}

// NewConvexHullShape implements the core.PhysicsSystem interface
func (p *PhysicsSystem) NewConvexHullShape() core.CollisionShape {
	return CollisionShape{C.plNewConvexHullShape()}
}

// NewStaticTriangleMeshShape implements the core.PhysicsSystem interface
func (p *PhysicsSystem) NewStaticTriangleMeshShape(mesh core.Mesh) core.CollisionShape {
	/*
		bulletMeshInterface := C.plNewMeshInterface()

		// add triangles
		for v := 0; v < len(indices); v += 3 {
			i1 := indices[v+0]
			i2 := indices[v+1]
			i3 := indices[v+2]

			v1 := vec3_to_bullet(positions[i1*3])
			v2 := vec3_to_bullet(positions[i2*3])
			v3 := vec3_to_bullet(positions[i3*3])

			C.plAddTriangle(bulletMeshInterface, &v1[0], &v2[0], &v3[0])
		}

		return CollisionShape{C.plNewStaticTriangleMeshShape(bulletMeshInterface)}
	*/
	return nil
}

// DeleteShape implements the core.PhysicsSystem interface
func (p *PhysicsSystem) DeleteShape(shape core.CollisionShape) {
	C.plDeleteShape(shape.(CollisionShape).handle)
}

// RigidBody implements the core.RigidBody interface
type RigidBody struct {
	handle C.plRigidBodyHandle
}

// GetTransform implements the core.RigidBody interface
func (r RigidBody) GetTransform() mgl64.Mat4 {
	mat := mat4ToBullet(mgl64.Ident4())
	C.plGetOpenGLMatrix(r.handle, &mat[0])
	return mat4FromBullet(mat)
}

// SetTransform implements the core.RigidBody interface
func (r RigidBody) SetTransform(transform mgl64.Mat4) {
	mat := mat4ToBullet(transform)
	C.plSetOpenGLMatrix(r.handle, &mat[0])
}

// ApplyImpulse implements the core.RigidBody interface
func (r RigidBody) ApplyImpulse(impulse mgl64.Vec3, localPoint mgl64.Vec3) {
	i := vec3ToBullet(impulse)
	p := vec3ToBullet(localPoint)
	C.plApplyImpulse(r.handle, &i[0], &p[0])
}

// CollisionShape implements the core.CollisionShape interface
type CollisionShape struct {
	handle C.plCollisionShapeHandle
}

// AddChildShape implements the core.CollisionShape interface
func (c CollisionShape) AddChildShape(s core.CollisionShape, p mgl64.Vec3, o mgl64.Quat) {
	vec := vec3ToBullet(p)
	quat := quatToBullet(o)
	C.plAddChildShape(c.handle, s.(CollisionShape).handle, &vec[0], &quat[0])
}

// AddVertex implements the core.CollisionShape interface
func (c CollisionShape) AddVertex(v mgl64.Vec3) {
	C.plAddVertex(c.handle, C.plReal(v.X()), C.plReal(v.Y()), C.plReal(v.Z()))
}
