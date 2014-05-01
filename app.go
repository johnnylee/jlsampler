package jlsampler

import (
	"github.com/johnnylee/glow"
	"os"
	"runtime"
)

func RunApp() {
	var err error

	// Load global config.
	if config, err = LoadConfig(); err != nil {
		Println("Failed to load config file:", err)
		return
	}

	// Load the appropriate player backend. 
	player := NewJackPlayer("jslampler")
	
	// Load global midi controls.
	if err = LoadMidiControls(); err != nil {
		Println("Failed to load midi controls:", err)
		return
	}

	runtime.GOMAXPROCS(config.Procs)

	// Change to sampler directory.
	samplerDir := os.Args[1] 
	originalDir, _ := os.Getwd()
	if err = os.Chdir(samplerDir); err != nil {
		Println("Failed to change to sampler directory:", err)
		return
	}

	// Load defaults.
	if err = controls.LoadFrom("defaults.js"); err != nil {
		Println("Failed to load defaults:", err)
		return
	}

	// Read additional command line arguments. 
	for i := 2; i < len(os.Args); i++ {
		controls.ProcessCommand(os.Args[i])
	}

	// Load the sampler.
	sampler.Load()

	// Change back to current directory. 
	os.Chdir(originalDir)

	// Print control states.
	controls.Print()
	
	// Run the program.
	junk := 0
	g := glow.NewGraph(junk)

	g.AddNode(MidiClient, "MidiClient", "MsgIn", "MsgOut")
	g.AddNode(sampler.Run,
		"Sampler", "MsgIn", "MsgOut", "BufIn", "BufOut", "QuitIn")
	g.AddNode(UI, "UI", "QuitOut")
	g.AddNode(player.Run, "Player", "BufIn", "BufOut")

	g.Connect(config.MidiBufSize, "Sampler:MsgIn", "MidiClient:MsgOut")
	g.Connect(config.MidiBufSize, "MidiClient:MsgIn", "Sampler:MsgOut")
	g.Connect(0, "Player:BufOut", "Sampler:BufIn")
	g.Connect(0, "Sampler:BufOut", "Player:BufIn")
	g.Connect(0, "UI:QuitOut", "Sampler:QuitIn")

	g.SetForeground("UI")

	//Println(g.DotString())
	g.Run()
}
