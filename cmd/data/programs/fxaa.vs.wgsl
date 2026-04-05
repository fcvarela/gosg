// fxaa vertex shader — orthographic projection, pass texcoords

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
    @location(0) position: vec3f,
    @location(1) normal: vec3f,
    @location(2) tcoords0: vec3f,
    @location(3) mMatrix0: vec4f,
    @location(4) mMatrix1: vec4f,
    @location(5) mMatrix2: vec4f,
    @location(6) mMatrix3: vec4f,
    @location(7) mvpMatrix0: vec4f,
    @location(8) mvpMatrix1: vec4f,
    @location(9) mvpMatrix2: vec4f,
    @location(10) mvpMatrix3: vec4f,
    @location(11) custom1: vec4f,
    @location(12) custom2: vec4f,
    @location(13) custom3: vec4f,
    @location(14) custom4: vec4f,
};

struct VertexOutput {
    @builtin(position) position: vec4f,
    @location(0) tcoords0: vec3f,
};

@vertex
fn main(in: VertexInput) -> VertexOutput {
    var out: VertexOutput;
    out.position = camera.pMatrix * vec4f(in.position, 1.0);
    out.tcoords0 = in.tcoords0;
    return out;
}
