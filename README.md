JLSampler
=========

A sampler for real-time use written in go. JLSampler uses alsa's midi sequencer
API for capturing midi events, and uses jack for output.

JLSampler is designed for live playing using a midi keyboard. To keep latency
low and avoid drop-outs, all samples are loaded into memory when the program
is launched. This might seem like overkill, but in practice, even a popular
sampled piano with 88 keys in 13 velocity layers is only around 10G in RAM.

I can get reliable playback on my laptop using an Alesis iO2 express USB
interface with 4ms latency.

## Gui

There is a basic GUI in the `gui` subfolder, written in Python. It only uses
tkinter, so it should run without any additional python packages installed.

## Usage

JLSampler is called on a directory having the following structure:
```
sampler_name/
    samples/    # Directory containing samples.
    defaults.js # JSON file containing with default settings.
    tuning.js   # JSON file containing tuning for each file.
```

## Samples

Samples currently must be 16-bit stereo 48 kHz flac files. Samples are named
with their key (midi note number), velocity layer (1 is softest), and 
round-robbin variation. All samples should be placed directly in the samples 
folder. For example the lowest key on an 88-key piano with four layers would
have filenames: 
```
on-021-001-001.flac
on-021-002-001.flac
on-021-003-001.flac
on-021-004-001.flac
```

Currently only key-on samples are supported. Key-off samples may be added in 
the future, but I haven't yet found a need for them. 

## Controls
 
Most controls can be modified in real time from the command line, or 
assigned to a midi control.

The following controls are available: 
<dl>
  <dt>Transpose (int)</dt>
  <dd>Added to midi note key values.</dd>

  <dt>Tau (float)</dt>
  <dd>Key-up decay time constant in seconds. 
  Try 0.1 for piano-like decay.</dd>
  
  <dt>TauCut (float)</dt>
  <dd>Decay time constant for a re-triggered key. 0 disables.</dd>
  
  <dt>CropThresh (float)</dt>
  <dd>Trim values below CropThresh from the beginning of samples.
  Acceptable range is [0,1].</dd>
  
  <dt>CropFade (float)</dt>
  <dd>Fade-in time in seconds. This can be used to get rid of clicks, 
  especially for cropped samples. Despite the name, the fade time applies
  to all samples, not just cropped samples.</dd>
  
  <dt>RmsTime (float)</dt>
  <dd>Time period from start of sample in seconds used to compute the sample's
  RMS value. This is used to normalize the amplitude of samples to provide 
  smooth transitions between velocity layers and across the keyboard.</dd>
  
  <dt>RmsLow (float)</dt>
  <dd>Peak RMS value for key 21 (Low A) on a keyboard. For a piano a good 
  place to start is 0.2. The RMS value for each note is linearly interpolated 
  across the keyboard. This provides a smooth response when moving from low
  to high notes on the keyboard, and matches fairly well with what I've 
  measured from a yamaha digital piano.</dd>

  <dt>RmsHigh (float)</dt>
  <dd>Peak RMS value for key 108 (High C) on a keyboard. For a piano a good 
  place to start is 0.04.</dd>

  <dt>PanLow (float)</dt>
  <dd>Pan value for key 21 on 88 key keyboard (low A). -1 is hard left and 1 
  is hard right. The pan value for each note is linearly interpolated from the 
  `PanLow` and `PanHigh` values.</dd>

  <dt>PanHigh (float)</dt>
  <dd>Like PanLow but for key 108 (high C).</dd>

  <dt>GammaAmp (float)</dt>
  <dd>The scaling of volume with key velocity. If velocity is scaled from zero 
  to one, then the amplitude is scaled like velocity^gamma. I've found 2.2 to 
  be a good starting value. I found that value by measuing the amplitude -vs- 
  velocity curve of a Yamaha stage paino.</dd>
  
  <dt>GammaLayer (float)</dt>
  <dd>Just like GammaAmp, but for selecting the velocity layer (sample).
  This will really depend on how the samples were captured. I've seen sample 
  sets with gamma on both sides of 1. </dd>
  
  <dt>VelMult (float)</dt>
  <dd>Multiplier for incoming midi velocity. I have one keyboard that I have 
  a very hard time reaching velocity levels of 100 out of 127.</dd>
  
  <dt>PitchBendMax (int)</dt>
  <dd>Maximum number of semitones of pitch bend available.</dd>
  
  <dt>RRBorrow (bool)</dt>
  <dd>If true, the sampler will pitch neighboring samples and use them as 
  round-robbins. This can be useful for avoiding a "machine gun" sound from a 
  sample set with a single round-robbin. NOTE: this is applied at load 
  time.</dd>
  
  <dt>MixLayer (bool)</dt>
  <dd>If true, the sampler will mix smoothly between velocity layers. This
  can be useful for certain types of prepared samples.</dd>
  
  <dt>Sustain (bool)</dt>
  <dd>This should be connected to your sustain pedal.</dd>
</dl>

## Configuration files

Configuration files are stored in `~/.jlsampler`. Currently there are two 
files: `config.js` and `controls.js`. These are both JSON files. 

### config.js

`config.js` contains the basic configuration variables. 
```
{
    "Procs": 4,
    "Poly": 16,
    "MidiIn": "20:0",
    "MidiBufSize": 32
}
```     
<dl>
  <dt>Procs</dt>
  <dd>The number of processors to use.</dd>

  <dt>Poly</dt>
  <dd>The single-key polyphony. The is the number of simultaneous playing 
  sounds available to each key. On an 88 key keyboard the max polyphony is 
  88 * Poly. You'll likely run out of CPU power when your actual polyphony 
  gets to a few hundred.</dd>
  
  <dt>MidiIn</dt>
  <dd>The midi client and port to use for midi input. JLSampler uses alsa's
  midi sequencer api. You can find your available clients and ports by calling 
  `aconnect -io`. This setting will look like `<client>:<port>`, for example
  `24:0`.</dd>
  
  <dt>MidiBufSize</dt>
  <dd>The size of the internal midi event buffer. I've never had a problem
  with the setting of 32. If this is too low it may be possible that some 
  midi events are dropped, I'm note sure.</dd>
</dl>

### controls.js

`controls.js` maps midi controls to the controls listed above. Below is an
example file with two bindings. Other controls are bound in the same manner. 
```
{
    "Sustain": {
        "Num": 64,
        "Min": 0,
        "Max": 1,
        "Gamma": 1.0
    },
    "PitchBendMax": {
        "Num": 11,
        "Min": 1,
        "Max": 12,
        "Gamma": 1
    }
}
```