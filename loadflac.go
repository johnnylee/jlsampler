package jlsampler

import (
	"github.com/johnnylee/flac"
)

func LoadFlac(path string) (*Sample, error) {
	L, R, err := flac.LoadInt16(path)
	if err != nil {
		return nil, err
	}
	return NewSampleFromArrays(L, R), nil 
}
