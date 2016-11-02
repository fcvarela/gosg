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

// this is the same for all our models
layout (location = 0) in vec2 position_in;
layout (location = 1) in vec3 normal_in;
layout (location = 2) in vec3 tcoords0_in;

out vec3 tcoords0;

void main() {
    gl_Position = pMatrix * vec4(position_in, 0.0, 1.0);
    tcoords0 = tcoords0_in;
}
