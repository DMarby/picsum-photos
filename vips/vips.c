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

// TODO: Add options for crop strategy
int resize_image(void *buf, size_t len, VipsImage **out, int width, int height) {
	VipsBlob *blob;
	int result;

	blob = vips_blob_new(NULL, buf, len);

	result = vips_call("thumbnail_buffer", blob, out, width, "height", height, "crop", VIPS_INTERESTING_CENTRE, NULL);

	vips_area_unref(VIPS_AREA(blob));

	return result;
}
