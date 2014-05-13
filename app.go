package jlsampler

import (
	"os"
	"runtime"
)

func RunApp() {
	var err error

	if len(os.Args) < 2 {
		Println("Usage:", os.Args[0], "sampler-path", "[name]")
		return
	}

	path := os.Args[1]

	name := "JLSampler"
	if len(os.Args) > 2 {
		name = os.Args[2]
	}

	runtime.GOMAXPROCS(runtime.NumCPU())
	sampler, err := NewSampler(name, path)
	if err != nil {
		Println("Error:", err)
		return
	}

	sampler.Run()
}
