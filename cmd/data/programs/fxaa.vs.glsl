#version 330 core

#define MAX_CASCADES 10
#define MAX_LIGHTS 16

// global uniforms
struct light {
    mat4 vpMatrix[MAX_CASCADES];
    vec4 zCuts[MAX_CASCADES];
    vec4 position;
    vec4 color;
};

layout (std140) uniform cameraConstants {
    mat4 vMatrix;
    mat4 pMatrix;
    mat4 vpMatrix;
    vec4 lightCount;
    light lights[MAX_LIGHTS];
};

// this is the same for all our models
layout (location =  0) in vec3 position_in;
layout (location =  1) in vec3 normal_in;
layout (location =  2) in vec3 tcoords0_in;
layout (location =  3) in mat4 mMatrix;
layout (location =  7) in mat4 mvpMatrix;
layout (location = 11) in vec4 custom1;
layout (location = 12) in vec4 custom2;
layout (location = 13) in vec4 custom3;
layout (location = 14) in vec4 custom4;

out vec3 tcoords0;

void main() {
    gl_Position = pMatrix * vec4(position_in, 1.0);
    tcoords0 = tcoords0_in;
}
