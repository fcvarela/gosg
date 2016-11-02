package core

// Framebuffer is an interface which wraps a render target. This contains
// information about depth and color attachments, dimensions.
type Framebuffer interface {
	// SetColorAttachment sets the color attachment at the specified index.
	SetColorAttachment(index int, attachment Texture)

	// ColorAttachments returns the framebuffer color attachments.
	ColorAttachments() map[int]Texture

	// ColorAttachment returns the color attachment at the specified index.
	ColorAttachment(index int) Texture

	// SetDepthAttachment sets the depth attachment.
	SetDepthAttachment(attachment Texture)

	// DepthAttachment returns the framebuffer depth attachment
	DepthAttachment() Texture
}
