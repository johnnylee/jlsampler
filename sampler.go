package jlsampler

import (
	"sync"
	"time"
)

type Sampler struct {
	keySamplers []*KeySampler // Per key (128).
}

func NewSampler() *Sampler {
	s := new(Sampler)
	s.keySamplers = make([]*KeySampler, 128)
	return s
}

func (s *Sampler) Load() {
	tuningFile := LoadTuningFile()
	var wg sync.WaitGroup
	
	for key := 0; key < 128; key++ {
		wg.Add(1)
		go func(key int) {
			defer wg.Done()

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
					continue
				}
				
				semitones := tuningFile.GetTuning(path)
				if semitones != 0 {
					sample = sample.Stretched(semitones)
				}
				
				ks.AddSample(sample, layer)
			}
			
			s.keySamplers[key] = ks
		}(key)
	}
	wg.Wait()
	
	// Nieghbor borrowing. 
	if controls.RRBorrow > 0 {
		s.BorrowSamples()
	}

	// Transpose to fill missing notes. 
	s.FillTransposeSamples()
}

func (s *Sampler) BorrowSamples() {
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
			for j := 1; j < controls.RRBorrow + 1; j++ {
				// Borrow from below. 
				if ks2 = originalKeySamplers[i - j]; ks2 != nil {
					ks.BorrowFrom(ks2)
				}
				
				// Borrow from above. 
				if ks2 = originalKeySamplers[i + j]; ks2 != nil {
					ks.BorrowFrom(ks2)
				}
			}
		}(i)
	}
	wg.Wait()
}

func (s *Sampler) FillTransposeSamples() {
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

func (s *Sampler) Run(junk interface{}, MsgIn, MsgOut chan *MidiMsg,
	BufIn, BufOut chan *Sound, QuitIn chan bool) {

	var msg *MidiMsg
	var buf *Sound

	var ks *KeySampler

	for {
		select {
		case <-QuitIn:
			close(MsgOut)
			close(BufOut)
			return
		case msg = <-MsgIn:
			switch msg.Type {
			case MsgNoteOn:
				if ks = s.keySamplers[msg.Note]; ks != nil {
					ks.NoteOn(msg.Value)
				}
			case MsgNoteOff:
				if ks = s.keySamplers[msg.Note]; ks != nil {
					ks.NoteOff()
				}
			case MsgPitchBend:
				controls.UpdatePitchBend(msg.Value)
			case MsgControl:
				if midiControls[msg.Control] != nil {
					midiControls[msg.Control](msg.Value)
				}
			}

			MsgOut <- msg

		case buf = <-BufIn:
			buf.T0 = time.Now().UnixNano()
			buf.Zero()

			// Write output.
			for _, ks = range s.keySamplers {
				if ks != nil && ks.HasData() {
					ks.WriteOutput(buf)
				}
			}
			BufOut <- buf
		}
	}
}

func (s *Sampler) UpdateCropThresh() {
	for _, ks := range s.keySamplers {
		if ks != nil {
			ks.UpdateCropThresh()
		}
	}
}

func (s *Sampler) UpdateRms() {
	for _, ks := range s.keySamplers {
		if ks != nil {
			ks.UpdateRms()
		}
	}
}

