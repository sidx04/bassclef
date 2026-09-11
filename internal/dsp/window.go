package dsp

import (
	"fmt"
	"math"
)

func HannWindow(size int) ([]float64, error) {
	if size <= 0 {
		return nil, fmt.Errorf("window size cannot be negative: %d", size)
	}
	window := make([]float64, size)

	if size == 1 {
		window[0] = 1
		return window, nil
	}

	for i := range window {
		window[i] = 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(size-1)))
	}

	return window, nil
}

func ApplyWindow(samples []float64, window []float64) error {
	if len(samples) != len(window) {
		return fmt.Errorf(
			"sample and window lengths should be same: %d != %d",
			len(samples),
			len(window),
		)
	}

	for i := range samples {
		samples[i] *= window[i]
	}

	return nil
}
