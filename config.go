package jlsampler

import (
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"
)

type Config struct {
	Procs int // Number of processors to use.
	Poly  int // Single key polyphony.
	MidiIn      string // Controller midi port (keyboard).
	MidiBufSize int    // The midi buffer size.
}

func LoadConfig() (*Config, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(usr.HomeDir, ".jlsampler/config.js")

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(f)
	config := new(Config)
	if err = decoder.Decode(config); err != nil {
		return nil, err
	}
	
	return config, nil
}
