package jlsampler

import (
	"math"
)

// ----------------------------------------------------------------------------
func absMaxSampleVal(s *Sample, idx int) float64 {
	var val, max float64

	val = math.Abs(float64(s.L[idx]))
	if val > max {
		max = val
	}

	val = math.Abs(float64(s.R[idx]))
	if val > max {
		max = val
	}

	return max / maxVal16
}

// See http://paulbourke.net/miscellaneous/interpolation/.
// Using nearest neighbor instead of linear interpolation sounds aweful.
// After trying several interpolation methods, inclusing cubic, cosine,
// and hermite polynomials, I've found linear interpolation to sound the best
// for the small rate changes we're making.
//
// Of course it's always possible that I had an error somewhere.
func InterpLinear16(y1, y2, mu float64) int16 {
	y := y1*(1-mu) + y2*mu
	if y > maxVal16 {
		return maxVal16
	} else if y < -maxVal16 {
		return -maxVal16
	} else {
		return int16(y)
	}
}

// ----------------------------------------------------------------------------
type Sample struct {
	Rms  float64 // The RMS value of the initial samples.
	Idx0 int     // Zero index.
	Len  int     // Number of samples in each channel.
	L    []int16 // Left channel samples.
	R    []int16 // Right channel samples.
}

func NewSample(size int) *Sample {
	s := new(Sample)
	s.Len = size
	s.L = make([]int16, size)
	s.R = make([]int16, size)
	return s
}

func NewSampleFromArrays(L, R []int16) *Sample {
	s := new(Sample)
	s.Len = len(L)
	s.L = L
	s.R = R
	return s
}

func (s *Sample) Stretched(semitones float64) *Sample {
	if semitones == 0 {
		return s
	}

	ratio := math.Pow(2.0, -semitones/12.0)
	newLen := int(float64(s.Len-1) * ratio)

	sNew := NewSample(newLen)

	for i := 0; i < newLen; i++ {
		jf := float64(i) / ratio
		j := int(jf)
		mu := jf - float64(j)
		sNew.L[i] = InterpLinear16(float64(s.L[j]), float64(s.L[j+1]), mu)
		sNew.R[i] = InterpLinear16(float64(s.R[j]), float64(s.R[j+1]), mu)
	}

	return sNew
}

func (s *Sample) UpdateCropThresh(thresh float64) {
	for s.Idx0 = 0; s.Idx0 < s.Len; s.Idx0++ {
		if absMaxSampleVal(s, s.Idx0) >= thresh {
			return
		}
	}
}

func (s *Sample) UpdateRms(rmsTime float64) {
	rms := 0.0
	num := 0.0
	var x float64

	imin := s.Idx0
	imax := s.Idx0 + int(sampleRate*rmsTime)
	if imax > s.Len {
		imax = s.Len
	}

	for i := imin; i < imax; i++ {
		x = float64(s.L[i]) / maxVal16
		rms += x * x

		x = float64(s.R[i]) / maxVal16
		rms += x * x

		num += 2
	}

	rms /= num
	s.Rms = math.Sqrt(rms)
}

// Return interpolated L and R samples for the given index.
// Samples are scaled to 1.0 max.
func (s *Sample) Interp(idx float32) (float32, float32) {
	iIdx := int64(idx)
	mu := idx - float32(iIdx)

	L := (float32(s.L[iIdx])*(1-mu) + float32(s.L[iIdx+1])*mu) / maxVal16
	R := (float32(s.R[iIdx])*(1-mu) + float32(s.R[iIdx+1])*mu) / maxVal16

	return L, R
}
