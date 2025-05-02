#include "animation.h"

#include <lvgl.h>
#include <zephyr/kernel.h>

void render_animation(lv_obj_t *widget, const struct nice_view_anim *anim) {
  lv_obj_t *art = lv_animimg_create(widget);
  lv_obj_center(art);
  lv_animimg_set_src(art, (const void **)anim->imgs, anim->len);
  lv_animimg_set_duration(art, anim->duration);
  lv_animimg_set_repeat_count(art, LV_ANIM_REPEAT_INFINITE);
  lv_animimg_start(art);
  lv_obj_align(art, LV_ALIGN_TOP_LEFT, 0, 0);
}
