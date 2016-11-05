package opengl

import (
	"reflect"
	"unsafe"

	"github.com/fcvarela/gosg/core"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/golang/glog"
)

// Uniform implements the core.Uniform interface
type Uniform struct {
	value interface{}
}

// UniformBuffer implements the core.UniformBuffer interface
type UniformBuffer struct {
	id uint32
}

type makeUniformBufferCommand struct {
	ub core.UniformBuffer
}

// NewUniform implements the core.RenderSystem interface
func (r *RenderSystem) NewUniform() core.Uniform {
	return &Uniform{nil}
}

func (r *RenderSystem) makeUniformBuffer(cmd *makeUniformBufferCommand) error {
	ub := &UniformBuffer{}
	gl.GenBuffers(1, &ub.id)
	cmd.ub = ub
	return nil
}

// NewUniformBuffer implements the core.RenderSystem interface
func (r *RenderSystem) NewUniformBuffer() core.UniformBuffer {
	var cmd = &makeUniformBufferCommand{}
	if err := r.Run(cmd, true); err != nil {
		glog.Fatal(err)
	}

	return cmd.ub
}

// Set implements the core.Uniform interface
func (u *Uniform) Set(value interface{}) {
	u.value = value
}

// Value implements the core.Uniform interface
func (u *Uniform) Value() interface{} {
	return u.value
}

// Copy returns a copy of the uniform.
func (u *Uniform) Copy() core.Uniform {
	switch uval := u.Value().(type) {
	case mgl32.Mat4:
		return &Uniform{uval}
	case mgl64.Mat4:
		return &Uniform{uval}
	case mgl64.Vec4:
		return &Uniform{uval}
	case mgl64.Vec3:
		return &Uniform{uval}
	case mgl64.Vec2:
		return &Uniform{uval}
	case []float32:
		return &Uniform{uval}
	case []mgl32.Vec2:
		return &Uniform{uval}
	case int:
		return &Uniform{uval}
	case float32:
		return &Uniform{uval}
	case float64:
		return &Uniform{uval}
	default:
		glog.Fatalf("Unsupported uniform type: %s\n", reflect.TypeOf(u.Value()))
	}

	return nil
}

// Set implements the core.UniformBuffer interface
func (ub *UniformBuffer) Set(data unsafe.Pointer, dataLen int) {
	var cmd = &lambdaRenderCommand{
		fn: func() error {
			gl.BindBuffer(gl.UNIFORM_BUFFER, ub.id)
			gl.BufferData(gl.UNIFORM_BUFFER, dataLen, nil, gl.DYNAMIC_DRAW)
			gl.BufferData(gl.UNIFORM_BUFFER, dataLen, data, gl.DYNAMIC_DRAW)
			return nil
		},
	}
	renderSystem.Run(cmd, true)
}

// Lt implements the core.UniformBuffer interface
func (ub *UniformBuffer) Lt(other core.UniformBuffer) bool {
	return ub.id < other.(*UniformBuffer).id
}

// Gt implements the core.UniformBuffer interface
func (ub *UniformBuffer) Gt(other core.UniformBuffer) bool {
	return ub.id > other.(*UniformBuffer).id
}
