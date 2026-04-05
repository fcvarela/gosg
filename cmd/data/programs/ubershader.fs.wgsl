// ubershader fragment shader — PBR Cook-Torrance with shadow cascades

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

// Material textures
@group(1) @binding(0) var albedoTex: texture_2d<f32>;
@group(1) @binding(1) var albedoSampler: sampler;
@group(1) @binding(2) var normalTex: texture_2d<f32>;
@group(1) @binding(3) var normalSampler: sampler;
@group(1) @binding(4) var roughTex: texture_2d<f32>;
@group(1) @binding(5) var roughSampler: sampler;
@group(1) @binding(6) var metalTex: texture_2d<f32>;
@group(1) @binding(7) var metalSampler: sampler;

// Shadow texture array — single depth array with comparison sampler
@group(1) @binding(8) var shadowTex: texture_depth_2d_array;
@group(1) @binding(9) var shadowSampler: sampler_comparison;

struct FragmentInput {
    @builtin(position) fragCoord: vec4f,
    @location(0) worldPosition: vec3f,
    @location(1) cameraPosition: vec3f,
    @location(2) tcoords0: vec3f,
    @location(3) tangent: vec3f,
    @location(4) bitangent: vec3f,
    @location(5) normal: vec3f,
};

// --- Shadow functions ---

fn pcfShadow(coords: vec2f, layer: i32, compare: f32, bias: f32) -> f32 {
    let texelSize = 1.0 / 2048.0;
    var shadow = 0.0;

    // 5x5 PCF kernel
    for (var x: i32 = -2; x <= 2; x++) {
        for (var y: i32 = -2; y <= 2; y++) {
            let offset = vec2f(f32(x), f32(y)) * texelSize;
            shadow += textureSampleCompareLevel(shadowTex, shadowSampler, coords + offset, layer, compare - bias);
        }
    }

    return shadow / 25.0;
}

fn shadow(worldPos: vec4f, lightIndex: i32, N: vec3f, L: vec3f) -> f32 {
    let baseBias = camera.lights[lightIndex].color.w;
    let cosTheta = clamp(dot(N, L), 0.0, 1.0);
    let bias = max(baseBias * (1.0 - cosTheta), baseBias * 0.1);

    let viewPos = camera.vMatrix * worldPos;
    let viewDepth = -viewPos.z;

    // Find which cascade this fragment belongs to
    for (var c: i32 = 0; c < 10; c++) {
        if (viewDepth < camera.lights[lightIndex].zCuts[c].x) {
            let ccoords = camera.lights[lightIndex].vpMatrix[c] * worldPos;
            let shadowCoords = ccoords.xyz / ccoords.w;
            return pcfShadow(shadowCoords.xy, c, shadowCoords.z, bias);
        }
    }

    return 1.0;
}

// Beckmann distribution
fn distribution(n: vec3f, h: vec3f, roughness: f32) -> f32 {
    let m_Sq = roughness * roughness;
    var NdotH_Sq = max(dot(n, h), 0.0);
    NdotH_Sq = NdotH_Sq * NdotH_Sq;
    return exp((NdotH_Sq - 1.0) / (m_Sq * NdotH_Sq)) / (3.14159265 * m_Sq * NdotH_Sq * NdotH_Sq);
}

// Cook-Torrance geometry term
fn geometry(n: vec3f, h: vec3f, v: vec3f, l: vec3f, roughness: f32) -> f32 {
    let NdotH = dot(n, h);
    let NdotL = dot(n, l);
    let NdotV = dot(n, v);
    let VdotH = dot(v, h);
    let NdotL_clamped = max(NdotL, 0.0);
    let NdotV_clamped = max(NdotV, 0.0);
    return min(min(2.0 * NdotH * NdotV_clamped / VdotH, 2.0 * NdotH * NdotL_clamped / VdotH), 1.0);
}

// Schlick fresnel
fn fresnel(f0: f32, n: vec3f, l: vec3f) -> f32 {
    return f0 + (1.0 - f0) * pow(1.0 - dot(n, l), 5.0);
}

// Fresnel diffuse energy ratio
fn diffuse_energy_ratio(f0: f32, n: vec3f, l: vec3f) -> f32 {
    return 1.0 - fresnel(f0, n, l);
}

@fragment
fn main(in: FragmentInput) -> @location(0) vec4f {
    var color = vec4f(0.0);

    // sample material textures
    let albedo = textureSample(albedoTex, albedoSampler, in.tcoords0.xy);
    let normalMap = textureSample(normalTex, normalSampler, in.tcoords0.xy).xyz * 2.0 - 1.0;
    let metalness = textureSample(metalTex, metalSampler, in.tcoords0.xy).r;
    let roughness = 0.1 + 0.8 * textureSample(roughTex, roughSampler, in.tcoords0.xy).r;

    // f0 adjusted from 0.118 to 0.818
    let f0 = 0.118 + metalness * 0.7;

    // view direction and TBN-transformed normal
    let V = normalize(in.cameraPosition - in.worldPosition);
    let tbn = mat3x3f(in.tangent, in.bitangent, in.normal);
    let N = normalize(tbn * normalMap);

    // shared products
    let NdotV = dot(N, V);
    let NdotV_clamped = max(NdotV, 0.0000000001);

    let lc = i32(camera.lightCount[0]);
    for (var i: i32 = 0; i < lc; i++) {
        // light direction and half vector
        let L = normalize(camera.lights[i].position.xyz - in.worldPosition);
        let H = normalize(L + V);

        let NdotL = dot(N, L);
        let NdotL_clamped = max(NdotL, 0.0);

        let fres = fresnel(f0, H, L);
        let geom = geometry(N, H, V, L, roughness);
        let ndf = distribution(N, H, roughness);

        var brdf_spec = (0.25 * fres * geom * ndf) / (NdotL_clamped * NdotV_clamped);
        if (NdotV <= 0.0 || NdotL <= 0.0) {
            brdf_spec = 0.0;
        }

        let color_spec = NdotL_clamped * brdf_spec * (camera.lights[i].color.rgb * (1.0 - metalness) + albedo.rgb * metalness);
        let color_diff = NdotL_clamped * diffuse_energy_ratio(f0, N, L) * albedo.rgb * camera.lights[i].color.rgb;
        let sh = shadow(vec4f(in.worldPosition, 1.0), i, N, L);
        color = vec4f(color.rgb + (0.05 * albedo.rgb + color_diff + color_spec) * sh, color.a);
    }

    color = vec4f(color.rgb, albedo.a);
    return color;
}
