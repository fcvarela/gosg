// fxaa fragment shader — FXAA + Uncharted2 tonemapping + gamma correction

@group(1) @binding(0) var colorTexture: texture_2d<f32>;
@group(1) @binding(1) var colorSampler: sampler;
@group(1) @binding(2) var ssgiTexture: texture_2d<f32>;
@group(1) @binding(3) var ssgiSampler: sampler;

struct FragmentInput {
    @location(0) tcoords0: vec3f,
};

// Uncharted2 tonemap constants
const A: f32 = 0.15;
const B: f32 = 0.50;
const C: f32 = 0.10;
const D: f32 = 0.20;
const E: f32 = 0.02;
const F: f32 = 0.30;
const W: f32 = 11.2;

fn Uncharted2Tonemap(x: vec3f) -> vec3f {
    return ((x * (A * x + C * B) + D * E) / (x * (A * x + B) + D * F)) - E / F;
}

fn tonemapUncharted2(color: vec3f) -> vec3f {
    let ExposureBias = 2.0;
    let curr = Uncharted2Tonemap(ExposureBias * color);
    let whiteScale = 1.0 / Uncharted2Tonemap(vec3f(W));
    return curr * whiteScale;
}

fn fxaa(tcoords: vec2f) -> vec3f {
    let FXAA_SPAN_MAX = 8.0;
    let FXAA_REDUCE_MIN = 1.0 / 128.0;
    let FXAA_REDUCE_MUL = 1.0 / 8.0;

    let colorTextureSize = vec2f(textureDimensions(colorTexture, 0));
    let texCoordOffset = 1.0 / colorTextureSize;

    let luma = vec3f(0.299, 0.587, 0.114);
    let lumaTL = dot(luma, textureSample(colorTexture, colorSampler, tcoords + vec2f(-1.0, -1.0) * texCoordOffset).xyz);
    let lumaTR = dot(luma, textureSample(colorTexture, colorSampler, tcoords + vec2f( 1.0, -1.0) * texCoordOffset).xyz);
    let lumaBL = dot(luma, textureSample(colorTexture, colorSampler, tcoords + vec2f(-1.0,  1.0) * texCoordOffset).xyz);
    let lumaBR = dot(luma, textureSample(colorTexture, colorSampler, tcoords + vec2f( 1.0,  1.0) * texCoordOffset).xyz);
    let lumaM  = dot(luma, textureSample(colorTexture, colorSampler, tcoords).xyz);

    var dir: vec2f;
    dir.x = -((lumaTL + lumaTR) - (lumaBL + lumaBR));
    dir.y = ((lumaTL + lumaBL) - (lumaTR + lumaBR));

    let dirReduce = max((lumaTL + lumaTR + lumaBL + lumaBR) * (FXAA_REDUCE_MUL * 0.25), FXAA_REDUCE_MIN);
    let inverseDirAdjustment = 1.0 / (min(abs(dir.x), abs(dir.y)) + dirReduce);

    dir = min(vec2f(FXAA_SPAN_MAX, FXAA_SPAN_MAX),
        max(vec2f(-FXAA_SPAN_MAX, -FXAA_SPAN_MAX), dir * inverseDirAdjustment)) * texCoordOffset;

    let result1 = (1.0 / 2.0) * (
        textureSample(colorTexture, colorSampler, tcoords + dir * vec2f(1.0 / 3.0 - 0.5)).xyz +
        textureSample(colorTexture, colorSampler, tcoords + dir * vec2f(2.0 / 3.0 - 0.5)).xyz);

    let result2 = result1 * (1.0 / 2.0) + (1.0 / 4.0) * (
        textureSample(colorTexture, colorSampler, tcoords + dir * vec2f(0.0 / 3.0 - 0.5)).xyz +
        textureSample(colorTexture, colorSampler, tcoords + dir * vec2f(3.0 / 3.0 - 0.5)).xyz);

    let lumaMin = min(lumaM, min(min(lumaTL, lumaTR), min(lumaBL, lumaBR)));
    let lumaMax = max(lumaM, max(max(lumaTL, lumaTR), max(lumaBL, lumaBR)));
    let lumaResult2 = dot(luma, result2);

    if (lumaResult2 < lumaMin || lumaResult2 > lumaMax) {
        return result1;
    } else {
        return result2;
    }
}

@fragment
fn main(in: FragmentInput) -> @location(0) vec4f {
    var color: vec3f = fxaa(in.tcoords0.xy);
    let ssgi = textureSample(ssgiTexture, ssgiSampler, in.tcoords0.xy).rgb;
    color = color + ssgi;
    color = tonemapUncharted2(color);
    color = pow(color, vec3f(1.0 / 2.2));
    return vec4f(color, 1.0);
}
