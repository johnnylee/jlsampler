package jlsampler

// ----------------------------------------------------------------------------
type PlayingSample struct {
	controls *Controls // Controls!

	sample1 *Sample // Quieter sample being played.
	sample2 *Sample // Louder sample being played if mixLayers is true.
	mix     float32 // The mix parameter: (1-mix)*sample1 + mix*sample2.
	idx     float32 // Index being played in the given sample.
	idxMax  float32 // Total length of sample.
	amp1    float32 // Current amplification for sample 1.
	amp2    float32 // Current amplification for sample 2.
	pan     float32 // Pan of playing sample: -1 is hard left, 1 is hard right.
	tau     float32 // Decay constant (0 is disabled).

	fadeAmp float32 // Fade in amplification if controls.TauFadeIn > 0.
}

// NewPlayingSample:
// If not mixing layers, then mix should be 0, and amp2 is ignored.
func NewPlayingSample(
	controls *Controls,
	sample1, sample2 *Sample,
	amp1, amp2, pan, mix float32) *PlayingSample {

	ps := new(PlayingSample)
	ps.controls = controls
	ps.sample1 = sample1
	ps.sample2 = sample2
	ps.mix = mix
	ps.idxMax = float32(sample1.Len - 1)
	ps.amp1 = amp1
	ps.amp2 = amp2
	ps.pan = pan
	ps.tau = 0

	if controls.TauFadeIn != 0 {
		ps.fadeAmp = 1
	} else {
		ps.fadeAmp = 0
	}

	if sample2 != nil {
		ps.idx = float32(sample2.Idx0) - controls.NFadeIn
		if float32(sample2.Len-1) > ps.idxMax {
			ps.idxMax = float32(sample2.Len - 1)
		}
	} else {
		ps.idx = float32(sample1.Idx0) - controls.NFadeIn
	}

	if ps.idx < 0 {
		ps.idx = 0
	}

	return ps
}

// Add the current sample value to the buffer. Applying fades and panning.
func (ps *PlayingSample) addCurrentSample(buf *Sound, amp float32, i int) {
	L, R := ps.sample1.Interp(ps.idx)
	L *= ps.amp1
	R *= ps.amp1

	if ps.mix != 0 {
		L2, R2 := ps.sample2.Interp(ps.idx)
		L = L*(1-ps.mix) + L2*ps.amp2*ps.mix
		R = R*(1-ps.mix) + R2*ps.amp2*ps.mix
	}

	// Fade in.
	L *= (1 - ps.fadeAmp)
	R *= (1 - ps.fadeAmp)

	// Pan.
	if ps.pan < 0 {
		L -= ps.pan * R
		R *= 1 + ps.pan
	} else if ps.pan > 0 {
		R += ps.pan * L
		L *= 1 - ps.pan
	}

	buf.L[i] += amp * L
	buf.R[i] += amp * R
}

func (ps *PlayingSample) WriteOutput(buf *Sound, amp, di []float32) bool {
	for i, _ := range buf.L {
		// Update decay amplitude.
		if ps.tau != 0 {
			ps.amp1 *= ps.tau
			ps.amp2 *= ps.tau
			// Amplitude too low. Done playing this.
			if ps.amp1 < ampCutoff && ps.amp2 < ampCutoff {
				return false
			}
		}

		// Update fade in.
		if ps.fadeAmp != 0 {
			ps.fadeAmp *= float32(ps.controls.TauFadeIn)
			if ps.fadeAmp < ampCutoff {
				ps.fadeAmp = 0
			}
		}

		ps.addCurrentSample(buf, amp[i], i)

		// Update index.
		ps.idx += di[i]

		// Done playing?
		if ps.idx >= ps.idxMax {
			return false
		}
	}

	return true
}
