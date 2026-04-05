// imgui fragment shader — sample texture * vertex color

@group(1) @binding(0) var colorTex: texture_2d<f32>;
@group(1) @binding(1) var colorSampler: sampler;

struct FragmentInput {
    @location(0) fragUV: vec2f,
    @location(1) fragColor: vec4f,
};

@fragment
fn main(in: FragmentInput) -> @location(0) vec4f {
    return in.fragColor * textureSample(colorTex, colorSampler, in.fragUV);
}
