package jlsampler

import (
	"runtime"
)

func RunApp() {
	var err error

	// Load global config.
	if config, err = LoadConfig(); err != nil {
		Println("Failed to load config file:", err)
		return
	}

	runtime.GOMAXPROCS(config.Procs)

	// Load global midi controls.
	if err = LoadMidiControls(); err != nil {
		Println("Failed to load midi controls:", err)
		return
	}
	
	// Create new sampler. 
	sampler, err = NewSampler("JLSampler")
	if err != nil {
		Println("Failed to create sampler:", err)
		return
	}
	
	// Create midi listener. 
	midiListener, err = NewMidiListener("JLSampler", config.MidiIn)
	if err != nil {
		Println("Failed to create midi listener:", err)
		return
	}
	
	// Run. 
	go midiListener.Run()
	controls.Run()
}
