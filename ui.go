package jlsampler

import (
	"bufio"
	"os"
)

func UI(junk interface{}, QuitOut chan bool) {
	reader := bufio.NewReader(os.Stdin)
	
	var err error
	var line string
	
	for {
		if line, err = reader.ReadString('\n'); err != nil {
			Println("Error reading input:", err)
			return
		}
		line = line[:len(line) - 1] // Strip \n. 
		if len(line) > 0 {
			if line == "help" {
				controls.Print()
			} else if line == "quit" {
				break
			} else {
				controls.ProcessCommand(line)
			}
		}
	}
	QuitOut <- true
}