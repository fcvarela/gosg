#version 410 core

#define MAX_CASCADES 10

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
    light lights[16];
};

// this is the same for all our models
layout (location = 0) in vec3 position_in;
layout (location = 1) in vec3 normal_in;
layout (location = 2) in vec3 tcoords0_in;
layout (location = 3) in mat4 mMatrix;

out vec3 position;
out vec3 normal;
out vec3 cameraPosition;
out vec3 tcoords0;

void main() {
    // clip position
    gl_Position = vpMatrix * mMatrix * vec4(position_in, 1.0);

    // world position & camera world position
    position = (mMatrix * vec4(position_in, 1.0)).xyz;
    cameraPosition = inverse(vMatrix)[3].rgb;

    // world TBN
    normal = normalize((mMatrix * vec4(normal_in, 0.0)).xyz);

    // texture coordinates
    tcoords0 = tcoords0_in;
}
