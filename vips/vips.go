package vips

/*
#cgo pkg-config: vips
#include "vips.h"
*/
import "C"

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"
)

// Image is a representation of the *C.VipsImage type
type Image *C.VipsImage

var once sync.Once

// Initialize libvips if it's not already started
func Initialize() error {
	var err error

	once.Do(func() {
		// vips_init needs to run on the main thread
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		if C.VIPS_MAJOR_VERSION != 8 || C.VIPS_MINOR_VERSION < 6 {
			err = fmt.Errorf("unsupported libvips version")
		}

		errorCode := C.vips_init(C.CString("picsum-photos"))
		if errorCode != 0 {
			err = fmt.Errorf("unable to initialize vips: %v", catchVipsError())
			return
		}

		// Set concurrency to 1 so that each job only uses one thread
		C.vips_concurrency_set(1)
	})

	return err
}

// Shutdown libvips
func Shutdown() {
	C.vips_shutdown()
}

// PrintDebugInfo prints libvips debug info to stdout
func PrintDebugInfo() {
	C.im__print_all()
}

// catchVipsError returns the vips error buffer as an error
func catchVipsError() error {
	defer C.vips_error_clear()

	s := C.GoString(C.vips_error_buffer())
	return fmt.Errorf("%s", s)
}

// ResizeImage loads an image from a buffer and resizes it.
func ResizeImage(buffer []byte, width int, height int) (Image, error) {
	if len(buffer) == 0 {
		return nil, fmt.Errorf("empty buffer")
	}

	// Prevent buffer from being garbage collected
	defer runtime.KeepAlive(buffer)

	imageBuffer := unsafe.Pointer(&buffer[0])
	imageBufferSize := C.size_t(len(buffer))

	var image *C.VipsImage

	errCode := C.resize_image(imageBuffer, imageBufferSize, &image, C.int(width), C.int(height), C.VIPS_INTERESTING_CENTRE)

	if errCode != 0 {
		return nil, fmt.Errorf("error processing image from buffer %s", catchVipsError())
	}

	return image, nil
}

// SaveToBuffer saves an image to a buffer
func SaveToBuffer(image Image) ([]byte, error) {
	defer UnrefImage(image)

	var bufferPointer unsafe.Pointer
	bufferLength := C.size_t(0)

	err := C.saveImageToJpegBuffer(image, &bufferPointer, &bufferLength)

	if err != 0 {
		return nil, fmt.Errorf("error saving to buffer %s", catchVipsError())
	}

	buffer := C.GoBytes(bufferPointer, C.int(bufferLength))

	C.g_free(C.gpointer(bufferPointer))

	return buffer, nil
}

// Grayscale converts an image to grayscale
func Grayscale(image Image) (Image, error) {
	defer UnrefImage(image)

	var result *C.VipsImage

	err := C.change_colorspace(image, &result, C.VIPS_INTERPRETATION_B_W)

	if err != 0 {
		return nil, fmt.Errorf("error changing image colorspace %s", catchVipsError())
	}

	return result, nil
}

// Blur applies gaussian blur to an image
func Blur(image Image, blur int) (Image, error) {
	defer UnrefImage(image)

	var result *C.VipsImage

	err := C.blur_image(image, &result, C.double(blur))

	if err != 0 {
		return nil, fmt.Errorf("error applying blur to image %s", catchVipsError())
	}

	return result, nil
}

// UnrefImage unrefs an image object
func UnrefImage(image Image) {
	C.g_object_unref(C.gpointer(image))
}

// NewEmptyImage returns an empty image object
func NewEmptyImage() Image {
	return C.vips_image_new()
}
