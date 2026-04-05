# gosg — Scene Graph 3D Rendering Engine

## Overview
Go-based 3D scenegraph engine using **wgpu-native** (WebGPU) for rendering and **SDL3** for windowing. Targets macOS (Metal), Linux (Vulkan), and Windows (DX12/Vulkan).

## Architecture
- `core/` — Engine core: renderer, scene graph, camera, input, windowing (SDL3), all GPU types
- `gpu/` — Thin cgo bindings for wgpu-native C API
- `imgui/dearimgui/` — Dear ImGui integration via cgo
- `physics/bullet/` — Bullet physics via cgo
- `audio/openal/` — OpenAL audio (stubbed)
- `resource/filesystem/` — Filesystem-based asset loading
- `cmd/demo/` — Demo application
- `cmd/modeltool/` — Model file pack/unpack CLI tool (YAML manifests)
- `cmd/data/` — Shader programs (WGSL), pipeline states (JSON), models, textures

## Critical Rules

### Sandbox Limitations
**If a command fails due to network restrictions or sandbox limitations, ASK the user to run it.** Never make alternative technical decisions (e.g., switching libraries) to work around sandbox restrictions. The user can run the command themselves.

### wgpu Buffer Writes
**NEVER write to the same GPU buffer multiple times per frame.** `wgpuQueueWriteBuffer` stages data that is only applied at queue submission time. If the same buffer is written multiple times during render pass recording, only the last write's data will be visible to ALL draw commands that reference that buffer. This causes flickering/corruption.

**Solutions:**
- Use a single large buffer with offsets (like the ImGui renderer does)
- Collect all instances before issuing the draw call (like AABBRenderTechnique)
- Use separate buffers per draw call if offsets aren't possible
- Call `renderer.Flush()` between passes that reuse the same buffers (submits the command buffer so the GPU consumes writes before they're overwritten)

### cgo Pointer Rules
When passing Go structs to wgpu C functions that contain `nextInChain` pointer chains or array pointers, the cgo pointer checker will panic if Go pointers point to other Go pointers. Solutions:
- Use `C.calloc` to allocate structs in C memory (for chained struct patterns like surface/shader creation)
- Use `runtime.Pinner` to pin Go memory (for complex structs like render pipeline descriptors)
- Never pass Go slice backing arrays directly as struct member pointers without pinning

### Buffer Alignment
wgpu requires buffer sizes and `QueueWriteBuffer` copy sizes to be aligned to **4 bytes** (`COPY_BUFFER_ALIGNMENT`). Always align: `size = (size + 3) &^ 3`

### Coordinate System (WebGPU vs OpenGL)
- **NDC Z range**: WebGPU uses [0, 1], OpenGL uses [-1, 1]
- **Frustum plane extraction**: Near plane is `rowZ` (not `rowW + rowZ`) for Z≥0 convention
- **Shadow bias matrix**: Z stays [0,1] (no remap needed), Y must be flipped for texture sampling
- Use `PerspectiveWebGPU()` and `OrthoWebGPU()` from `core/projection.go` — never use `mgl64.Perspective` or `mgl64.Ortho` directly

### High-DPI / Retina
- `WindowSize()` returns **pixel** dimensions (for GPU surfaces, viewports, framebuffers)
- `WindowSizePoints()` returns **point** dimensions (for UI layout, ImGui)
- `PixelDensity()` returns the scale factor (e.g., 2.0 on Retina)
- ImGui works in point space; scissor rects must be scaled by PixelDensity for the GPU
- Fullscreen mode queries native display resolution via `SDL_GetDesktopDisplayMode`
- SDL3 mouse coordinates in relative mode: use `xrel`/`yrel` for deltas, not absolute position differences

### Pipeline State
- In wgpu, primitive topology (triangles/lines/points) is baked into the render pipeline, not set per-draw
- The `topology` field in state JSON files controls this (default: "triangles")
- Depth/blend/cull state are also baked into pipelines (cached by `pipelineCache`)
- Depth-only render passes (e.g., shadow maps) use 0 color targets in the pipeline

### Shadow Mapping
- Uses depth-only `texture_depth_2d_array` with hardware PCF via `textureSampleCompareLevel`
- Cascade count is configurable via `NewShadowMap(size, cascades)`
- PSSM splits (logarithmic + uniform blend) with tight frustum fitting
- Front-face culling during shadow pass to prevent self-shadowing
- Comparison sampler (`sampler_comparison` with `LessEqual`) for hardware depth test
- Shadow bias tunable at runtime via ImGui slider (passed through `light.ShadowBias` → `light.Block.Color.w`)

### Input Handling
- ImGui only consumes input it actually uses (`WantsCaptureMouse`/`WantsCaptureKeyboard`)
- Unconsumed input passes through to underlying scenes in the scene stack

### User Preferences
- YAML for config/manifest files
- No vendored libraries — use system packages (homebrew)

## Build
```sh
# Dependencies (macOS): brew install sdl3 wgpu-native
go build ./...

# Run demo
go build -o demo ./cmd/demo && ./demo -data ./cmd/data -logtostderr

# Model tool
go build -o modeltool ./cmd/modeltool
modeltool unpack model.model output_dir    # extracts + generates manifest.yaml
modeltool pack manifest.yaml               # packs from manifest
```

## Key Files
- `gpu/gpu.go` — wgpu-native cgo bindings
- `core/rendersystem.go` — Renderer, Dispatch, render techniques
- `core/pipeline.go` — Pipeline cache
- `core/program.go` — Shader program loading (.wgpu.json specs)
- `core/mesh.go` — Mesh with GPU buffers
- `core/texture.go` — Texture with view + sampler
- `core/shadow.go` — Cascaded shadow maps with PCF
- `core/window.go` — SDL3 windowing
- `core/projection.go` — WebGPU-correct projection matrices
- `core/imgui_renderer.go` — ImGui wgpu rendering
- `core/state.go` — Pipeline state definition (JSON with string enums)
- `core/model.go` — Model loading with minimal protobuf wire decoder
