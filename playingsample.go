package jlsampler

import (
	"math"
)

// ----------------------------------------------------------------------------
type PlayingSample struct {
	sample1 *Sample // Quieter sample being played.
	sample2 *Sample // Louder sample being played if mixLayers is true.
	mix     float64 // The mix parameter: (1-mix)*sample1 + mix*sample2.
	idx     float32 // Index being played in the given sample.
	amp1    float64 // Current amplification for sample 1. 
	amp2    float64 // Current amplification for sample 2. 
	pan     float64 // Pan of playing sample: -1 is hard left, 1 is hard right.
	tau     float64 // Decay constant (0 is disabled).

	fadeAmp float64 // Fade in amplification if controls.TauFadeIn > 0.
}

// NewPlayingSample:
// If not mixing layers, then mix should be 0, and amp2 is ignored. 
func NewPlayingSample(
	sample1, sample2 *Sample, amp1, amp2, pan, mix float64) *PlayingSample {

	ps := new(PlayingSample)
	ps.sample1 = sample1
	ps.sample2 = sample2
	ps.mix = mix
	ps.amp1 = amp1
	ps.amp2 = amp2
	ps.pan = pan
	ps.tau = 0
	
	if controls.TauFadeIn != 0 {
		ps.fadeAmp = 1
	} else {
		ps.fadeAmp = 0
	}
	
	N := float32(0)
	if controls.TauFadeIn != 0 {
		N = float32(math.Log(ampCutoff) / math.Log(controls.TauFadeIn))
	}

	if sample2 != nil {
		ps.idx = float32(sample2.Idx0) - N
	} else {
		ps.idx = float32(sample1.Idx0) - N
	}
	
	if ps.idx < 0 {
		ps.idx = 0 
	}

	return ps
}

// Get the current sample value, applying amp and pan, converting to 
// a float64 value.
func (ps *PlayingSample) getCurrentSample(idx int) (float32, float32) {
	var s *Sample
	var amp float32

	if idx == 0 {
		s = ps.sample1
		amp = float32(ps.amp1)
	} else {
		s = ps.sample2
		amp = float32(ps.amp2)
	}
	
	L, R := s.Interp(ps.idx)

	L *= amp
	R *= amp

	if ps.pan == 0 {
		return L, R
	} else if ps.pan < 0 {
		return L - float32(ps.pan) * R, float32(1+ps.pan) * R
	} else {
		return float32(1-ps.pan) * L, R + float32(ps.pan) * L
	}
}

func (ps *PlayingSample) WriteOutput(buf *Sound) bool {
	var L, R float32

	for i, _ := range buf.L {
		// Decay amplitude.
		if ps.tau != 0 {
			ps.amp1 *= ps.tau
			ps.amp2 *= ps.tau
			// Amplitude too low. Done playing this.
			if ps.amp1 < ampCutoff && ps.amp2 < ampCutoff {
				return false
			}
		} 

		// Fade in.
		if ps.fadeAmp > 0 {
			ps.fadeAmp *= controls.TauFadeIn
			if ps.fadeAmp < ampCutoff {
				ps.fadeAmp = 0
			}
		}
		
		// Add layer1. 
		L, R = ps.getCurrentSample(0)
		buf.L[i] += L * float32((1 - ps.mix) * (1 - ps.fadeAmp))
		buf.R[i] += R * float32((1 - ps.mix) * (1 - ps.fadeAmp))

		// If mixing layers, add second sample.
		if ps.mix != 0 {
			L, R = ps.getCurrentSample(1)
			buf.L[i] += L * float32((ps.mix) * (1 - ps.fadeAmp))
			buf.R[i] += R * float32((ps.mix) * (1 - ps.fadeAmp))
		}

		// Update index.
		ps.idx += controls.di

		// Done playing.
		if ps.idx >= float32(ps.sample1.Len - 1) ||
			(ps.sample2 != nil && ps.idx >= float32(ps.sample2.Len - 1)) {
			return false
		}
	}

	return true
}
