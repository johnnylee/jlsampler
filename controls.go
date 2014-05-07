package jlsampler

import (
	"bufio"
	"encoding/json"
	"math"
	"os"
	"strconv"
	"strings"
)

// ----------------------------------------------------------------------------
// Input is 1/e decay time in seconds. Output is the per-sample amplitude
// decay factor.
func computeTau(tau float64) float64 {
	if tau == 0 {
		return 0
	}
	return math.Exp(-1.0 / (float64(sampleRate) * tau))
}

// ----------------------------------------------------------------------------
type Controls struct {
	Transpose    int8    // Added to midi note on input.
	Tau          float64 // Key-up decay time constant.
	TauCut       float64 // Key-repeat or cut decay time constant.
	TauFadeIn    float64 // Sample fade in time. 
	CropThresh   float64 // Cut beginning of samples below this threshold.
	RmsTime      float64 // Time period to use to compute sample RMS.
	RmsLow       float64 // RMS for key 21 (Low A).
	RmsHigh      float64 // RMS for key 108 (High C).
	PanLow       float64 // Panning for key 21. -1 is left, 1 is right.
	PanHigh      float64 // Panning for key 108.
	GammaAmp     float64 // Amplitude scaling x^gamma.
	GammaLayer   float64 // Layer scaling.
	VelMult      float64 // Velocity multiplier.
	PitchBendMax int8    // Maximum pitch bend in semitones.
	RRBorrow     int     // Distance to borrow round-robbin samples.
	MixLayers    bool    // It True, mix layers together.
	Sustain      bool    // Sustain pedal value (0-1).
	di           float32 // Step size (due to pitch shift).
}

func NewControls() *Controls {
	c := new(Controls)
	c.LoadDefaults()
	return c
}

func (c *Controls) LoadDefaults() {
	c.Transpose = 0
	c.Tau = 0
	c.TauCut = 0
	c.TauFadeIn = 0
	c.CropThresh = 0
	c.RmsTime = 0.25
	c.RmsLow = 0.20
	c.RmsHigh = 0.04
	c.PanLow = 0.0
	c.PanHigh = 0.0
	c.GammaAmp = 2.2
	c.GammaLayer = 1.0
	c.VelMult = 1.0
	c.PitchBendMax = 1
	c.RRBorrow = 0
	c.MixLayers = false
	c.di = 1.0
	c.Sustain = false
}

func (c *Controls) LoadFrom(path string) error {
	c.LoadDefaults()

	// Open the file.
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	// Decode the json file.
	decoder := json.NewDecoder(f)
	if err = decoder.Decode(c); err != nil {
		return err
	}

	c.UpdateTau(c.Tau)
	c.UpdateTauCut(c.TauCut)
	c.UpdateTauFadeIn(c.TauFadeIn)
	
	return nil
}

func (c *Controls) Print() {
	Println("--------------------------------------------------")
	Println("Transpose:   ", c.Transpose)
	Println("Tau:         ", -1/(math.Log(c.Tau)*sampleRate))
	Println("TauCut:      ", -1/(math.Log(c.TauCut)*sampleRate))
	Println("TauFadeIn:   ", -1/(math.Log(c.TauFadeIn)*sampleRate))
	Println("CropThresh:  ", c.CropThresh)
	Println("RmsTime:     ", c.RmsTime)
	Println("RmsLow:      ", c.RmsLow)
	Println("RmsHigh:     ", c.RmsHigh)
	Println("PanLow:      ", c.PanLow)
	Println("PanHigh:     ", c.PanHigh)
	Println("GammaAmp:    ", c.GammaAmp)
	Println("GammaLayer:  ", c.GammaLayer)
	Println("VelMult:     ", c.VelMult)
	Println("PitchBendMax:", c.PitchBendMax)
	Println("RRBorrow:    ", c.RRBorrow)
	Println("MixLayers:   ", c.MixLayers)
}

func (c *Controls) CalcAmp(key int, velocity, rms float64) float64 {
	m := (c.RmsHigh - c.RmsLow) / 87.0
	amp := (c.RmsLow + m*(float64(key)-21)) / rms
	return amp * math.Pow(velocity, c.GammaAmp)
}

func (c *Controls) CalcPan(key int) float64 {
	m := (c.PanHigh - c.PanLow) / 87.0
	return c.PanLow + m*(float64(key)-21)
}

func (c *Controls) Run() {
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
			if line == "print" {
				c.Print()
			} else if line == "quit" {
				break
			} else {
				c.ProcessCommand(line)
			}
		}
	}
}

