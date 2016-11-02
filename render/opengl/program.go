package opengl

import (
	"encoding/json"
	"reflect"
	"runtime"
	"strings"

	"github.com/fcvarela/gosg/core"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/golang/glog"
)

// Program is an OpenGL program
type Program struct {
	name                  string
	id                    uint32
	shaders               map[string]*shader
	compileLog            string
	uniformLog            string
	uniformLocations      map[string]int32
	uniformBlockIndexes   map[string]uint32
	uniformBufferBindings map[string]uint32
	samplerBindings       map[string]uint32
	dirtySamplerBindings  bool
}

// programSpec is the struct we get as bytes on calls to NewProgram
type programSpec struct {
	Shaders               map[string]string `json:"shaders"`
	UniformBufferBindings map[string]uint32 `json:"uniformBufferBindings"`
	SamplerBindings       map[string]uint32 `json:"samplerBindings"`
}

type makeProgramCommand struct {
	core.RenderCommand
	name    string
	data    []byte
	program *Program
}

func programCleanup(p *Program) {
	glog.Infof("Finalizer called for program: %v\n", p)
}

var (
	programTypeMap = map[string]uint32{
		"compute":            gl.COMPUTE_SHADER,
		"tesselationControl": gl.TESS_CONTROL_SHADER,
		"tesselationEval":    gl.TESS_EVALUATION_SHADER,
		"vertex":             gl.VERTEX_SHADER,
		"geometry":           gl.GEOMETRY_SHADER,
		"fragment":           gl.FRAGMENT_SHADER,
	}
	currentProgram *Program
)

// ProgramExtension implements the core.RenderSystem interface.
func (r *RenderSystem) ProgramExtension() string {
	return "gl.json"
}

func (r *RenderSystem) makeProgram(cmd *makeProgramCommand) error {
	var spec programSpec
	if err := json.Unmarshal(cmd.data, &spec); err != nil {
		glog.Fatal("Error reading program spec: ", err)
	}

	// create program
	prog := &Program{
		cmd.name,
		0,
		make(map[string]*shader),
		"",
		"",
		make(map[string]int32),
		make(map[string]uint32),
		make(map[string]uint32),
		make(map[string]uint32),
		false,
	}

	// set shaders
	for k, v := range spec.Shaders {
		source := core.GetResourceManager().ProgramData(v)
		prog.shaders[k] = newShader(v, programTypeMap[k], source)
	}

	// set ubo bindings
	for k, v := range spec.UniformBufferBindings {
		prog.uniformBufferBindings[k] = v
	}

	// set finalizer (hooks into gl to cleanup)
	runtime.SetFinalizer(prog, programCleanup)

	// compile it
	prog.id = gl.CreateProgram()

	for _, s := range prog.shaders {
		if s != nil {
			s.compile()
			gl.AttachShader(prog.id, s.id)
		}
	}

	gl.LinkProgram(prog.id)
	status := int32(0)
	gl.GetProgramiv(prog.id, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		logLength := int32(0)
		gl.GetProgramiv(prog.id, gl.INFO_LOG_LENGTH, &logLength)

		progLog := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(prog.id, logLength, nil, gl.Str(progLog))
		prog.compileLog = progLog

		glog.Fatalf("failed to link program: %v", progLog)
	}

	for _, s := range prog.shaders {
		if s != nil {
			gl.DeleteShader(s.id)
		}
	}

	// populate uniform name map
	prog.extractUniformNames()

	// populate uniform block index map
	prog.extractUniformBlockIndexes()

	// bind uniform buffers
	for name, bindLocation := range spec.UniformBufferBindings {
		if index, ok := prog.uniformBlockIndexes[name]; ok {
			gl.UniformBlockBinding(prog.id, index, bindLocation)
		}
	}

	// bind samplers to textureunits, this requires the program to be active
	for name, textureUnit := range spec.SamplerBindings {
		prog.samplerBindings[name] = textureUnit
	}

	prog.dirtySamplerBindings = true
	cmd.program = prog
	return nil
}

// NewProgram implements the core.RenderSystem interface.
func (r *RenderSystem) NewProgram(name string, data []byte) core.Program {
	var cmd = &makeProgramCommand{name: name, data: data}
	if err := r.Run(cmd, true); err != nil {
		glog.Fatal(err)
	}
	return cmd.program
}

