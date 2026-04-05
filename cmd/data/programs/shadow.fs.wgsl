// shadow fragment shader — output depth and moment2 for VSM

@fragment
fn main(@builtin(position) fragCoord: vec4f) -> @location(0) vec4f {
    let depth = fragCoord.z;
    let dx = dpdx(depth);
    let dy = dpdy(depth);
    let moment2 = depth * depth + 0.25 * (dx * dx + dy * dy);
    return vec4f(depth, moment2, 0.0, 1.0);
}
