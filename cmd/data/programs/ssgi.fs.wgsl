// SSGI fragment shader — screen-space global illumination via hemisphere sampling

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

@group(1) @binding(0) var colorTex: texture_2d<f32>;
@group(1) @binding(1) var colorSamp: sampler;
@group(1) @binding(2) var normalTex: texture_2d<f32>;
@group(1) @binding(3) var normalSamp: sampler;
@group(1) @binding(4) var depthTex: texture_2d<f32>;
@group(1) @binding(5) var depthSamp: sampler;

struct FragmentInput {
    @builtin(position) fragCoord: vec4f,
    @location(0) tcoords0: vec3f,
};

const NUM_SAMPLES: i32 = 16;
const SAMPLE_RADIUS: f32 = 2.0;
const THICKNESS: f32 = 0.5;
const INTENSITY: f32 = 1.5;

// Hash function for per-pixel randomization
fn hash(p: vec2f) -> f32 {
    var h = dot(p, vec2f(127.1, 311.7));
    return fract(sin(h) * 43758.5453123);
}

fn hash2(p: vec2f) -> vec2f {
    return vec2f(
        hash(p),
        hash(p + vec2f(37.0, 17.0))
    );
}

// Reconstruct view-space position from depth and UV
fn reconstructViewPos(uv: vec2f, depth: f32) -> vec3f {
    // UV is already flipped (0=top), convert to NDC: x [-1,1], y [-1,1] (flip back)
    let ndc = vec4f(uv.x * 2.0 - 1.0, (1.0 - uv.y) * 2.0 - 1.0, depth, 1.0);
    let invProj = camera.pMatrix; // We'll pass inverse projection via pMatrix slot
    // For now, use the projection matrix to reconstruct
    // viewPos.x = ndc.x / P[0][0], viewPos.y = ndc.y / P[1][1], viewPos.z from depth
    let viewPos = vec3f(
        ndc.x / camera.pMatrix[0][0],
        ndc.y / camera.pMatrix[1][1],
        -1.0 // placeholder
    );

    // Linearize depth: WebGPU Z in [0,1], perspective projection
    // z_ndc = (far * z_view + near * far) / (-z_view * (far - near))
    // Simplified: z_view = pMatrix[3][2] / (depth - pMatrix[2][2])
    let zView = camera.pMatrix[3][2] / (depth - camera.pMatrix[2][2]);

    return vec3f(viewPos.x * (-zView), viewPos.y * (-zView), zView);
}

// Generate a cosine-weighted hemisphere sample direction
fn hemisphereDir(normal: vec3f, idx: i32, noise: vec2f) -> vec3f {
    let fi = f32(idx);
    // Golden angle spiral + noise for better distribution
    let phi = 2.399963 * fi + noise.x * 6.283185; // golden angle
    let cosTheta = sqrt(1.0 - (fi + noise.y) / f32(NUM_SAMPLES));
    let sinTheta = sqrt(1.0 - cosTheta * cosTheta);

    let localDir = vec3f(cos(phi) * sinTheta, sin(phi) * sinTheta, cosTheta);

    // Build TBN from normal
    var up = vec3f(0.0, 1.0, 0.0);
    if (abs(normal.y) > 0.99) {
        up = vec3f(1.0, 0.0, 0.0);
    }
    let tangent = normalize(cross(up, normal));
    let bitangent = cross(normal, tangent);

    return tangent * localDir.x + bitangent * localDir.y + normal * localDir.z;
}

@fragment
fn main(in: FragmentInput) -> @location(0) vec4f {
    let uv = vec2f(in.tcoords0.x, 1.0 - in.tcoords0.y);
    let texSize = vec2f(textureDimensions(colorTex, 0));

    // Read G-buffer
    let depth = textureSample(depthTex, depthSamp, uv).r;
    if (depth >= 1.0) {
        return vec4f(0.0); // sky, no SSGI
    }

    let normalEnc = textureSample(normalTex, normalSamp, uv).xyz;
    let normal = normalize(normalEnc * 2.0 - 1.0);
    let viewPos = reconstructViewPos(uv, depth);

    // Per-pixel noise for sample rotation
    let noise = hash2(in.fragCoord.xy);

    var indirect = vec3f(0.0);

    for (var i: i32 = 0; i < NUM_SAMPLES; i++) {
        // Random direction in hemisphere around the view-space normal
        let sampleDir = hemisphereDir(normal, i, noise);

        // Offset position along sample direction
        let samplePos = viewPos + sampleDir * SAMPLE_RADIUS;

        // Project sample position to screen space
        let projPos = camera.pMatrix * vec4f(samplePos, 1.0);
        let sampleNDC = projPos.xy / projPos.w;
        let sampleUV = vec2f(sampleNDC.x * 0.5 + 0.5, 1.0 - (sampleNDC.y * 0.5 + 0.5));

        // Bounds check
        if (sampleUV.x < 0.0 || sampleUV.x > 1.0 || sampleUV.y < 0.0 || sampleUV.y > 1.0) {
            continue;
        }

        // Read depth at sample screen position
        let sampleDepth = textureSample(depthTex, depthSamp, sampleUV).r;
        let sampleViewPos = reconstructViewPos(sampleUV, sampleDepth);

        // Check if sample is occluded (behind actual geometry)
        let depthDiff = viewPos.z - sampleViewPos.z;

        if (depthDiff > 0.0 && depthDiff < THICKNESS) {
            // Hit: sample the lit color and weight by NdotL equivalent
            let hitColor = textureSample(colorTex, colorSamp, sampleUV).rgb;
            let weight = max(dot(normal, sampleDir), 0.0);
            indirect += hitColor * weight;
        }
    }

    indirect = indirect / f32(NUM_SAMPLES) * INTENSITY;

    return vec4f(indirect, 1.0);
}
