package dsp

import (
	"fmt"
	"math"
)

// Spectrogram windows each frame, runs it through fft, and returns the
// per-frame magnitude spectrum. Frames are left untouched: each frame is
// copied into a reused buffer before the window is applied.
//
// Every frame must have the same length as window. The returned rows share
// one backing array; each row's capacity is capped at its length, so an
// append on one row cannot overwrite the next.
func Spectrogram(frames [][]float64, window []float64, fft FFT) ([][]float64, error) {
	spectrogram := make([][]float64, len(frames))
	if len(frames) == 0 {
		return spectrogram, nil
	}

	// Reused for every frame instead of allocating per frame.
	windowed := make([]float64, len(window))
	var coeffs []complex128
	var backing []float64

	for i, frame := range frames {
		if len(frame) != len(window) {
			return nil, fmt.Errorf(
				"frame %d and window lengths should be same: %d != %d",
				i,
				len(frame),
				len(window),
			)
		}

		copy(windowed, frame)
		if err := applyWindow(windowed, window); err != nil {
			return nil, err
		}

		// The first call allocates the FFT output; later calls reuse it.
		coeffs = fft.Forward(coeffs, windowed)

		// The bin count comes from the FFT itself, so it holds for any
		// implementation. Size the shared backing array once it is known.
		bins := len(coeffs)
		if backing == nil {
			backing = make([]float64, len(frames)*bins)
		}

		row := backing[i*bins : (i+1)*bins : (i+1)*bins]
		spectrogram[i] = Magnitude(row, coeffs)
	}

	return spectrogram, nil
}

// Magnitude returns the magnitude of each bin in spectrum. If dst is nil, it
// allocates a new slice. Otherwise dst must have len(spectrum) elements, and
// Magnitude writes into it and returns it.
func Magnitude(dst []float64, spectrum []complex128) []float64 {
	if dst == nil {
		dst = make([]float64, len(spectrum))
	}
	for i, c := range spectrum {
		dst[i] = math.Sqrt(real(c)*real(c) + imag(c)*imag(c)) // instead of computing `cmplx.Abs(c)`
	}
	return dst
}
