package jlsampler

import (
	"errors"
	"math"
)

// ----------------------------------------------------------------------------
type KeySampler struct {
	Key     int            // The midi key number.
	hasData bool           // True if the sampler has data to write.
	on      bool           // True if key is on (down).
	layers  []*SampleLayer // The sample layers.

	// A queue of samples that are currently being played for this key.
	playing chan *PlayingSample
}

func NewKeySampler(key, numLayers int) *KeySampler {
	ks := new(KeySampler)
	ks.Key = key
	ks.on = false

	// Make sample layers.
	ks.layers = make([]*SampleLayer, numLayers)
	for i := 0; i < numLayers; i++ {
		ks.layers[i] = new(SampleLayer)
	}

	ks.playing = make(chan *PlayingSample, config.Poly)
	return ks
}

func (ks *KeySampler) Copy() *KeySampler {
	ks2 := NewKeySampler(ks.Key, len(ks.layers))
	for i := 0; i < len(ks.layers); i++ {
		ks2.layers[i] = ks.layers[i].Copy()
	}
	return ks2
}

func (ks *KeySampler) AddSample(sample *Sample, layer int) {
	ks.layers[layer].AddSample(sample)
}

func (ks *KeySampler) BorrowFrom(ks2 *KeySampler) error {
	// Make sure both KeySamplers have the same number of layers. 
	if len(ks.layers) != len(ks2.layers) {
		return errors.New("Borrowing requires the same number of layers.")
	}
	
	Println("Borrowing samples:", ks.Key, "<-", ks2.Key)
	
	// Compute the amount of stretching necessary. 
	semitones := ks.Key - ks2.Key
	
	for i := 0; i < len(ks.layers); i++ {
		layer := ks.layers[i]
		layer2 := ks2.layers[i]
		layer.BorrowFrom(layer2, semitones)
	}
	
	return nil
}

func (ks *KeySampler) Transpose(trans int) *KeySampler {
	ks2 := new(KeySampler)
	ks2.Key = ks.Key + trans

	ks2.layers = make([]*SampleLayer, 0)
	for _, layer := range ks.layers {
		ks2.layers = append(ks2.layers, layer.Transpose(trans))
	}

	ks2.playing = make(chan *PlayingSample, cap(ks.playing))

	return ks2
}

func (ks *KeySampler) getPlayingSample(velocity float64) *PlayingSample {
	if controls.MixLayers {
		return ks.getPlayingSampleMix(velocity)
	} else {
		return ks.getPlayingSampleBasic(velocity)
	}
}

func (ks *KeySampler) getPlayingSampleBasic(velocity float64) *PlayingSample {
	numLayers := int64(len(ks.layers))

	// Get the layer. 
	layer := int64(
		float64(numLayers) * math.Pow(velocity, controls.GammaLayer))

	if layer > numLayers - 1 {
		layer = numLayers - 1
	}

	// Get a sample from the first layer.
	_, sample := ks.layers[layer].GetSample(-1)

	// Compute the amplitude of the sample.
	amp := controls.CalcAmp(ks.Key, velocity, sample.Rms)

	// Compute the pan.
	pan := controls.CalcPan(ks.Key)
	
	return NewPlayingSample(sample, nil, amp, 0, pan, 0)
}

func (ks *KeySampler) getPlayingSampleMix(velocity float64) *PlayingSample {
	numLayers := int64(len(ks.layers))

	layerVal := float64(numLayers-1) * math.Pow(velocity, controls.GammaLayer)
	layer1 := int64(layerVal)
	layer2 := layer1 + 1
	
	if layer2 > numLayers - 1 {
		layer2 = numLayers - 1
	}
	
	mix := float32(layerVal) - float32(layer1)
	
	// Samples. 
	sIdx, sample1 := ks.layers[layer1].GetSample(-1)
	_, sample2 := ks.layers[layer2].GetSample(sIdx)
	
	// Amps. 
	amp1 := controls.CalcAmp(ks.Key, velocity, sample1.Rms)
	amp2 := controls.CalcAmp(ks.Key, velocity, sample2.Rms)
	
	// Compute pan. 
	pan := controls.CalcPan(ks.Key)
	
	return NewPlayingSample(sample1, sample2, amp1, amp2, pan, mix)
}

func (ks *KeySampler) NoteOn(velocity float64) {
	ks.hasData = true
	ks.on = true

	// Kick a sound out of the queue if we have to.
	if len(ks.playing) == cap(ks.playing) {
		Println("Sound stopped. Not enough polyphony.")
		<-ks.playing
	}

	// Loop through playing samples. All currently playing samples should
	// decay with constant tauCut.
	if controls.TauCut != 0 {
		N := len(ks.playing)
		for i := 0; i < N; i++ {
			ps := <-ks.playing
			ps.tau = controls.TauCut
			ks.playing <- ps
		}
	}

	// Add a new playing sample.
	ks.playing <- ks.getPlayingSample(velocity)
}

func (ks *KeySampler) NoteOff() {
	ks.on = false

	// If sustaining, there's nothing to do.
	if controls.Sustain {
		return
	}

	// Loop through playing sounds. If any aren't decaying, then
	// they need to have tau set.
	if controls.Tau != 0 {
		N := len(ks.playing)
		for i := 0; i < N; i++ {
			ps := <-ks.playing
			if ps.tau == 0 {
				ps.tau = controls.Tau
			}
			ks.playing <- ps
		}
	}
}

func (ks *KeySampler) HasData() bool {
	return ks.hasData
}

func (ks *KeySampler) WriteOutput(buf *Sound) {
	ks.hasData = false
	var ps *PlayingSample

	N := len(ks.playing)
	for i := 0; i < N; i++ {
		ps = <-ks.playing

		// Check for sustain pedal depressed.
		if i == N-1 && controls.Sustain && ps.tau != 0 {
			ps.tau = 0
		}

		// Check for sustain pedal lift.
		if !ks.on && !controls.Sustain && ps.tau == 0 {
			ps.tau = controls.Tau
		}

		if ps.WriteOutput(buf) {
			ks.hasData = true
			ks.playing <- ps
		}
	}
}

func (ks *KeySampler) UpdateCropThresh() {
	for _, sl := range ks.layers {
		sl.UpdateCropThresh()
	}
}

func (ks *KeySampler) UpdateRms() {
	for _, sl := range ks.layers {
		sl.UpdateRms()
	}
}

