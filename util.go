package jlsampler

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unsafe"
)

import "C"

// ----------------------------------------------------------------------------
func Println(a ...interface{}) {
	os.Stderr.Write([]byte(fmt.Sprintln(a...)))
}

func Printf(format string, a ...interface{}) {
	os.Stderr.Write([]byte(fmt.Sprintf(format, a...)))
}

func Clamp16(val float64) float64 {
	if val > maxVal16 {
		return maxVal16
	} else if val < -maxVal16 {
		return -maxVal16
	}
	return val
}

// ----------------------------------------------------------------------------
func cArrayToSlice16(cArray *C.int16_t, length int) []int16 {
	var goSlice []int16
	sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&goSlice)))
	sliceHeader.Cap = length
	sliceHeader.Len = length
	sliceHeader.Data = uintptr(unsafe.Pointer(cArray))
	return goSlice
}

func cArrayToSlice32f(cArray *C.float, length int) []float32 {
	var goSlice []float32
	sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&goSlice)))
	sliceHeader.Cap = length
	sliceHeader.Len = length
	sliceHeader.Data = uintptr(unsafe.Pointer(cArray))
	return goSlice
}

// ----------------------------------------------------------------------------
// Return key, layer, variation.
func samplePathInfo(path string) (int, int, int) {
	items := strings.Split(path, "-")
	if len(items) < 4 {
		return 0, 0, 0
	}

	key, _ := strconv.Atoi(items[1])
	layer, _ := strconv.Atoi(items[2])
	variation, _ := strconv.Atoi(items[3])

	return key, layer - 1, variation
}

// Return number of layers and slice of samples paths.
func samplePaths(key int) (int, []string) {
	path := fmt.Sprintf("samples/on-%03d-*", key)
	paths, err := filepath.Glob(path)
	if err != nil {
		Println("No samples for key:", key, err)
		return 0, paths
	}

	sort.Strings(paths)

	layers := 0
	for _, path := range paths {
		layer, _ := strconv.Atoi(strings.Split(path, "-")[2])
		if layer > layers {
			layers = layer
		}
	}
	return layers, paths
}

// ----------------------------------------------------------------------------
// 
func FileExists(path string) bool {
	 _, err := os.Stat(path)
	return err == nil
}

