package jlsampler

// ----------------------------------------------------------------------------
// Global constants.
const (
	sampleRate = 48000   // Fixed sample rate.
	maxVal16   = 32767   // 16-bit maximum sample value.
	maxVal24   = 8388607 // 24-bit maximum sample value.
	ampCutoff  = 1e-5    // Cut-off amplitude of decaying sample.
)

// ----------------------------------------------------------------------------
// Global objects.
var config *Config 
var controls *Controls = NewControls()
var midiControls []func(float64) = make([]func(float64), 128)
var sampler *Sampler 
var midiListener *MidiListener

