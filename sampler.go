package jlsampler

import (
	"github.com/johnnylee/jackclient"
	"os"
	"sync"
)

type Sampler struct {
	mutex        *sync.Mutex
	controls     *Controls
	midiListener *MidiListener
	jackClient   *jackclient.JackClient
	keySamplers  []*KeySampler // Per key (128).

	buf *Sound
	diBase float32
	di  []float32
}

func NewSampler(name, path string) (*Sampler, error) {
	var err error

	// Try to load config.
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	// Change to sampler directory.
	originalDir, _ := os.Getwd()
	if err = os.Chdir(path); err != nil {
		return nil, err
	}
	defer os.Chdir(originalDir)

	// New sampler object.
	s := new(Sampler)
	s.mutex = new(sync.Mutex)

	// Create controls.
	s.controls = NewControls(s)

	if err = s.controls.LoadMidiConfig(); err != nil {
		return nil, err
	}

	if err = s.controls.LoadFrom("defaults.js"); err != nil {
		return nil, err
	}

	// Create midiListener.
	s.midiListener, err = NewMidiListener(s, name, config.MidiIn)
	if err != nil {
		return nil, err
	}

	// The buffer doesn't need any size because it will use the
	// slices passed in by the jack callback.
	s.buf = NewSound(0)

	// Make key samplers.
	s.keySamplers = make([]*KeySampler, 128)

	// Load samples.
	if err = s.loadSamples(); err != nil {
		return nil, err
	}

	// Update crop threshold. This will also update the RMS value.
	s.UpdateCropThresh()

	// Create jackClient.
	s.jackClient, err = jackclient.New(name, 0, 2)
	if err != nil {
		return nil, err
	}

	// Get output sample rate. 
	s.diBase = sampleRate / float32(s.jackClient.GetSampleRate())

	return s, nil
}

func (s *Sampler) Run() {
	go s.midiListener.Run()
	s.jackClient.RegisterCallback(s.JackProcess)
	s.controls.Run()
}

// ----------------------------------------------------------------------------
// These functions run slowly and could cause skipping.
func (s *Sampler) UpdateCropThresh() {
	for _, ks := range s.keySamplers {
		if ks != nil {
			ks.UpdateCropThresh(s.controls.CropThresh)
			ks.UpdateRms(s.controls.RmsTime)
		}
	}
}

func (s *Sampler) UpdateRms() {
	for _, ks := range s.keySamplers {
		if ks != nil {
			ks.UpdateRms(s.controls.RmsTime)
		}
	}
}

// ----------------------------------------------------------------------------
// Functions below protected by mutex.
func (s *Sampler) NoteOnEvent(note int8, value float64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	note += s.controls.Transpose
	if note > 0 && note < 127 {
		if s.keySamplers[note] != nil {
			s.keySamplers[note].NoteOn(value)
		}
	}
}

func (s *Sampler) NoteOffEvent(note int8, value float64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	note += s.controls.Transpose
	if note > 0 && note < 127 {
		if s.keySamplers[note] != nil {
			s.keySamplers[note].NoteOff()
		}
	}
}

func (s *Sampler) ControllerEvent(control int32, value float64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.controls.ProcessMidi(control, value)
}

func (s *Sampler) PitchBendEvent(value float64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	// TODO.
}

// Jack processing callback.
func (s *Sampler) JackProcess(bufIn, bufOut [][]float32) error {
	// Can we just remove this lock? We'll see. 
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.buf.L = bufOut[0]
	s.buf.R = bufOut[1]
	s.buf.Len = len(s.buf.L)

	if len(s.di) != s.buf.Len {
		s.di = make([]float32, s.buf.Len)
	}

	for i := 0; i < s.buf.Len; i++ {
		s.buf.L[i] = 0
		s.buf.R[i] = 0
	}

	for i, _ := range s.di {
		s.di[i] = s.diBase
	}

	for _, ks := range s.keySamplers {
		if ks != nil && ks.HasData() {
			ks.WriteOutput(s.buf, s.di)
		}
	}

	return nil
}
