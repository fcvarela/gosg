//
// Created by Filipe Varela on 03/01/2016.
//

#include <stdio.h>
#include <stdlib.h>
#include "imgui.h"

#ifdef __cplusplus
extern "C" {
#endif

#include "gosg_imgui.h"

#define max(a,b) a>b?a:b
#define min(a,b) a<b?a:b

void set_dt(double dt) {
    ImGuiIO &io = ImGui::GetIO();
    io.DeltaTime = float(dt);
}

int wants_capture_mouse() {
    ImGuiIO &io = ImGui::GetIO();
    return io.WantCaptureMouse;
}

int wants_capture_keyboard() {
    ImGuiIO &io = ImGui::GetIO();
    return io.WantCaptureKeyboard;
}

void set_display_size(float x, float y) {
    ImGuiIO &io = ImGui::GetIO();
    io.DisplaySize.x = x;
    io.DisplaySize.y = y;
    io.RenderDrawListsFn = NULL;
    io.IniFilename = NULL;
}

void set_texture_id(void *texture) {
    ImGuiIO &io = ImGui::GetIO();
    io.Fonts->SetTexID(texture);
}

void set_mouse_position(double x, double y) {
    ImGuiIO &io = ImGui::GetIO();
    io.MousePos.x = x;
    io.MousePos.y = y;
}


void set_mouse_scroll_position(double xoffset, double yoffset) {
    ImGuiIO &io = ImGui::GetIO();
    io.MouseWheel += yoffset;
}

void set_mouse_buttons(int b0, int b1, int b2) {
    ImGuiIO &io = ImGui::GetIO();
    io.MouseDown[0] = b0;
    io.MouseDown[1] = b1;
    io.MouseDown[2] = b2;
}

unsigned char *get_texture_data(int *width, int *height) {
    ImGuiStyle& style = ImGui::GetStyle();
    style.WindowRounding = 0.0;

    ImGuiIO &io = ImGui::GetIO();
    unsigned char *pixels;
    ImGui::GetIO().Fonts->GetTexDataAsRGBA32(&pixels, width, height);
    return pixels;
}

void frame_new() {
    ImGuiIO &io = ImGui::GetIO();
    io.MouseDrawCursor = true;
    ImGui::NewFrame();
}

void render() {
    ImGui::Render();
}

void *get_draw_data() {
    return (void *) ImGui::GetDrawData();
}

int get_cmdlist_count(void *drawdata) {
    ImDrawData *data = (ImDrawData *)drawdata;
    return data->CmdListsCount;
}

cmdlist_t get_cmdlist(void *drawdata, int index) {
    ImDrawData *data = (ImDrawData *)drawdata;

    cmdlist_t out_cmdlist;
    ImDrawList *cmd_list = data->CmdLists[index];

    out_cmdlist.commandBufferSize = cmd_list->CmdBuffer.size();
    out_cmdlist.vertexBufferSize = cmd_list->VtxBuffer.size();
    out_cmdlist.indexBufferSize = cmd_list->IdxBuffer.size();

    out_cmdlist.vertexPointer = (float *)&cmd_list->VtxBuffer.front();
    out_cmdlist.indexPointer = (unsigned short *)&cmd_list->IdxBuffer.front();

    return out_cmdlist;
}

void *get_cmdlist_cmd(void *drawdata, int index, int cmdindex, int *elementCount, float *clipRect) {
    ImDrawData *data = (ImDrawData *)drawdata;
    ImDrawList *cmd_list = data->CmdLists[index];
    ImDrawCmd pcmd = cmd_list->CmdBuffer[cmdindex];

    *elementCount = pcmd.ElemCount;
    clipRect[0] = pcmd.ClipRect.x;
    clipRect[1] = pcmd.ClipRect.y;
    clipRect[2] = pcmd.ClipRect.z;
    clipRect[3] = pcmd.ClipRect.w;
    
    return pcmd.TextureId;
}

int begin(const char* name, int flags) {
    return ImGui::Begin(name, NULL, flags);
}

void end() {
    ImGui::End();
}

void plot_histogram(const char* label, const float* values, int values_count, float scale_min, float scale_max, const float *graph_size) {
    ImGui::PlotHistogram(label, values, values_count, 0, NULL, scale_min, scale_max, ImVec2(graph_size[0], graph_size[1]), sizeof(float));
}

int collapsing_header(const char* name) {
    return ImGui::CollapsingHeader(name);
}

void image(void *texture, float *size) { 
    ImGui::Image(texture, ImVec2(size[0], size[1]));
}

void text(const char *t) {
    ImGui::Text("%s", t);
}

void set_next_window_pos(float x, float y) {
    ImGui::SetNextWindowPos(ImVec2(x, y));
}

void set_next_window_size(float x, float y) {
    ImGui::SetNextWindowSize(ImVec2(x, y));
}

#ifdef __cplusplus
}
#endif
