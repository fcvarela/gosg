package core

// Framebuffer holds render target attachments.
type Framebuffer struct {
	colorAttachments map[int]*Texture
	depthAttachment  *Texture
	// GPU handle will be added in Phase 2
}

// NewFramebuffer creates a new framebuffer.
func NewFramebuffer() *Framebuffer {
	return &Framebuffer{
		colorAttachments: make(map[int]*Texture),
	}
}

// SetColorAttachment sets the color attachment at the specified index.
func (f *Framebuffer) SetColorAttachment(index int, attachment *Texture) {
	f.colorAttachments[index] = attachment
}

// ColorAttachments returns the framebuffer color attachments.
func (f *Framebuffer) ColorAttachments() map[int]*Texture {
	return f.colorAttachments
}

// ColorAttachment returns the color attachment at the specified index.
func (f *Framebuffer) ColorAttachment(index int) *Texture {
	return f.colorAttachments[index]
}

// SetDepthAttachment sets the depth attachment.
func (f *Framebuffer) SetDepthAttachment(attachment *Texture) {
	f.depthAttachment = attachment
}

// DepthAttachment returns the framebuffer depth attachment.
func (f *Framebuffer) DepthAttachment() *Texture {
	return f.depthAttachment
}
