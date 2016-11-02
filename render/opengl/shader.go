package opengl

import (
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/golang/glog"
)

type shader struct {
	name   string
	stype  uint32
	id     uint32
	source []byte
}

func newShader(name string, shaderType uint32, data []byte) *shader {
	s := shader{name, shaderType, 0, data}
	return &s
}

func (s *shader) compile() {
	s.id = gl.CreateShader(s.stype)

	sources, free := gl.Strs(string(s.source) + "\x00")
	gl.ShaderSource(s.id, 1, sources, nil)
	free()
	gl.CompileShader(s.id)

	var status int32
	gl.GetShaderiv(s.id, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(s.id, gl.INFO_LOG_LENGTH, &logLength)
		compileLog := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(s.id, logLength, nil, gl.Str(compileLog))
		glog.Fatalf("Failed to compile shader %s\n%v\n%v\n", s.name, string(s.source), compileLog)
	}

	s.source = nil
}
