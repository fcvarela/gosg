package core

import (
	"unsafe"

	"github.com/fcvarela/gosg/gpu"
)

// imguiRenderer handles Dear ImGui rendering with its own pipeline and vertex format.
type imguiRenderer struct {
	vertexBuffer gpu.Buffer
	indexBuffer  gpu.Buffer
	vertexSize   uint64
	indexSize    uint64
	pipeline     gpu.RenderPipeline
	initialized  bool
}

var imgui imguiRenderer

func (ir *imguiRenderer) ensurePipeline(program *Program, colorFormat gpu.TextureFormat) {
	if ir.initialized {
		return
	}

	buffers := []gpu.VertexBufferLayout{{
		ArrayStride: 20, // ImDrawVert: pos(2xf32) + uv(2xf32) + col(u32) = 20 bytes
		StepMode:    gpu.VertexStepModeVertex,
		Attributes: []gpu.VertexAttribute{
			{Format: gpu.VertexFormatFloat32x2, Offset: 0, ShaderLocation: 0},
			{Format: gpu.VertexFormatFloat32x2, Offset: 8, ShaderLocation: 1},
			{Format: gpu.VertexFormatUnorm8x4, Offset: 16, ShaderLocation: 2},
		},
	}}

	desc := gpu.RenderPipelineDescriptor{
		Layout:         program.pipelineLayout,
		VertexModule:   program.vertexModule,
		VertexEntry:    "main",
		FragmentModule: program.fragmentModule,
		FragmentEntry:  "main",
		Buffers:        buffers,
		Primitive:      gpu.PrimitiveTopologyTriangleList,
		FrontFace:      gpu.FrontFaceCCW,
		CullMode:       gpu.CullModeNone,
		Targets: []gpu.ColorTargetState{{
			Format:    colorFormat,
			WriteMask: gpu.ColorWriteMaskAll,
			Blend: &gpu.BlendState{
				Color: gpu.BlendComponent{
					SrcFactor: gpu.BlendFactorSrcAlpha,
					DstFactor: gpu.BlendFactorOneMinusSrcAlpha,
					Operation: gpu.BlendOperationAdd,
				},
				Alpha: gpu.BlendComponent{
					SrcFactor: gpu.BlendFactorOne,
					DstFactor: gpu.BlendFactorOneMinusSrcAlpha,
					Operation: gpu.BlendOperationAdd,
				},
			},
		}},
	}

	ir.pipeline = renderer.device.CreateRenderPipeline(desc)
	ir.initialized = true
}

func (ir *imguiRenderer) ensureBuffers(vertexBytes, indexBytes uint64) {
	vertexBytes = (vertexBytes + 3) &^ 3
	indexBytes = (indexBytes + 3) &^ 3

	if ir.vertexSize < vertexBytes {
		ir.vertexBuffer.Release()
		ir.vertexBuffer = renderer.device.CreateBuffer(vertexBytes, gpu.BufferUsageVertex|gpu.BufferUsageCopyDst)
		ir.vertexSize = vertexBytes
	}
	if ir.indexSize < indexBytes {
		ir.indexBuffer.Release()
		ir.indexBuffer = renderer.device.CreateBuffer(indexBytes, gpu.BufferUsageIndex|gpu.BufferUsageCopyDst)
		ir.indexSize = indexBytes
	}
}

func (ir *imguiRenderer) draw(rp *RenderPass) {
	if imguiSystem == nil {
		return
	}

	drawData := imguiSystem.GetDrawData()
	if drawData == nil || drawData.CommandListCount() == 0 {
		return
	}

	program := rp.CurrentProgram()
	if program == nil {
		return
	}

	ir.ensurePipeline(program, renderer.surfaceFormat)
	rp.SetGPUPipeline(ir.pipeline)

	listCount := drawData.CommandListCount()

	// Compute total buffer sizes across all command lists
	totalVertexBytes := uint64(0)
	totalIndexBytes := uint64(0)
	for i := 0; i < listCount; i++ {
		cl := drawData.GetCommandList(i)
		totalVertexBytes += uint64(cl.VertexBufferSize * 20)
		totalIndexBytes += (uint64(cl.IndexBufferSize*2) + 3) &^ 3
	}

	ir.ensureBuffers(totalVertexBytes, totalIndexBytes)

	// Upload all command list data at offsets, then draw
	vertexOffset := uint64(0)
	indexOffset := uint64(0)

	for i := 0; i < listCount; i++ {
		cmdList := drawData.GetCommandList(i)

		vBytes := uint64(cmdList.VertexBufferSize * 20)
		iBytes := uint64(cmdList.IndexBufferSize * 2)
		iBytesAligned := (iBytes + 3) &^ 3

		// Upload vertex data at offset
		renderer.queue.WriteBuffer(ir.vertexBuffer, vertexOffset, cmdList.VertexPointer, vBytes)

		// Upload index data at offset (with padding)
		if iBytes > 0 {
			padded := make([]byte, iBytesAligned)
			copy(padded, unsafe.Slice((*byte)(cmdList.IndexPointer), iBytes))
			renderer.queue.WriteBuffer(ir.indexBuffer, indexOffset, unsafe.Pointer(&padded[0]), iBytesAligned)
		}

		// Bind this command list's region of the buffers
		rp.SetVertexBuffer(0, ir.vertexBuffer, vertexOffset, vBytes)
		rp.SetIndexBuffer(ir.indexBuffer, gpu.IndexFormatUint16, indexOffset, iBytesAligned)

		var elemOffset uint32
		for _, cmd := range cmdList.Commands {
			// Scissor rect — ImGui clip rects are in point space, GPU needs pixels
			scale := windowManager.PixelDensity()
			vpW := int32(renderer.surfaceWidth)
			vpH := int32(renderer.surfaceHeight)
			cx := int32(cmd.ClipRect[0] * scale)
			cy := int32(cmd.ClipRect[1] * scale)
			cr := int32(cmd.ClipRect[2] * scale)
			cb := int32(cmd.ClipRect[3] * scale)
			if cx < 0 { cx = 0 }
			if cy < 0 { cy = 0 }
			if cr > vpW { cr = vpW }
			if cb > vpH { cb = vpH }
			cw, ch := cr-cx, cb-cy
			if cw > 0 && ch > 0 {
				rp.SetScissorRect(uint32(cx), uint32(cy), uint32(cw), uint32(ch))
			}

			// Bind texture
			if cmd.TextureID != nil && len(program.bindGroupLayouts) >= 2 {
				tex := (*Texture)(cmd.TextureID)
				bg := renderer.device.CreateBindGroup(program.bindGroupLayouts[1], []gpu.BindGroupEntry{
					{Binding: 0, TextureView: tex.view},
					{Binding: 1, Sampler: tex.sampler},
				})
				rp.SetBindGroup(1, bg)
				bg.Release()
			}

			rp.DrawIndexed(uint32(cmd.ElementCount), 1, elemOffset, 0, 0)
			elemOffset += uint32(cmd.ElementCount)
		}

		vertexOffset += vBytes
		indexOffset += iBytesAligned
	}
}
