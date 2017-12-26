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

var (
	mutex       sync.Mutex
	initialized bool
)

// Initialize libvips if it's not already started
func Initialize() error {
	// Ensure that this doesn't run concurrenctly
	mutex.Lock()
	defer mutex.Unlock()

	if initialized {
		return nil
	}

	// vips_init needs to run on the main thread
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if C.VIPS_MAJOR_VERSION != 8 || C.VIPS_MINOR_VERSION < 6 {
		return fmt.Errorf("unsupported libvips version")
	}

	err := C.vips_init(C.CString("picsum-photos"))
	if err != 0 {
		return fmt.Errorf("unable to initialize vips: %v", catchVipsError())
	}

	// TODO: Do we want this?
	//C.vips_cache_set_max_mem(maxCacheMem)
	//C.vips_cache_set_max(maxCacheSize)

	// Set concurrency to 1 so that each job only uses one thread
	C.vips_concurrency_set(1)

	initialized = true
	return nil
}

// Shutdown libvips if it's initialized
func Shutdown() {
	// Ensure that this doesn't run concurrenctly
	mutex.Lock()
	defer mutex.Unlock()

	if !initialized {
		return
	}

	C.vips_shutdown()
	initialized = false
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

// TODO: Vips thread shutdown?
// TODO: send autorotate param to load?
func LoadFromBuffer(buffer []byte) (*C.VipsImage, error) { // TODO: Wrap this in something so that we don't expose VipsImage? Or do that in image package instead?
	// Prevent buffer from being garbage collected
	// TODO: Do we need to do anything to clean up? Copy instead?
	defer runtime.KeepAlive(buffer)

	imageBuffer := unsafe.Pointer(&buffer[0])
	imageBufferSize := C.size_t(len(buffer))

	// TODO: Validate against what loaders we have? Needed?
	// TODO: If return is NULL then error
	loader := C.vips_foreign_find_load_buffer(imageBuffer, imageBufferSize)

	var image *C.VipsImage

	err := C.loadImageFromBuffer(loader, imageBuffer, imageBufferSize, &image)

	if err != 0 {
		return nil, fmt.Errorf("error loading image from buffer %v", catchVipsError())
	}

	fmt.Printf("Code %v, width %v", err, C.vips_image_get_width(image))

	var invertedImage *C.VipsImage

	err = C.invert_image(image, &invertedImage)

	if err != 0 {
		return nil, fmt.Errorf("error inverting image %v", catchVipsError())
	}

	return invertedImage, nil
}

// TODO: Support other formats
func SaveToBuffer(image *C.VipsImage) ([]byte, error) {
	defer C.g_object_unref(C.gpointer(image)) // TODO: What if we want to use it more?

	var bufferPointer unsafe.Pointer
	bufferLength := C.size_t(0)

	err := C.saveImageToBuffer(C.CString("jpegsave_buffer"), image, &bufferPointer, &bufferLength)

	if err != 0 {
		return nil, fmt.Errorf("error saving to buffer %v", catchVipsError())
	}

	buffer := C.GoBytes(bufferPointer, C.int(bufferLength))

	// TODO: Do we need additional cleanup? vips_error_clear?
	C.g_free(C.gpointer(bufferPointer))

	return buffer, nil
}

func freeCString(s *C.char) {
	C.free(unsafe.Pointer(s))
}
