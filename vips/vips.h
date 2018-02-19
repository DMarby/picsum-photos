#include <stdlib.h>
#include <vips/vips.h>
#include <vips/foreign.h>

// Require libvips 8 at compile time
#if (VIPS_MAJOR_VERSION != 8 || VIPS_MINOR_VERSION < 6)
  #error "unsupported libvips version"
#endif

int saveImageToJpegBuffer(VipsImage *image, void **buf, size_t *len);
int resize_image(void *buf, size_t len, VipsImage **out, int width, int height);
