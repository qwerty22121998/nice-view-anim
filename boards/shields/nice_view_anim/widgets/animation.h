#include <lvgl.h>


struct nice_view_anim {
    uint8_t len;
    uint32_t duration;
    const lv_img_dsc_t **imgs;
};

void render_animation(lv_obj_t *widget, const struct nice_view_anim *anim);