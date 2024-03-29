#include <stdlib.h>
#include <vips/vips.h>
#include <vips/foreign.h>
#include <vips/vector.h>

// Require libvips 8 at compile time
#if (VIPS_MAJOR_VERSION != 8 || VIPS_MINOR_VERSION < 6)
  #error "unsupported libvips version"
#endif


void setup_logging();
void log_handler(char const* log_domain, GLogLevelFlags log_level, char const* message, void* ignore);
extern void log_callback(char* message);

int save_image_to_jpeg_buffer(VipsImage *image, void **buf, size_t *len);
int save_image_to_webp_buffer(VipsImage *image, void **buf, size_t *len);
int resize_image(void *buf, size_t len, VipsImage **out, int width, int height, VipsInteresting interesting);
int change_colorspace(VipsImage *in, VipsImage **out, VipsInterpretation colorspace);
int blur_image(VipsImage *in, VipsImage **out, double blur);
void set_user_comment(VipsImage *image, char const* comment);
