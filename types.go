package jlsampler

// ----------------------------------------------------------------------------
type SampleLayer struct {
	samples []*Sample
	idx     int
}

func (sl *SampleLayer) Copy() *SampleLayer {
	sl2 := new(SampleLayer)
	for _, s := range sl.samples {
		sl2.AddSample(s)
	}
	return sl2
}

func (sl *SampleLayer) AddSample(sample *Sample) {
	sl.samples = append(sl.samples, sample)
}

func (sl *SampleLayer) BorrowFrom(sl2 *SampleLayer, semitones int) {
	for _, s := range(sl2.samples) {
		sl.AddSample(s.Stretched(float64(semitones)))
	}
}

// GetSample: Get a sample from the layer. If idx is -1, then get the next
// round-robbin sample. Otherwise return the indicated sample. 
// Return: The sample and the sample's index. 
func (sl *SampleLayer) GetSample(idx int) (int, *Sample) {
	if idx == -1 {
		sl.idx = (sl.idx + 1) % len(sl.samples)
		return sl.idx, sl.samples[sl.idx]
	} else {
		return idx, sl.samples[idx]
	}
}

func (sl *SampleLayer) Transpose(trans int) *SampleLayer {
	sl2 := new(SampleLayer)
	for _, sample := range sl.samples {
		sl2.AddSample(sample.Stretched(float64(trans)))
	}
	return sl2
}

func (sl *SampleLayer) UpdateCropThresh() {
	for _, sample := range sl.samples {
		sample.UpdateCropThresh()
	}
}

func (sl *SampleLayer) UpdateRms() {
	for _, sample := range sl.samples {
		sample.UpdateRms()
	}
}

// ----------------------------------------------------------------------------
type Sound struct {
	T0  int64     // Creation time in nanoseconds.
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

