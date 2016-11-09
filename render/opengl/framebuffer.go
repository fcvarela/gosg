package opengl

import (
	"github.com/fcvarela/gosg/core"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/golang/glog"
)

// Framebuffer implements the core.Framebuffer interface
type Framebuffer struct {
	fbo              uint32
	depthAttachment  core.Texture
	colorAttachments map[int]core.Texture
}

var (
	currentFBO = uint32(0)
)

// NewFramebuffer implements the core.RenderSystem interface
func (r *RenderSystem) NewFramebuffer() core.Framebuffer {
	fb := &Framebuffer{0, nil, make(map[int]core.Texture)}
	gl.GenFramebuffers(1, &fb.fbo)
	return fb
}

// validates the current bound framebuffer
func validateCurrentFramebuffer() {
	// check status
	status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER)
	switch status {
	case gl.FRAMEBUFFER_INCOMPLETE_ATTACHMENT:
		glog.Fatal("Incomplete attachment")
	case gl.FRAMEBUFFER_INCOMPLETE_MISSING_ATTACHMENT:
		glog.Fatal("Missing attachment")
	case gl.FRAMEBUFFER_UNSUPPORTED:
		glog.Fatal("Invalid combination of internal formats")
	case gl.FRAMEBUFFER_INCOMPLETE_DRAW_BUFFER:
		glog.Fatal("Incomplete draw buffer")
	case gl.FRAMEBUFFER_INCOMPLETE_READ_BUFFER:
		glog.Fatal("Incomplete read buffer")
	default:
		break
	}
}

// SetDepthAttachment implements the core.RenderTarger interface
func (f *Framebuffer) SetDepthAttachment(attachment core.Texture) {
	f.depthAttachment = attachment
	bindFramebuffer(f.fbo)
	gl.FramebufferTexture(gl.DRAW_FRAMEBUFFER, gl.DEPTH_ATTACHMENT, f.depthAttachment.(*Texture).id, 0)
	validateCurrentFramebuffer()
}

// DepthAttachment implements the core.RenderTarger interface
func (f *Framebuffer) DepthAttachment() core.Texture {
	return f.depthAttachment
}

// SetColorAttachment implements the core.RenderTarger interface
func (f *Framebuffer) SetColorAttachment(index int, attachment core.Texture) {
	f.colorAttachments[index] = attachment
	bindFramebuffer(f.fbo)

	// rebuild drawbuffers
	drawBuffers := make([]uint32, len(f.colorAttachments))
	for i := range f.colorAttachments {
		drawBuffers[i] = uint32(gl.COLOR_ATTACHMENT0 + i)
		gl.FramebufferTexture(gl.DRAW_FRAMEBUFFER, uint32(gl.COLOR_ATTACHMENT0+i), f.colorAttachments[i].(*Texture).id, 0)
	}

	if len(drawBuffers) > 0 {
		gl.DrawBuffers(int32(len(drawBuffers)), &drawBuffers[0])
	} else {
		gl.DrawBuffer(gl.NONE)
	}
	validateCurrentFramebuffer()
}

// ColorAttachment implements the core.RenderTarger interface
func (f *Framebuffer) ColorAttachment(index int) core.Texture {
	return f.colorAttachments[index]
}

// ColorAttachments implements the core.RenderTarger interface
func (f *Framebuffer) ColorAttachments() map[int]core.Texture {
	return f.colorAttachments
}

func bindFramebuffer(fbo uint32) {
	if currentFBO == fbo {
		return
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)
	currentFBO = fbo
}
