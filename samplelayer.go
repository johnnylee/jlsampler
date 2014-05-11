package jlsampler

// ----------------------------------------------------------------------------
type SampleLayer struct {
	samples []*Sample
	idx     int
}

func (sl *SampleLayer) NumSamples() int {
	return len(sl.samples)
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
	for _, s := range sl2.samples {
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

func (sl *SampleLayer) UpdateCropThresh(thresh float64) {
	for _, sample := range sl.samples {
		sample.UpdateCropThresh(thresh)
	}
}

func (sl *SampleLayer) UpdateRms(rmsTime float64) {
	for _, sample := range sl.samples {
		sample.UpdateRms(rmsTime)
	}
}
