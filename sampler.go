package jlsampler

import (
	"errors"
	"github.com/johnnylee/jackclient"
	"os"
	"sync"
)

type Sampler struct {
	// It looks like we need two mutexes: one to protect the jack callback
	// and midi processing calls, and one to deal with loading / UpdateRms, 
	// etc.
	procMutex   *sync.Mutex 
	mute        bool // Protected by procMutex. 
	loadMutex   *sync.Mutex
	client      *jackclient.JackClient
	buf         *Sound
	keySamplers []*KeySampler // Per key (128).
}

func NewSampler(name string) (*Sampler, error) {
	var err error 

	s := new(Sampler)
	s.procMutex = new(sync.Mutex)
	s.loadMutex = new(sync.Mutex)
	s.mute = true

	s.client, err = jackclient.New(name, 0, 2)
	if err != nil {
		return nil, err
	}
	s.buf = NewSound(0)
	s.keySamplers = make([]*KeySampler, 128)
	s.client.RegisterCallback(s.JackProcess)
	return s, nil
}

func (s *Sampler) setMute(value bool) {
	s.procMutex.Lock()
	s.mute = value
	s.procMutex.Unlock()
}

// ----------------------------------------------------------------------------
// Functions below protected by loadMutex. 

func (s *Sampler) Load(path string) error {
	var err error 

	// Lock and mute output. 
	s.loadMutex.Lock()
	defer s.loadMutex.Unlock()
	s.setMute(true)
	
	// Unload. 
	s.keySamplers = make([]*KeySampler, 128)

	// Change to sample directory. 
	originalDir, _ := os.Getwd()
	if err = os.Chdir(path); err != nil {
		return err
	}
	
	// Load controls. 
	if err = controls.LoadFrom("defaults.js"); err != nil {
		return err
	}
	
	// Load samples. 
	if err = s.loadSamples(); err != nil {
		return err
	}
	
	// Change back to original directory. 
	os.Chdir(originalDir)
	
	s.setMute(false)
	return nil 
}

func (s *Sampler) Unload() {
	s.loadMutex.Lock()
	defer s.loadMutex.Unlock()
	s.setMute(true)
	s.keySamplers = make([]*KeySampler, 128)
}

func (s *Sampler) UpdateCropThresh() {
	s.setMute(true)
	defer s.setMute(false)
	s.loadMutex.Lock()
	defer s.loadMutex.Unlock()

	for _, ks := range s.keySamplers {
		if ks != nil {
			ks.UpdateCropThresh()
		}
	}
}

func (s *Sampler) UpdateRms() {
	s.setMute(true)
	defer s.setMute(false)
	s.loadMutex.Lock()
	defer s.loadMutex.Unlock()

	for _, ks := range s.keySamplers {
		if ks != nil {
			ks.UpdateRms()
		}
	}
}

// ----------------------------------------------------------------------------
// Functions below protected by procMutex. 

func (s *Sampler) NoteOnEvent(note int8, value float64) {
	s.procMutex.Lock()
	defer s.procMutex.Unlock()
	if s.mute {
		return
	}
	if s.keySamplers[note] != nil {
		s.keySamplers[note].NoteOn(value)
	}
}

func (s *Sampler) NoteOffEvent(note int8, value float64) {
	s.procMutex.Lock()
	defer s.procMutex.Unlock()
	if s.mute {
		return
	}
	if s.keySamplers[note] != nil {
		s.keySamplers[note].NoteOff()
	}
}

func (s *Sampler) ControllerEvent(control int32, value float64) {
	s.procMutex.Lock()
	defer s.procMutex.Unlock()
	if s.mute {
		return
	}
	if midiControls[control] != nil {
		midiControls[control](value)
	}
}

func (s *Sampler) PitchBendEvent(value float64) {
	s.procMutex.Lock()
	defer s.procMutex.Unlock()
	if s.mute {
		return
	}
	controls.UpdatePitchBend(value)
}

// Jack processing callback. 
func (s *Sampler) JackProcess(bufIn, bufOut [][]float32) error {
	s.procMutex.Lock()
	defer s.procMutex.Unlock()
	if s.mute {
		return nil 
	}
	
	s.buf.L = bufOut[0]
	s.buf.R = bufOut[1]
	s.buf.Len = len(s.buf.L)

	for i := 0; i < s.buf.Len; i++ {
		s.buf.L[i] = 0
		s.buf.R[i] = 0
	}
	
	// TODO: ideally we'd get smoothed pitchbend information here and share
	// it with all our playing samples. 
	// It would also be nice to have amplitude modulation here. 
	for _, ks := range s.keySamplers {
		if ks != nil && ks.HasData() {
			ks.WriteOutput(s.buf)
		}
	}
	
	return nil
}

