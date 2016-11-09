package core

// SceneManager manages a stack of scenes
type SceneManager struct {
	managedScenes []*Scene
}

var (
	sceneManager *SceneManager
)

func init() {
	sceneManager = &SceneManager{
		make([]*Scene, 0),
	}
}

// GetSceneManager returns the scene manager.
func GetSceneManager() *SceneManager {
	return sceneManager
}

// PushScene pushes a scene to the stack.
func (sm *SceneManager) PushScene(s *Scene) {
	sm.managedScenes = append(sm.managedScenes, s)
}

// PopScene pops a scene from the stack
func (sm *SceneManager) PopScene() *Scene {
	if len(sm.managedScenes) == 0 {
		return nil
	}

	previousFrontScene := sm.FrontScene()
	sm.managedScenes = sm.managedScenes[:len(sm.managedScenes)-1]

	return previousFrontScene
}

// FrontScene returns the front scene.
func (sm *SceneManager) FrontScene() *Scene {
	if sm.managedScenes == nil || len(sm.managedScenes) == 0 {
		return nil
	}

	return sm.managedScenes[len(sm.managedScenes)-1]
}

func (sm *SceneManager) update(dt float64) {
	// we update scenes in reverse order, frontmost processes input first
	for i := range sm.managedScenes {
		var currentScene = sm.managedScenes[len(sm.managedScenes)-1-i]
		if currentScene.active {
			currentScene.update(dt)
		}
	}
}

func (sm *SceneManager) cull() {
	for _, s := range sm.managedScenes {
		if s.active {
			s.cull()
		}
	}
}

func (sm *SceneManager) draw(rc chan RenderCommand) {
	for _, s := range sm.managedScenes {
		if s.active {
			s.draw(rc)
		}
	}
}
