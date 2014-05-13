package jlsampler

import (
	"math"
)

// freq3db: Compute the 3db point for a low pass filter of the given order.
func freq3db(freq float64, order int) float64 {
	return freq / math.Sqrt(math.Pow(2, 1.0/float64(order)) - 1)
}

func rcLowPass1(x []int16, freq float64) {
	y := make([]float64, len(x))

	dt := float64(1.0 / sampleRate)
	rc := 1.0 / (2.0 * math.Pi * freq)
	alpha := dt / (rc + dt)
	
	ymax := float64(0)
	prev := float64(0)
	
	for i, _ := range x {
		prev *= (1 - alpha)
		prev += alpha * float64(x[i])
		if prev > ymax {
			ymax = prev
		} else if -prev > ymax {
			ymax = -prev
		}
		y[i] = prev
	}
	
	for i, _ := range x {
		x[i] = int16(maxVal16 * y[i] / ymax)
	}
}

func rcLowPass(x []int16, freq float64, order int) {
	freq = freq3db(freq, order)
	for i := 0; i < order; i++ {
		rcLowPass1(x, freq)
	}
}

