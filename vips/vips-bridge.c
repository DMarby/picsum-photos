#include "vips-bridge.h"

void setup_logging() {
  g_log_set_handler("VIPS", G_LOG_LEVEL_WARNING, log_handler, NULL);
}

void log_handler(char const* log_domain, GLogLevelFlags log_level, char const* message, void* ignore) {
  log_callback((char*)message);
}

int save_image_to_jpeg_buffer(VipsImage *image, void **buf, size_t *len) {
  return vips_jpegsave_buffer(image, buf, len, "interlace", TRUE, "optimize_coding", TRUE, NULL);
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

static void * remove_metadata(VipsImage *image, const char *field, GValue *value, void *my_data) {
	if (vips_isprefix("exif-", field)) {
    vips_image_remove(image, field);
  }

	return (NULL);
}

void set_user_comment(VipsImage *image, char const* comment) {
  // Strip all the metadata
  vips_image_remove(image, VIPS_META_EXIF_NAME);
  vips_image_remove(image, VIPS_META_XMP_NAME);
  vips_image_remove(image, VIPS_META_IPTC_NAME);
  vips_image_remove(image, VIPS_META_ICC_NAME);
  vips_image_remove(image, VIPS_META_ORIENTATION);
  vips_image_remove(image, "jpeg-thumbnail-data");
  vips_image_map(image, remove_metadata, NULL);

  // Set the user comment
  vips_image_set_string(image, "exif-ifd2-UserComment", comment);
}
