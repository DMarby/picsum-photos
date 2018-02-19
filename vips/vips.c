#include "vips.h"

int saveImageToJpegBuffer(VipsImage *image, void **buf, size_t *len) {
  VipsArea *area = NULL;
  int result;

  // Progressive, strip metadata
  result = vips_call("jpegsave_buffer", image, &area, "interlace", TRUE, "strip", TRUE, "optimize_coding", TRUE, NULL);

  if (!result && area) {
    *buf = area->data;
    area->free_fn = NULL;
    *len = area->length;
    vips_area_unref(area);
  }

  return result;
}

int resize_image(void *buf, size_t len, VipsImage **out, int width, int height, VipsInteresting interesting) {
  VipsBlob *blob;
  int result;

  blob = vips_blob_new(NULL, buf, len);

  result = vips_call("thumbnail_buffer", blob, out, width, "height", height, "crop", interesting, NULL);

  vips_area_unref(VIPS_AREA(blob));

  return result;
}

int change_colorspace(VipsImage *in, VipsImage **out, VipsInterpretation colorspace) {
  return vips_call("colourspace", in, out, colorspace, NULL);
}
