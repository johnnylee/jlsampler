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
  <dd>Fade-in time for samples. This can be used to get rid of clicks, 
  especially for cropped samples. Despite the name, the fade time applies
  to all samples, not just cropped samples.</dd>
  
</dl>