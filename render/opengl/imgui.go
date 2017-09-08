package opengl

import (
	"github.com/fcvarela/gosg/core"
	"github.com/go-gl/gl/v3.3-core/gl"
)

var (
	imguiBuffers *buffers
)

func (r *RenderSystem) drawIMGUI(cmd *core.DrawIMGUICommand) {
	bindVAO(imguiBuffers.vao)

	imguiSystem := core.GetIMGUISystem()
	drawData := imguiSystem.GetDrawData()

	gl.EnableVertexAttribArray(0)
	gl.EnableVertexAttribArray(1)
	gl.EnableVertexAttribArray(2)

	var lastTexture int32
	var lastMipmapMode int32
	gl.ActiveTexture(gl.TEXTURE0)
	gl.GetIntegerv(gl.TEXTURE_BINDING_2D, &lastTexture)
	gl.GetIntegerv(gl.TEXTURE_MIN_FILTER, &lastMipmapMode)

	for i := 0; i < drawData.CommandListCount(); i++ {
		cmdlist := drawData.GetCommandList(i)

		gl.BindBuffer(gl.ARRAY_BUFFER, imguiBuffers.buffers[positionBuffer])
		gl.BufferData(gl.ARRAY_BUFFER, cmdlist.VertexBufferSize*5*4, cmdlist.VertexPointer, gl.STREAM_DRAW)

		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, imguiBuffers.buffers[indexBuffer])
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, cmdlist.IndexBufferSize*2, cmdlist.IndexPointer, gl.STREAM_DRAW)

		// position = 0, tcoords = 1, normals/color = 2
		gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(0))
		gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(8))
		gl.VertexAttribPointer(2, 4, gl.UNSIGNED_BYTE, true, 5*4, gl.PtrOffset(16))

		var elementIndex int
		for _, cmd := range cmdlist.Commands {
			if tex := (*Texture)(cmd.TextureID); tex != nil {
				gl.BindTexture(gl.TEXTURE_2D, tex.id)
				gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
			}

			gl.Scissor(
				int32(cmd.ClipRect[0]),
				int32(imguiSystem.DisplaySize().Y()-cmd.ClipRect[3]),
				int32(cmd.ClipRect[2]-cmd.ClipRect[0]),
				int32(cmd.ClipRect[3]-cmd.ClipRect[1]))
			gl.DrawElements(gl.TRIANGLES, int32(cmd.ElementCount), gl.UNSIGNED_SHORT, gl.PtrOffset(elementIndex))
			elementIndex += cmd.ElementCount * 2
		}
	}

	gl.BindTexture(gl.TEXTURE_2D, (uint32)(lastTexture))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, lastMipmapMode)
	bindVAO(0)
}
