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