func (p *Program) extractUniformNames() {
	var uniformCount int32
	gl.GetProgramiv(p.id, gl.ACTIVE_UNIFORMS, &uniformCount)

	for i := uint32(0); i < uint32(uniformCount); i++ {
		uniformName := strings.Repeat("\x00", int(128)+1)
		var uniformLen int32
		var uniformSize int32
		var uniformType uint32

		// extract locations and name
		gl.GetActiveUniform(p.id, i, 128, &uniformLen, &uniformSize, &uniformType, gl.Str(uniformName))

		// extract location
		location := gl.GetUniformLocation(p.id, gl.Str(uniformName))

		// save into location map
		goUniformname := gl.GoStr(gl.Str(uniformName))
		p.uniformLocations[goUniformname] = location
	}
}

func (p *Program) extractUniformBlockIndexes() {
	var blockCount int32
	gl.GetProgramiv(p.id, gl.ACTIVE_UNIFORM_BLOCKS, &blockCount)
	for i := uint32(0); i < uint32(blockCount); i++ {
		// extract name length
		var nameLen int32
		gl.GetActiveUniformBlockiv(p.id, i, gl.UNIFORM_BLOCK_NAME_LENGTH, &nameLen)
		uniformName := strings.Repeat("\x00", int(nameLen))

		// extract name
		gl.GetActiveUniformBlockName(p.id, i, nameLen, nil, gl.Str(uniformName))

		// extract location
		location := gl.GetUniformBlockIndex(p.id, gl.Str(uniformName))
		p.uniformBlockIndexes[gl.GoStr(gl.Str(uniformName))] = location
	}
}

func (p *Program) bind() {
	gl.UseProgram(p.id)

	if p.dirtySamplerBindings {
		for name, textureUnit := range p.samplerBindings {
			p.setUniform(name, &Uniform{textureUnit})
		}
		p.dirtySamplerBindings = false
	}

	currentProgram = p
}

func (p *Program) setUniform(name string, u *Uniform) {
	if u.value == nil {
		return
	}

	uloc, found := p.uniformLocations[name]
	if !found {
		return
	}

	glog.Infof("Name: %s Value: %#v", name, u)

	switch uval := u.Value().(type) {
	case mgl32.Mat4:
		gl.UniformMatrix4fv(uloc, 1, false, &uval[0])
	case mgl64.Mat4:
		newval := core.Mat4DoubleToFloat(uval)
		gl.UniformMatrix4fv(uloc, 1, false, &newval[0])
	case mgl64.Vec4:
		newval := core.Vec4DoubleToFloat(uval)
		gl.Uniform4fv(uloc, 1, &newval[0])
	case mgl64.Vec3:
		newval := core.Vec3DoubleToFloat(uval)
		gl.Uniform3fv(uloc, 1, &newval[0])
	case mgl64.Vec2:
		newval := core.Vec2DoubleToFloat(uval)
		gl.Uniform2fv(uloc, 1, &newval[0])
	case []float32:
		gl.Uniform1fv(uloc, int32(len(uval)), &uval[0])
	case []mgl32.Vec2:
		newval := make([]float32, len(uval)*2)
		for i := 0; i < len(uval); i++ {
			newval[i*2+0] = uval[i].X()
			newval[i*2+1] = uval[i].Y()
		}
		gl.Uniform2fv(uloc, int32(len(uval)), &newval[0])
	case int:
		gl.Uniform1i(uloc, int32(uval))
	case uint32:
		gl.Uniform1i(uloc, int32(uval))
	case float32:
		gl.Uniform1f(uloc, uval)
	case float64:
		gl.Uniform1f(uloc, float32(uval))
	default:
		glog.Fatalf("UNSUPPORTED -- Uniform: %s Type: %s\n", name, reflect.TypeOf(u.Value()))
	}
}

func (p *Program) setUniformBufferByName(name string, ub *UniformBuffer) {
	if _, ok := p.uniformBufferBindings[name]; !ok {
		return
	}

	gl.BindBufferBase(gl.UNIFORM_BUFFER, p.uniformBufferBindings[name], ub.id)
}

func (p *Program) setTexture(name string, texture *Texture) {
	if _, ok := p.samplerBindings[name]; !ok {
		return
	}

	textureUnit := p.samplerBindings[name]
	if textureUnitBindings[textureUnit] != texture.id {
		gl.ActiveTexture(gl.TEXTURE0 + textureUnit)
		gl.BindTexture(gl.TEXTURE_2D, texture.id)
		textureUnitBindings[textureUnit] = texture.id
	}
}

// Name implements the core.Program interface
func (p *Program) Name() string {
	return p.name
}