// ----------------------------------------------------------------------------
// Loading. 

// loadKey: Create a KeySampler in keySamplers[key]. If an error occurs, 
// set ok to false. 
func (s *Sampler) loadKey(
	key int, tuningFile *TuningFile, ok *bool, wg *sync.WaitGroup) {

	defer wg.Done()
	
	// Get paths for the files in each sample layer. 
	layers, paths := samplePaths(key)
	if layers == 0 {
		return
	}
	
	// We have at least one layer. 
	ks := NewKeySampler(key, layers)
	
	// Loop through paths, loading samples. 
	for _, path := range paths {
		_, layer, _ := samplePathInfo(path)
		Println("Loading:", path)
		
		sample, err := LoadFlac(path)
		if err != nil {
			Println(
				"Failed to load sample:", path, "\nError:", err)
			*ok = false
			return
		}
		
		semitones := tuningFile.GetTuning(path)
		if semitones != 0 {
			sample = sample.Stretched(semitones)
		}
		
		ks.AddSample(sample, layer)
	}
	
	// If we ended up with an empty SampleLayer there's a problem. 
	if ks.HasMissingLayers() {
		*ok = false
	}
	
	s.keySamplers[key] = ks
}

// loadSamples: Load all samples for the sampler. 
func (s *Sampler) loadSamples() error {
	tuningFile := LoadTuningFile()
	wg := new(sync.WaitGroup)

	ok := true
	for key := 0; key < 128; key++ {
		wg.Add(1)
		go s.loadKey(key, tuningFile, &ok, wg)
	}
	wg.Wait()
	
	if !ok {
		return errors.New("Error loading samples.")
	}

	// Nieghbor borrowing.
	if controls.RRBorrow > 0 {
		s.borrowSamples()
	}

	// Transpose to fill missing notes.
	s.fillTransposeSamples()
	
	return nil 
}

func (s *Sampler) borrowSamples() {
	// We need to copy the original KeySamplers so we don't
	// propagate borrows as we go.
	originalKeySamplers := make([]*KeySampler, 128)

	for i, ks := range s.keySamplers {
		if ks != nil {
			originalKeySamplers[i] = ks.Copy()
		}
	}

	var wg sync.WaitGroup

	for i := 21; i < 109; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			if s.keySamplers[i] == nil {
				return
			}

			var ks2 *KeySampler
			ks := s.keySamplers[i]
			for j := 1; j < controls.RRBorrow+1; j++ {
				// Borrow from below.
				if ks2 = originalKeySamplers[i-j]; ks2 != nil {
					ks.BorrowFrom(ks2)
				}

				// Borrow from above.
				if ks2 = originalKeySamplers[i+j]; ks2 != nil {
					ks.BorrowFrom(ks2)
				}
			}
		}(i)
	}
	wg.Wait()
}

func (s *Sampler) fillTransposeSamples() {
	doTranspose := func(iFrom, iTo int) {
		Println("Transposing:", iFrom, "->", iTo)
		s.keySamplers[iTo] = s.keySamplers[iFrom].Transpose(iTo - iFrom)
	}

	done := false

	var wg sync.WaitGroup
	inds := make([]int, 0)

	// Fill in the high samples first.
	idxHigh := 108
	for s.keySamplers[idxHigh] == nil {
		idxHigh--
	}

	for i := idxHigh + 1; i < 109; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			doTranspose(idxHigh, i)
		}(i)
	}
	wg.Wait()

	// Fill in the low samples.
	idxLow := 21
	for s.keySamplers[idxLow] == nil {
		idxLow++
	}

	for i := 21; i < idxLow; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			doTranspose(idxLow, i)
		}(i)
	}
	wg.Wait()

	for !done {
		done = true
		inds = inds[:0]

		// Look for samples with an adjacent lower note.
		for i := 108; i > 21; i-- {
			if s.keySamplers[i] != nil {
				continue
			} else if s.keySamplers[i-1] != nil {
				done = false
				inds = append(inds, i)
			}
		}

		for _, i := range inds {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				doTranspose(i-1, i)
			}(i)
		}
		wg.Wait()

		inds = inds[:0]
		// Look for samples with an adjacent higher note.
		for i := 21; i < 108; i++ {
			if s.keySamplers[i] != nil {
				continue
			} else if s.keySamplers[i+1] != nil {
				done = false
				inds = append(inds, i)
			}
		}

		for _, i := range inds {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				doTranspose(i+1, i)
			}(i)
		}
		wg.Wait()
	}
}
