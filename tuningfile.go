package jlsampler

import (
	"encoding/json"
	"os"
)

type TuningFile struct {
	vals map[string]interface{}
}

func LoadTuningFile() *TuningFile {
	tf := new(TuningFile)

	f, err := os.Open("tuning.js")
	if err != nil {
		Println("Failed to open tuning file:", err)
		return tf
	}

	// read file into interface,
	decoder := json.NewDecoder(f)
	var x interface{}
	if err = decoder.Decode(&x); err != nil {
		Println("Error decoding tuning file:", err)
		return tf
	}

	tf.vals = x.(map[string]interface{})
	return tf
}

func (tf *TuningFile) GetTuning(filename string) float64 {
	value, ok := tf.vals[filename]
	if !ok {
		return 0.0
	}
	return value.(float64)
}
