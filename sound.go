package jlsampler

// ----------------------------------------------------------------------------
type Sound struct {
	Len int       // The length of the sound.
	L   []float32 // Left channel.
	R   []float32 // Right channel.
}

func NewSound(size int) *Sound {
	s := new(Sound)
	s.Len = size
	s.L = make([]float32, size)
	s.R = make([]float32, size)
	return s
}

func NewSoundFromSample(samp *Sample) *Sound {
	s := NewSound(samp.Len)
	for i := 0; i < samp.Len; i++ {
		s.L[i] = float32(samp.L[i]) / maxVal16
		s.R[i] = float32(samp.R[i]) / maxVal16
	}
	return s
}

func (s *Sound) Zero() {
	for i := 0; i < s.Len; i++ {
		s.L[i] = 0
		s.R[i] = 0
	}
}
