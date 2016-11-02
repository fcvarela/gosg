#version 410 core

uniform sampler2D colorTex;

in vec2 Frag_UV;
in vec4 Frag_Color;

out vec4 Out_Color;

void main() {
    Out_Color = Frag_Color * texture(colorTex, Frag_UV.st);
}
