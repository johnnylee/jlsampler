JLSampler
=========

See http://www.johnnylee.org/jlsampler for details. 

A sampler for real-time use written in go. JLSampler uses alsa's midi sequencer
API for capturing midi events, and uses jack for output.

JLSampler is designed for live playing using a midi keyboard. To keep latency
low and avoid drop-outs, all samples are loaded into memory when the program
is launched. This might seem like overkill, but in practice, even a popular
sampled piano with 88 keys in 13 velocity layers is only around 10G in RAM.

I can get reliable playback on my laptop using an Alesis iO2 express USB
interface with 4ms latency.
