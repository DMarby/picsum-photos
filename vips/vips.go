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

	// Vips cannot be initialized after it's been shut down, thus we don't set initialize to false
	if !initialized {
		return
	}

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

func loadFromBuffer(buffer []byte) (*C.VipsImage, error) {
	// Prevent buffer from being garbage collected
	// TODO: Do we need to do anything to clean up? Copy instead?
	defer runtime.KeepAlive(buffer)

	imageBuffer := unsafe.Pointer(&buffer[0])
	imageBufferSize := C.size_t(len(buffer))

	loader := C.vips_foreign_find_load_buffer(imageBuffer, imageBufferSize)

	if loader == nil {
		return nil, fmt.Errorf("error finding image type %v", catchVipsError())
	}

	var image *C.VipsImage

	err := C.loadImageFromBuffer(loader, imageBuffer, imageBufferSize, &image)

	if err != 0 {
		return nil, fmt.Errorf("error loading image from buffer %v", catchVipsError())
	}

	return image, nil
}

func saveToBuffer(image *C.VipsImage) ([]byte, error) {
	defer C.g_object_unref(C.gpointer(image))

	var bufferPointer unsafe.Pointer
	bufferLength := C.size_t(0)

	err := C.saveImageToBuffer(C.CString("jpegsave_buffer"), image, &bufferPointer, &bufferLength)

	if err != 0 {
		return nil, fmt.Errorf("error saving to buffer %v", catchVipsError())
	}

	buffer := C.GoBytes(bufferPointer, C.int(bufferLength))

	C.g_free(C.gpointer(bufferPointer))

	return buffer, nil
}

// ProcessImage performs the specified image operations and returns the resulting image
func ProcessImage(image []byte) ([]byte, error) {
	defer C.vips_thread_shutdown()

	vipsImage, err := loadFromBuffer(image)
	if err != nil {
		return nil, err
	}

	var invertedImage *C.VipsImage

	invertErr := C.invert_image(vipsImage, &invertedImage)
	if invertErr != 0 {
		return nil, fmt.Errorf("error inverting image %v", catchVipsError())
	}

	buffer, err := saveToBuffer(invertedImage)
	if err != nil {
		return nil, err
	}

	return buffer, nil
}

func freeCString(s *C.char) {
	C.free(unsafe.Pointer(s))
}
