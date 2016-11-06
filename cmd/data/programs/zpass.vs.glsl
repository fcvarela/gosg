#version 410 core

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
    dmat4 dvMatrix;
    dmat4 dpMatrix;
    dmat4 dvpMatrix;
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
layout (location =  7) in vec4 custom1;
layout (location =  8) in vec4 custom2;
layout (location =  9) in vec4 custom3;
layout (location = 10) in vec4 custom4;

void main() {
    gl_Position = vpMatrix * mMatrix * vec4(position_in, 1.0);
}