func (c *Controls) ProcessCommand(cmd string) {
	sp := strings.Split(cmd, "=")
	
	cmd = sp[0]
	
	// Load and Unload commands are strings. 
	if cmd == "Load" {
		if err := sampler.Load(sp[1]); err != nil {
			Println("Error loading samples:", err)
		}
		return
	} else if cmd == "Unload" {
		sampler.Unload()
		return
	}
	
	// Convert bools to floats. 
	sp[1] = strings.ToLower(sp[1])
	if sp[1] == "true" {
		sp[1] = "1"
	} else if sp[1] == "false" {
		sp[1] = "0"
	}
	
	// All other commands take floats. 
	val, err := strconv.ParseFloat(sp[1], 64)
	if err != nil {
		Println("Couldn't parse numerical value:", cmd, val, err)
		return
	}

	switch cmd {
	case "Transpose":
		c.UpdateTranspose(val)
	case "Tau":
		c.UpdateTau(val)
	case "TauCut":
		c.UpdateTauCut(val)
	case "TauFadeIn":
		c.UpdateTauFadeIn(val)
	case "CropThresh":
		c.UpdateCropThresh(val)
	case "RmsTime":
		c.UpdateRmsTime(val)
	case "RmsLow":
		c.UpdateRmsLow(val)
	case "RmsHigh":
		c.UpdateRmsHigh(val)
	case "PanLow":
		c.UpdatePanLow(val)
	case "PanHigh":
		c.UpdatePanHigh(val)
	case "GammaAmp":
		c.UpdateGammaAmp(val)
	case "GammaLayer":
		c.UpdateGammaLayer(val)
	case "VelMult":
		c.UpdateVelMult(val)
	case "PitchBendMax":
		c.UpdatePitchBendMax(val)
	case "MixLayers":
		c.UpdateMixLayers(val)
	default:
		Println("Unknown command:", cmd)
		return
	}
}

// ----------------------------------------------------------------------------
// Update functions, one for each control.
func (c *Controls) UpdateTranspose(x float64) {
	c.Transpose = int8(x)
	Println("Transpose:", c.Transpose)
}

func (c *Controls) UpdateTau(x float64) {
	c.Tau = computeTau(x)
	Println("Tau:", x)
}

func (c *Controls) UpdateTauCut(x float64) {
	c.TauCut = computeTau(x)
	Println("TauCut:", x)
}

func (c *Controls) UpdateTauFadeIn(x float64) {
	c.TauFadeIn = computeTau(x)
	Println("TauFadeIn:", x)
}

func (c *Controls) UpdateCropThresh(x float64) {
	c.CropThresh = x
	Println("CropThresh:", x)
	sampler.UpdateCropThresh()
}

func (c *Controls) UpdateRmsTime(x float64) {
	c.RmsTime = x
	Println("RmsTime:", x)
	sampler.UpdateRms()
}

func (c *Controls) UpdateRmsLow(x float64) {
	c.RmsLow = x
	Println("RmsLow:", x)
}

func (c *Controls) UpdateRmsHigh(x float64) {
	c.RmsHigh = x
	Println("RmsHigh:", x)
}

func (c *Controls) UpdatePanLow(x float64) {
	c.PanLow = x
	Println("PanLow:", x)
}

func (c *Controls) UpdatePanHigh(x float64) {
	c.PanHigh = x
	Println("PanHigh:", x)
}

func (c *Controls) UpdateGammaAmp(x float64) {
	c.GammaAmp = x
	Println("GammaAmp:", x)
}

func (c *Controls) UpdateGammaLayer(x float64) {
	c.GammaLayer = x
	Println("GammaLayer:", x)
}

func (c *Controls) UpdateVelMult(x float64) {
	c.VelMult = x
	Println("VelMult:", x)
}

func (c *Controls) UpdatePitchBendMax(x float64) {
	c.PitchBendMax = int8(x)
	Println("PitchBendMax:", c.PitchBendMax)
}

func (c *Controls) UpdateMixLayers(x float64) {
	c.MixLayers = x > 0.5
	Println("MixLayers:", c.MixLayers)
}

func (c *Controls) UpdateSustain(x float64) {
	c.Sustain = x > 0.5
}

func (c *Controls) UpdatePitchBend(x float64) {
	c.di = float32(math.Pow(2.0, x*float64(c.PitchBendMax)/12.0))
	Println("di:", x, c.di)
}

