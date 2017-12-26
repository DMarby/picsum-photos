#include "vips.h"

int loadImageFromBuffer(char *operation_name, void *buf, size_t len, VipsImage **out) {
	VipsBlob *blob;
	int result;

	// TODO: Is this needed?
	blob = vips_blob_new(NULL, buf, len);

	result = vips_call(operation_name, blob, out, "fail", TRUE, NULL); // TODO: Does this get cached?

	vips_area_unref(VIPS_AREA(blob));

	return result;
}

// TODO: Make this support optional args (per format?)
int saveImageToBuffer(char *operation_name, VipsImage *image, void **buf, size_t *len) {
	VipsArea *area = NULL;
	int result;

	// TODO: Progressive?
	result = vips_call(operation_name, image, &area, NULL);

	if (!result && area) {
		*buf = area->data;
		area->free_fn = NULL;
		*len = area->length;
		vips_area_unref(area);
	}

	return result;
}

// TODO: Remove, just a test
int invert_image(VipsImage *in, VipsImage **out) {
	return vips_invert(in, out, NULL);
}
