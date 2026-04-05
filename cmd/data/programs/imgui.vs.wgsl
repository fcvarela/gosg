// imgui vertex shader — 2D vertex with projection matrix

const MAX_CASCADES: u32 = 10;
const MAX_LIGHTS: u32 = 16;

struct Light {
    vpMatrix: array<mat4x4f, MAX_CASCADES>,
    zCuts: array<vec4f, MAX_CASCADES>,
    position: vec4f,
    color: vec4f,
};

struct CameraConstants {
    vMatrix: mat4x4f,
    pMatrix: mat4x4f,
    vpMatrix: mat4x4f,
    lightCount: vec4f,
    cameraWorldPosition: vec4f,
    lights: array<Light, MAX_LIGHTS>,
};

@group(0) @binding(0) var<uniform> camera: CameraConstants;

struct VertexInput {
    @location(0) position: vec2f,
    @location(1) tcoords0: vec2f,
    @location(2) color: vec4f,
};

struct VertexOutput {
    @builtin(position) position: vec4f,
    @location(0) fragUV: vec2f,
    @location(1) fragColor: vec4f,
};

@vertex
fn main(in: VertexInput) -> VertexOutput {
    var out: VertexOutput;
    out.fragUV = in.tcoords0;
    out.fragColor = in.color;
    out.position = camera.pMatrix * vec4f(in.position, 0.0, 1.0);
    return out;
}
