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

in vec3 position;
in vec3 cameraPosition;
in vec3 tcoords0;
in vec3 normal;

layout (location = 0) out vec4 color;

uniform sampler2D albedoTex;
uniform sampler2D normalTex;
uniform sampler2D roughTex;
uniform sampler2D metalTex;

uniform sampler2D shadowTex0;
uniform sampler2D shadowTex1;
uniform sampler2D shadowTex2;
uniform sampler2D shadowTex3;

float linstep(float low, float high, float v) {
    return clamp((v - low) / (high - low), 0.0, 1.0);
}

float varianceShadowMap(sampler2D shadowTex, vec2 coords, float compare) {
    vec2 moments = vec2(0.0);
    float tOffset = 1.0/2048.0;

    for (int i=2; i<5; i++) {
        moments += texture(shadowTex, coords.xy + tOffset * i * vec2(-1.0, -1.0)).xy;
        moments += texture(shadowTex, coords.xy + tOffset * i * vec2(+0.0, -1.0)).xy;
        moments += texture(shadowTex, coords.xy + tOffset * i * vec2(+1.0, -1.0)).xy;
        moments += texture(shadowTex, coords.xy + tOffset * i * vec2(-1.0, -0.0)).xy;
        moments += texture(shadowTex, coords.xy + tOffset * i * vec2(+1.0, -0.0)).xy;
        moments += texture(shadowTex, coords.xy + tOffset * i * vec2(-1.0, +1.0)).xy;
        moments += texture(shadowTex, coords.xy + tOffset * i * vec2(+0.0, +1.0)).xy;
        moments += texture(shadowTex, coords.xy + tOffset * i * vec2(+1.0, +1.0)).xy;
    }

    moments /= (8.0 * 3);

    float p = step(compare, moments.x);
    float variance = max(moments.y - moments.x * moments.x, 0.0000001);

    float d = compare - moments.x;
    float pMax = linstep(0.2, 1.0, variance / (variance + d*d));

    return min(max(p, pMax), 1.0);
}

float shadow(vec4 coords, int lightIndex) {
    const int numCascades = 3;

    vec3 shadowMapCoords[numCascades];

    for (int i=0; i<numCascades; i++) {
        vec4 ccoords = lights[lightIndex].vpMatrix[i] * coords;
        shadowMapCoords[i] = ccoords.xyz/ccoords.w;
    }

    float fragZV = length(cameraPosition-position);

    if (fragZV < lights[lightIndex].zCuts[0].x) {
        return varianceShadowMap(shadowTex0, shadowMapCoords[0].xy, shadowMapCoords[0].z);
    }

    if (fragZV < lights[lightIndex].zCuts[1].x) {
        return varianceShadowMap(shadowTex1, shadowMapCoords[1].xy, shadowMapCoords[1].z);
    }

    if (fragZV < lights[lightIndex].zCuts[2].x) {
        return varianceShadowMap(shadowTex2, shadowMapCoords[2].xy, shadowMapCoords[2].z);
    }

    return 1.0;
}

// beckmann
float distribution(vec3 n, vec3 h, float roughness) {
    float m_Sq = roughness * roughness;
    float NdotH_Sq = max(dot(n, h), 0.0);
    NdotH_Sq = NdotH_Sq * NdotH_Sq;
    return exp((NdotH_Sq - 1.0) / (m_Sq*NdotH_Sq)) / (3.14159265 * m_Sq * NdotH_Sq * NdotH_Sq);
}

// cook-torrance
float geometry(vec3 n, vec3 h, vec3 v, vec3 l, float roughness) {
    float NdotH = dot(n, h);
    float NdotL = dot(n, l);
    float NdotV = dot(n, v);
    float VdotH = dot(v, h);
    float NdotL_clamped = max(NdotL, 0.0);
    float NdotV_clamped = max(NdotV, 0.0);
    return min(min(2.0 * NdotH * NdotV_clamped / VdotH, 2.0 * NdotH * NdotL_clamped / VdotH), 1.0);
}

// schlich
float fresnel(float f0, vec3 n, vec3 l) {
    return f0 + (1.0 - f0) * pow(1.0 - dot(n, l), 5.0);
}

// fresnel diff
float diffuse_energy_ratio(float f0, vec3 n, vec3 l) {
    return 1.0 - fresnel(f0, n, l);
}

mat3 cotangent_frame(vec3 N, vec3 p, vec2 uv) {
    // get edge vectors of the pixel triangle
    vec3 dp1 = dFdx( p );
    vec3 dp2 = dFdy( p );
    vec2 duv1 = dFdx( uv );
    vec2 duv2 = dFdy( uv );

    // solve the linear system
    vec3 dp2perp = cross( dp2, N );
    vec3 dp1perp = cross( N, dp1 );
    vec3 T = dp2perp * duv1.x + dp1perp * duv2.x;
    vec3 B = dp2perp * duv1.y + dp1perp * duv2.y;

    // construct a scale-invariant frame
    float invmax = inversesqrt( max( dot(T,T), dot(B,B) ) );
    return mat3( T * invmax, B * invmax, N );
}

vec3 perturb_normal(vec3 N, vec3 V, vec2 texcoord) {
    // assume N, the interpolated vertex normal and
    // V, the view vector (vertex to eye)
    vec3 map = texture(normalTex, texcoord).xyz * 2.0 - 1.0;
    mat3 TBN = cotangent_frame(N, -V, texcoord);
    return normalize(TBN * map);
}

void main() {
    // init
    color = vec4(0.0);

    // init materials
    vec4 albedo = texture(albedoTex, tcoords0.st);
    float metalness = texture(metalTex, tcoords0.st).r;
    float roughness = 0.1 + 0.8 * texture(roughTex, tcoords0.st).r;

    // adjust f0 from 0.118 to 0.818, this will normally be discrete
    float f0 = 0.118 + metalness * 0.7; //max 0.818

    // normal, eye/view
    vec3 V = normalize(cameraPosition - position);
    vec3 N = perturb_normal(normal, V, tcoords0.st);

    // shared products
    float NdotV = dot(N, V);
    float NdotV_clamped = max(NdotV, 0.0000000001);

    int lc = int(lightCount[0]);
    for (int i=0; i<lc; i++) {
        // lightdir, halfvec
        vec3 L = normalize(lights[i].position.xyz - position);
        vec3 H = normalize(L + V);

        float NdotL = dot(N, L);
        float NdotL_clamped = max(NdotL, 0.0);

        float fres = fresnel(f0, H, L);
        float geom = geometry(N, H, V, L, roughness);
        float ndf = distribution(N, H,  roughness);

        float brdf_spec = (0.25 * fres * geom * ndf) / (NdotL_clamped * NdotV_clamped);
        if (NdotV <= 0.0 || NdotL <= 0.0) {
            brdf_spec = 0.0;
        }

        vec3 color_spec = NdotL_clamped * brdf_spec * (lights[i].color.rgb*(1.0-metalness) + albedo.rgb*metalness);
        vec3 color_diff = NdotL_clamped * diffuse_energy_ratio(f0, N, L) * albedo.rgb * lights[i].color.rgb;
        float sh = shadow(vec4(position, 1.0), i);
        color.rgb += (color_diff + color_spec) * sh;
    }

    color.a = albedo.a;
}