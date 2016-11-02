//
// Created by Filipe Varela on 03/01/2016.
//

typedef struct {
    int commandBufferSize;
    int vertexBufferSize;
    int indexBufferSize;
    unsigned short *indexPointer;
    float *vertexPointer;
} cmdlist_t;

// windowing calls
void set_dt(double dt);
void set_display_size(float x, float y);
void set_texture_id(void *texture);
void set_mouse_position(double x, double y);
void set_mouse_buttons(int b0, int b1, int b2);
void set_mouse_scroll_position(double xoffset, double yoffset);

int wants_capture_mouse();
int wants_capture_keyboard();

// draw loop
void frame_new();
void render();

// renderer calls
unsigned char *get_texture_data(int *width, int *height);
void *get_draw_data();
int get_cmdlist_count(void *drawdata);
cmdlist_t get_cmdlist(void *drawdata, int index);
void *get_cmdlist_cmd(void *drawdata, int index, int cmdindex, int *elementCount, float *clipRect);

// passthrough
int begin(const char* name, int flags);
void end();
void plot_histogram(const char* label, const float* values, int values_count, float scale_min, float scale_max, const float *graph_size);
int collapsing_header(const char* name);
void image(void *texture, float *size);
void text(const char *t);

void set_next_window_pos(float x, float y);
void set_next_window_size(float x, float y);

