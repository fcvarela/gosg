package opengl

import (
	"reflect"
	"unsafe"

	"github.com/fcvarela/gosg/core"

	"github.com/go-gl/gl/v3.3-core/gl"
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

// NewUniform implements the core.RenderSystem interface
func (r *RenderSystem) NewUniform() core.Uniform {
	return &Uniform{nil}
}

// NewUniformBuffer implements the core.RenderSystem interface
func (r *RenderSystem) NewUniformBuffer() core.UniformBuffer {
	ub := &UniformBuffer{}
	gl.GenBuffers(1, &ub.id)
	return ub
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
	gl.BindBuffer(gl.UNIFORM_BUFFER, ub.id)
	gl.BufferData(gl.UNIFORM_BUFFER, dataLen, nil, gl.DYNAMIC_DRAW)
	gl.BufferData(gl.UNIFORM_BUFFER, dataLen, data, gl.DYNAMIC_DRAW)
}

// Lt implements the core.UniformBuffer interface
func (ub *UniformBuffer) Lt(other core.UniformBuffer) bool {
	return ub.id < other.(*UniformBuffer).id
}

// Gt implements the core.UniformBuffer interface
func (ub *UniformBuffer) Gt(other core.UniformBuffer) bool {
	return ub.id > other.(*UniformBuffer).id
}
