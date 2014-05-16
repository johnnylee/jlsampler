package jlsampler

import (
	"bufio"
	"encoding/json"
	"errors"
	"math"
	"os"
	"os/user"
	"path/filepath"
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
	sampler *Sampler // For callbacks.
	NFadeIn float32  // Fade-in length in samples. Computed from TauFadeIn.

	Transpose     int8 // Added to midi note on input.
	PitchBendMax  int8 // Maximum pitch bend in semitones.
	RRBorrow      int8 // Distance to borrow round-robbin samples.

	Tau       float64 // Key-up decay time constant.
	TauCut    float64 // Key-repeat or cut decay time constant.
	TauFadeIn float64 // Sample fade in time.

	CropThresh float64 // Cut beginning of samples below this threshold.
	RmsTime    float64 // Time period to use to compute sample RMS.
	RmsLow     float64 // RMS for key 21 (Low A).
	RmsHigh    float64 // RMS for key 108 (High C).
	PanLow     float64 // Panning for key 21. -1 is left, 1 is right.
	PanHigh    float64 // Panning for key 108.
	GammaAmp   float64 // Amplitude scaling x^gamma.
	GammaLayer float64 // Layer scaling.
	VelMult    float64 // Velocity multiplier.

	MixLayers   bool // It True, mix layers together.
	FakeLayerRC bool // Use RC filter to construct fake zero-layer. 
	Sustain     bool // Sustain pedal value (0-1).

	// A map from control name to update function.
	updateMap map[string]func(float64)

	// Bindings for midi controls.
	midiControls []func(float64)
}

func NewControls(sampler *Sampler) *Controls {
	c := new(Controls)
	c.sampler = sampler
	c.Transpose = 0
	c.PitchBendMax = 1
	c.RRBorrow = 0
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
	c.MixLayers = false
	c.FakeLayerRC = false
	c.Sustain = false

	c.updateMap = map[string]func(float64){
		"Transpose":    c.UpdateTranspose,
		"PitchBendMax": c.UpdatePitchBendMax,
		"Tau":          c.UpdateTau,
		"TauCut":       c.UpdateTauCut,
		"TauFadeIn":    c.UpdateTauFadeIn,
		"CropThresh":   c.UpdateCropThresh,
		"RmsTime":      c.UpdateRmsTime,
		"RmsLow":       c.UpdateRmsLow,
		"RmsHigh":      c.UpdateRmsHigh,
		"PanLow":       c.UpdatePanLow,
		"PanHigh":      c.UpdatePanHigh,
		"GammaAmp":     c.UpdateGammaAmp,
		"GammaLayer":   c.UpdateGammaLayer,
		"VelMult":      c.UpdateVelMult,
		"MixLayers":    c.UpdateMixLayers,
		"Sustain":      c.UpdateSustain,
	}

	c.midiControls = make([]func(float64), 128)

	return c
}

func (c *Controls) LoadDefaults() {
	println("LoadDefaults: REMOVE FUNCTION.")
}

func (c *Controls) LoadFrom(path string) error {
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

	// Some values have processing applied.
	c.UpdateTau(c.Tau)
	c.UpdateTauCut(c.TauCut)
	c.UpdateTauFadeIn(c.TauFadeIn)

	return nil
}

type ctrlCfg struct {
	Name  string
	Num   int8
	Min   float64
	Max   float64
	Gamma float64
}

func (c *Controls) LoadMidiConfig() error {
	// Get the control config path.
	usr, err := user.Current()
	if err != nil {
		return err
	}
	path := filepath.Join(usr.HomeDir, ".jlsampler/controls.js")

	// Open the file.
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	// Decode the file.
	decoder := json.NewDecoder(f)
	configs := make([]ctrlCfg, 128)
	if err = decoder.Decode(&configs); err != nil {
		return err
	}

	// Load configs.
	for _, cfg := range configs {
		err = c.bind(cfg.Name, cfg.Num, cfg.Min, cfg.Max, cfg.Gamma)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Controls) bind(name string, num int8, min, max, gamma float64) error {
	if num > 119 || num < 0 {
		return errors.New("Midi control number out of range.")
	}

	fn, ok := c.updateMap[name]
	if !ok {
		return errors.New("Unknown control: " + name)
	}

	c.midiControls[num] = func(x float64) {
		fn(min + (max-min)*math.Pow(x, gamma))
	}

	return nil
}

func (c *Controls) Print() {
	Println("--------------------------------------------------")
	Println("Transpose:    ", c.Transpose)
	Println("RRBorrow:     ", c.RRBorrow)
	Println("Tau:          ", -1/(math.Log(c.Tau)*sampleRate))
	Println("TauCut:       ", -1/(math.Log(c.TauCut)*sampleRate))
	Println("TauFadeIn:    ", -1/(math.Log(c.TauFadeIn)*sampleRate))
	Println("CropThresh:   ", c.CropThresh)
	Println("RmsTime:      ", c.RmsTime)
	Println("RmsLow:       ", c.RmsLow)
	Println("RmsHigh:      ", c.RmsHigh)
	Println("PanLow:       ", c.PanLow)
	Println("PanHigh:      ", c.PanHigh)
	Println("GammaAmp:     ", c.GammaAmp)
	Println("GammaLayer:   ", c.GammaLayer)
	Println("VelMult:      ", c.VelMult)
	Println("PitchBendMax: ", c.PitchBendMax)
	Println("MixLayers:    ", c.MixLayers)
	Println("FakeLayerRC:  ", c.FakeLayerRC)
}

func (c *Controls) CalcAmp(key int, velocity, rms float64) float32 {
	if rms <= 0 {
		Println("CalcAmp: RMS value is <= 0.")
		return 0
	}
	m := (c.RmsHigh - c.RmsLow) / 87.0
	amp := (c.RmsLow + m*(float64(key)-21)) / rms
	return float32(amp * math.Pow(velocity, c.GammaAmp))
}

func (c *Controls) CalcPan(key int) float32 {
	m := (c.PanHigh - c.PanLow) / 87.0
	return float32(c.PanLow + m*(float64(key)-21))
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
		line = line[:len(line)-1] // Strip \n.
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
	if len(sp) != 2 {
		Println("No command value given:", cmd)
		return
	}

	// Get command and value.
	cmd = sp[0]
	val, err := strconv.ParseFloat(sp[1], 64)
	if err != nil {
		Println("Couldn't parse numerical value:", cmd, val, err)
		return
	}

	f, ok := c.updateMap[cmd]
	if !ok {
		Println("Unknown command:", cmd)
		return
	}

	f(val)
}

func (c *Controls) ProcessMidi(num int32, value float64) {
	if num < 0 || num > 119 {
		Println("Midi control message out of range:", num)
		return
	}
	if fn := c.midiControls[num]; fn != nil {
		fn(value)
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
	c.NFadeIn = float32(math.Log(ampCutoff) / math.Log(c.TauFadeIn))
	Println("TauFadeIn:", x)
}

func (c *Controls) UpdateCropThresh(x float64) {
	c.CropThresh = x
	Println("CropThresh:", x)
	c.sampler.UpdateCropThresh()
}

func (c *Controls) UpdateRmsTime(x float64) {
	c.RmsTime = x
	Println("RmsTime:", x)
	c.sampler.UpdateRms()
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
