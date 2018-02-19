#include "vips.h"

int saveImageToJpegBuffer(VipsImage *image, void **buf, size_t *len) {
  // Progressive, strip metadata
  return vips_jpegsave_buffer(image, buf, len, "interlace", TRUE, "strip", TRUE, "optimize_coding", TRUE, NULL);
}

int resize_image(void *buf, size_t len, VipsImage **out, int width, int height, VipsInteresting interesting) {
  return vips_thumbnail_buffer(buf, len, out, width, "height", height, "crop", interesting, NULL);
}

int change_colorspace(VipsImage *in, VipsImage **out, VipsInterpretation colorspace) {
  return vips_call("colourspace", in, out, colorspace, NULL);
}

int blur_image(VipsImage *in, VipsImage **out, double blur) {
  return vips_call("gaussblur", in, out, blur, NULL);
}
