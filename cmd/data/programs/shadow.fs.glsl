#version 410 core

layout (location = 0) out vec4 color;

void main() {
    float depth = gl_FragCoord.z;
    float dx = dFdx(depth);
    float dy = dFdy(depth);
    float moment2 = depth * depth + 0.25 * (dx * dx + dy * dy);
    color = vec4(depth, moment2, 0.0, 1.0);
}
