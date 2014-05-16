package jlsampler

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// ----------------------------------------------------------------------------
func samplePaths(key int) []string {
	glob := fmt.Sprintf("samples/on-%03d-*.flac", key)
	paths, err := filepath.Glob(glob)
	if err != nil {
		return []string{}
	}

	sort.Strings(paths)

	return paths
}

func samplePathInfo(path string) (int, int, int, error) {
	ext := filepath.Ext(path)
	path = path[0 : len(path)-len(ext)]

	items := strings.Split(path, "-")
	if len(items) < 4 {
		return 0, 0, 0, errors.New("Incorrect filename format: " + path)
	}

	key, err := strconv.Atoi(items[1])
	if err != nil {
		return 0, 0, 0, err
	}
	layer, err := strconv.Atoi(items[2])
	if err != nil {
		return 0, 0, 0, err
	}
	variation, err := strconv.Atoi(items[3])
	if err != nil {
		return 0, 0, 0, err
	}

	return key, layer - 1, variation, nil
}

// ----------------------------------------------------------------------------
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

	s.borrowSamples()
	s.fillTransposeSamples()

	return nil
}

func (s *Sampler) loadKey(
	key int, tuningFile *TuningFile, ok *bool, wg *sync.WaitGroup) {

	defer wg.Done()

	// Get paths for the files in each sample layer.
	paths := samplePaths(key)
	if len(paths) == 0 {
		return
	}

	// We have at least one file.
	ks := NewKeySampler(s.controls, key)

	// Loop through paths, loading samples.
	for _, path := range paths {
		_, layer, _, err := samplePathInfo(path)
		if err != nil {
			Println("Failed to get sample info:", path)
			*ok = false
			return
		}
		
		sample, err := LoadFlac(path)
		if err != nil {
			Println("Failed to load sample:", path, "\nError:", err)
			*ok = false
			return
		}

		semitones := tuningFile.GetTuning(path)
		if semitones != 0 {
			sample = sample.Stretched(semitones)
		}

		s.loadKeySample(sample, layer, ks)
	}

	Println("Loaded key:", key)
	s.keySamplers[key] = ks
	runtime.GC() // Force garbage collection here? 
}

func (s *Sampler) loadKeySample(sample *Sample, layer int, ks *KeySampler) {
	if s.controls.FakeLayerRC { 
		// Generate fake layer. 
		// Must have two layers. 
		for ks.NumLayers() < 2 {
			ks.AddLayer()
		}
		
		fakeSample := sample.FakeLayerRC()
		ks.AddSample(fakeSample, 0)
		ks.AddSample(sample, 1)
		return
	} else {
		for ks.NumLayers() < layer + 1 {
			ks.AddLayer()
		}
		ks.AddSample(sample, layer)
	}
}

func (s *Sampler) borrowSamples() {
	rrBorrow := int(s.controls.RRBorrow)

	if rrBorrow <= 0 {
		return
	}

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
			for j := 1; j < rrBorrow+1; j++ {
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

			if originalKeySamplers[i] != nil {
				return
			}

			for j := 1; j < 87; j++ {
				// Try lower note.
				if i-j > 20 && i-j < 109 {
					if ks2 := originalKeySamplers[i-j]; ks2 != nil {
						Println("Transposing:", i-j, "->", i)
						s.keySamplers[i] = ks2.Transpose(j)
						return
					}
				}
				// Try higher note.
				if i+j > 20 && i+j < 109 {
					if ks2 := originalKeySamplers[i+j]; ks2 != nil {
						Println("Transposing:", i+j, "->", i)
						s.keySamplers[i] = ks2.Transpose(-j)
						return
					}
				}
			}
		}(i)
	}
	wg.Wait()
}
