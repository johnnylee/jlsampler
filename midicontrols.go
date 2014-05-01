package jlsampler

import (
	"encoding/json"
	"math"
	"os"
	"os/user"
	"path/filepath"
)

// ----------------------------------------------------------------------------
// Load controls from file.
type ControlConfig struct {
	Num   int
	Min   float64
	Max   float64
	Gamma float64
}

type ControlConfigFile struct {
	Transpose    ControlConfig
	Tau          ControlConfig
	TauCut       ControlConfig
	CropThresh   ControlConfig
	CropFade     ControlConfig
	RmsTime      ControlConfig
	RmsLow       ControlConfig
	RmsHigh      ControlConfig
	PanLow       ControlConfig
	PanHigh      ControlConfig
	GammaAmp     ControlConfig
	GammaLayer   ControlConfig
	VelMult      ControlConfig
	PitchBendMax ControlConfig
	MixLayers    ControlConfig
	PrintLatency ControlConfig
	Sustain      ControlConfig
}

func wrapControlFn(f func(float64), min, max, gamma float64) func(float64) {
	return func(x float64) {
		f(min + (max-min)*math.Pow(x, gamma))
	}
}

func setMidiControl(cc ControlConfig, updateFn func(float64)) {
	if cc.Num >= 0 && cc.Num < 128 && cc.Gamma != 0 {
		midiControls[cc.Num] = wrapControlFn(
			updateFn, cc.Min, cc.Max, cc.Gamma)
	}
}

func LoadMidiControls() error {
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
	cfg := new(ControlConfigFile)
	if err = decoder.Decode(cfg); err != nil {
		return err
	}

	setMidiControl(cfg.Transpose, controls.UpdateTranspose)
	setMidiControl(cfg.Tau, controls.UpdateTau)
	setMidiControl(cfg.TauCut, controls.UpdateTauCut)
	setMidiControl(cfg.CropThresh, controls.UpdateCropThresh)
	setMidiControl(cfg.CropFade, controls.UpdateCropFade)
	setMidiControl(cfg.RmsTime, controls.UpdateRmsTime)
	setMidiControl(cfg.RmsLow, controls.UpdateRmsLow)
	setMidiControl(cfg.RmsHigh, controls.UpdateRmsHigh)
	setMidiControl(cfg.PanLow, controls.UpdatePanLow)
	setMidiControl(cfg.PanHigh, controls.UpdatePanHigh)
	setMidiControl(cfg.GammaAmp, controls.UpdateGammaAmp)
	setMidiControl(cfg.GammaLayer, controls.UpdateGammaLayer)
	setMidiControl(cfg.VelMult, controls.UpdateVelMult)
	setMidiControl(cfg.PitchBendMax, controls.UpdatePitchBendMax)
	setMidiControl(cfg.MixLayers, controls.UpdateMixLayers)
	setMidiControl(cfg.PrintLatency, controls.UpdatePrintLatency)
	setMidiControl(cfg.Sustain, controls.UpdateSustain)
	return nil
}
