package core

import (
	"math"
	"runtime"
	"sort"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/golang/glog"
)

// Scene represents a scenegraph and contains information about how it should be composed with other scenes on
// a scene stack. Scenes are not meant to be wrapped by users, but to be data configured for the expected behaviour.
type Scene struct {
	name string
	root *Node

	// should the scenemanager call update and draw this scene
	active bool

	cameraList []*Camera
	cameraMap  map[string]int

	// per scene lights list
	lights []*Light
}

func deleteScene(s *Scene) {
	glog.Info("Scene finalizer started: ", s.name)

	s.root.RemoveChildren()
	s.root = nil

	glog.Info("Scene finalizer finished: ", s.name)
}

// NewScene returns a new scene.
func NewScene(name string) *Scene {
	s := Scene{}

	s.name = name
	s.active = true
	s.cameraList = make([]*Camera, 0)
	s.cameraMap = make(map[string]int)
	s.lights = make([]*Light, 0)

	runtime.SetFinalizer(&s, deleteScene)

	return &s
}

// Root returns the scene's root node
func (s *Scene) Root() *Node {
	return s.root
}

// Cameras returns the scene's cameras
func (s *Scene) Cameras() []*Camera {
	return s.cameraList
}

// Camera returns a scene camera by name
func (s *Scene) Camera(name string) *Camera {
	return s.cameraList[s.cameraMap[name]]
}

// SetRoot returns the scene's root node
func (s *Scene) SetRoot(root *Node) {
	s.root = root
}

// Name returns the scene's name
func (s *Scene) Name() string {
	return s.name
}

// SetActive sets the 'active' state of this scene
func (s *Scene) SetActive(active bool) {
	s.active = active
}

// Active returns whether this scene is active or not
func (s *Scene) Active() bool {
	return s.active
}

// AddCamera adds a camera to the scene by attaching it to the given node.
func (s *Scene) AddCamera(node *Node, camera *Camera) {
	node.AddChild(camera.node)

	s.cameraList = append(s.cameraList, camera)
	s.cameraMap[camera.name] = len(s.cameraList) - 1

	// resort camera list by renderorder
	if len(s.cameraList) > 1 {
		sort.Sort(CamerasByRenderOrder(s.cameraList))
	}
}

func (s *Scene) update(dt float64) {
	// physics update
	var physicsNodes []*Node
	if s.root.physicsComponent != nil {
		s.root.physicsComponent.Run(s.root, &physicsNodes)
	}

	physicsSystem.Update(dt, physicsNodes)

	// update transforms and bounds
	s.root.update(s, dt)
}

func (s *Scene) cull() {
	for _, c := range s.cameraList {
		if c.autoFrustum {
			// Compute clip distances from scene bounds in view space,
			// not Euclidean distance, so they stay correct as the camera rotates.
			viewBounds := c.scene.worldBounds.Transformed(c.viewMatrix)
			// In view space, camera looks down -Z: closest objects have max Z, farthest have min Z
			near := math.Max(1.0, -viewBounds.max[2])
			far := -viewBounds.min[2]
			if far <= near {
				far = near + 1.0
			}
			c.SetClipDistance(mgl64.Vec2{near, far})
		}
		c.Reshape(windowManager.WindowSize())
	}

	s.lights = s.lights[:0]
	if s.root.lightExtractor != nil {
		s.root.lightExtractor.Run(s.root, &s.lights)
	}

	for _, c := range s.cameraList {
		for bk := range c.pipelineBuckets {
			c.pipelineBuckets[bk] = c.pipelineBuckets[bk][:0]
			c.visibleOpaqueNodes = c.visibleOpaqueNodes[:0]
		}

		c.scene.CullComponent().Run(s, c, c.scene)

		for bk := range c.pipelineBuckets {
			for _, n := range c.pipelineBuckets[bk] {
				var meshID uint32
				if n.mesh != nil {
					meshID = n.mesh.id
				}
				n.sortKey = (n.material.sortKey << 32) | uint64(meshID)
			}
			sort.Sort(NodesByMaterial(c.pipelineBuckets[bk]))
		}
		sort.Sort(NodesByCameraDistanceNearToFar{c.visibleOpaqueNodes, c.node})
	}
}

func (s *Scene) draw() {
	for _, camera := range s.cameraList {
		if camera.projectionType == PerspectiveProjection {
			for _, light := range s.lights {
				if light.Shadower != nil {
					light.Shadower.Render(light, camera)
				}
			}
		}

		camera.constants.SetData(camera.ProjectionMatrix(), camera.ViewMatrix(), s.lights)
		camera.renderTechnique(camera, camera.pipelineBuckets)
	}
}
