package jlsampler

import (
	"fmt"
	"github.com/johnnylee/flac"
	"os"
)

// ----------------------------------------------------------------------------
// Global constants.
const (
	version    = 0.90    // Version number.
	sampleRate = 48000   // Fixed sample rate.
	maxVal16   = 32767   // 16-bit maximum sample value.
	maxVal24   = 8388607 // 24-bit maximum sample value.
	ampCutoff  = 1e-5    // Cut-off amplitude of decaying sample.
)

// ----------------------------------------------------------------------------
func LoadFlac(path string) (*Sample, error) {
	L, R, err := flac.LoadInt16(path)
	if err != nil {
		return nil, err
	}
	return NewSampleFromArrays(L, R), nil
}

// ----------------------------------------------------------------------------
func Println(a ...interface{}) {
	os.Stderr.Write([]byte(fmt.Sprintln(a...)))
}
