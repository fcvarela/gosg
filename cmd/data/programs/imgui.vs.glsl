#version 410 core

// global uniforms
struct light {
    mat4 vpMatrix;
    vec4 position;
    vec4 color;
};

layout (std140) uniform cameraConstants {
    mat4 vMatrix;
    mat4 pMatrix;
    mat4 vpMatrix;
    vec4 lightCount;
    light lights[16];
};

layout (location = 0) in vec2 position_in;
layout (location = 1) in vec2 tcoords0_in;
layout (location = 2) in vec4 color_in;

out vec2 Frag_UV;
out vec4 Frag_Color;

void main() {
    Frag_UV = tcoords0_in;
    Frag_Color = color_in;

    gl_Position = pMatrix * vec4(position_in, 0.0, 1.0);
}
