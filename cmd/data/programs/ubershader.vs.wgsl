// ubershader vertex shader — full PBR: world position, TBN matrix, texcoords

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
    @builtin(position) clipPosition: vec4f,
    @location(0) worldPosition: vec3f,
    @location(1) cameraPosition: vec3f,
    @location(2) tcoords0: vec3f,
    // TBN matrix passed as 3 row vectors
    @location(3) tangent: vec3f,
    @location(4) bitangent: vec3f,
    @location(5) normal: vec3f,
};

@vertex
fn main(in: VertexInput) -> VertexOutput {
    let mMatrix = mat4x4f(in.mMatrix0, in.mMatrix1, in.mMatrix2, in.mMatrix3);
    let mvpMatrix = mat4x4f(in.mvpMatrix0, in.mvpMatrix1, in.mvpMatrix2, in.mvpMatrix3);

    var out: VertexOutput;

    // clip position
    out.clipPosition = mvpMatrix * vec4f(in.position, 1.0);

    // world position
    out.worldPosition = (mMatrix * vec4f(in.position, 1.0)).xyz;

    // camera world position from UBO (no inverse needed)
    out.cameraPosition = camera.cameraWorldPosition.xyz;

    // world-space normal
    var normal = normalize((mMatrix * vec4f(in.normal, 0.0)).xyz);

    // compute tangent and bitangent
    var tangent = normalize(cross(normal, vec3f(0.0, 0.0, 1.0)));
    var bitangent = normalize(cross(normal, vec3f(1.0, 0.0, 0.0)));

    tangent = normalize((mMatrix * vec4f(tangent, 0.0)).xyz);
    bitangent = normalize((mMatrix * vec4f(bitangent, 0.0)).xyz);

    out.tangent = tangent;
    out.bitangent = bitangent;
    out.normal = normal;

    // texture coordinates
    out.tcoords0 = in.tcoords0;

    return out;
}
