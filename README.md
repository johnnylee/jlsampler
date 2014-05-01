jlsampler
=========

Sampler for real-time use written in go. 

# Usage

JLSampler is called on a directory having the following structure:
```
sampler_name/   
    samples/    # Directory containing samples. 
    defaults.js # Javascript file containing with default settings. 
    tuning.js   # Javascript file containing tuning for each file. 
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
the future, but I personally haven't yet found a need for them. 

## Controls
 
The following controls are available: 
<dl>
  <dt>Transpose (int)</dt>
  <dd>Added to midi note key values.</dd>

  <dt>Tau (float)</dt>
  <dd>Key-up decay time constant in seconds. 
  Try 0.1 for piano-like decay.</dd>
  
  <dt>TauCut (float)</dt>
  <dd>Decay time constant for a re-triggered key.</dd>
  
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
  place to start is 0.2.</dd>

  <dt>RmsHigh (float)</dt>
  <dd>Peak RMS value for key 108 (High C) on a keyboard. For a piano a good 
  place to start is 0.04.</dd>

  <dt>PanLow (float)</dt>
  <dd>Pan value for key 21 on 88 key keyboard (low A). -1 is hard left and 1 
  is hard right.</dd>

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
</dl>