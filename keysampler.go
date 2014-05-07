package jlsampler

import (
	"errors"
	"math"
)

// ----------------------------------------------------------------------------
type KeySampler struct {
	Key     int            // The midi key number.
	on      bool           // True if key is on (down).
	layers  []*SampleLayer // The sample layers.

	// A slice of playing samples. The length is the number of playing samples,
	// and the capacity is config.Poly. 
	playing []*PlayingSample
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

	ks.playing = make([]*PlayingSample, 0, config.Poly)

	return ks
}

func (ks *KeySampler) HasMissingLayers() bool {
	for _, sl := range ks.layers {
		if sl.NumSamples() == 0 {
			return true
		}
	}
	return false
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

	ks2.playing = make([]*PlayingSample, 0, cap(ks.playing))
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
	
	mix := layerVal - float64(layer1)
	
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
	ks.on = true

	// Kick a sound out of the queue if we have to.
	if len(ks.playing) == cap(ks.playing) {
		for i := 1; i < len(ks.playing); i++ {
			ks.playing[i-1] = ks.playing[i]
		}
		ks.playing = ks.playing[:len(ks.playing) - 1]
		Println("Sound stopped. Not enough polyphony.")
	}

	// Loop through playing samples. All currently playing samples should
	// decay with constant tauCut.
	if controls.TauCut != 0 {
		for _, ps := range(ks.playing) {
			ps.tau = controls.TauCut
		}
	}

	// Add a new playing sample.
	ks.playing = append(ks.playing, ks.getPlayingSample(velocity))
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
		for _, ps := range ks.playing {
			if ps.tau == 0 {
				ps.tau = controls.Tau
			}
		}
	}
}

func (ks *KeySampler) HasData() bool {
	return len(ks.playing) != 0
}

func (ks *KeySampler) WriteOutput(buf *Sound) {
	var ps *PlayingSample
		
	// Check for sustain pedal depressed. 
	ps = ks.playing[len(ks.playing) - 1]
	if controls.Sustain && ps.tau != 0 {
		ps.tau = 0
	}
		
	iIn := 0
	for _, ps = range ks.playing {
		// Check for sustain pedal lift.
		if !ks.on && !controls.Sustain && ps.tau == 0 {
			ps.tau = controls.Tau
		}

		if ps.WriteOutput(buf) {
			ks.playing[iIn] = ps
			iIn++
		}
	}

	ks.playing = ks.playing[:iIn]
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

