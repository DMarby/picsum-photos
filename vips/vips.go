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

		// TODO: Do we want this?
		//C.vips_cache_set_max_mem(maxCacheMem)
		//C.vips_cache_set_max(maxCacheSize)

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
	defer C.vips_thread_shutdown()
	defer C.vips_error_clear()

	s := C.GoString(C.vips_error_buffer())
	return fmt.Errorf("%s", s)
}

// ResizeImage loads an image from a buffer and resizes it.
func ResizeImage(buffer []byte, width int, height int) (*C.VipsImage, error) {
	// Prevent buffer from being garbage collected
	defer runtime.KeepAlive(buffer)

	imageBuffer := unsafe.Pointer(&buffer[0])
	imageBufferSize := C.size_t(len(buffer))

	var image *C.VipsImage

	errCode := C.resize_image(imageBuffer, imageBufferSize, &image, C.int(width), C.int(height), C.VIPS_INTERESTING_CENTRE)

	if errCode != 0 {
		return nil, fmt.Errorf("error processing image from buffer %v", catchVipsError())
	}

	return image, nil
}

// SaveToBuffer saves an image to a buffer
func SaveToBuffer(image *C.VipsImage) ([]byte, error) {
	defer C.g_object_unref(C.gpointer(image))

	var bufferPointer unsafe.Pointer
	bufferLength := C.size_t(0)

	err := C.saveImageToJpegBuffer(image, &bufferPointer, &bufferLength)

	if err != 0 {
		return nil, fmt.Errorf("error saving to buffer %v", catchVipsError())
	}

	buffer := C.GoBytes(bufferPointer, C.int(bufferLength))

	C.g_free(C.gpointer(bufferPointer))

	return buffer, nil
}

// Grayscale converts an image to grayscale
func Grayscale(image *C.VipsImage) (*C.VipsImage, error) {
	defer C.g_object_unref(C.gpointer(image))

	var result *C.VipsImage

	err := C.change_colorspace(image, &result, C.VIPS_INTERPRETATION_B_W)

	if err != 0 {
		return nil, fmt.Errorf("error changing image colorspace %v", catchVipsError())
	}

	return result, nil
}

// Blur applies gaussian blur to an image
func Blur(image *C.VipsImage, blur int) (*C.VipsImage, error) {
	defer C.g_object_unref(C.gpointer(image))

	var result *C.VipsImage

	err := C.blur_image(image, &result, C.double(blur))

	if err != 0 {
		return nil, fmt.Errorf("error applying blur to image %v", catchVipsError())
	}

	return result, nil
}
