package dsp

import "math/cmplx"

// Spectrogram windows each frame, runs it through fft, and returns the
// per-frame magnitude spectrum. Frames are left untouched; ApplyWindow
// runs against a copy.
func Spectrogram(frames [][]float64, window []float64, fft FFT) ([][]float64, error) {
	spectrogram := make([][]float64, len(frames))

	for i, frame := range frames {
		windowed := make([]float64, len(frame))
		copy(windowed, frame)

		if err := ApplyWindow(windowed, window); err != nil {
			return nil, err
		}

		spectrogram[i] = Magnitude(fft.Forward(windowed))
	}

	return spectrogram, nil
}

func Magnitude(spectrum []complex128) []float64 {
	magnitudes := make([]float64, len(spectrum))
	for i, c := range spectrum {
		magnitudes[i] = cmplx.Abs(c)
	}
	return magnitudes
}
