package dsp

import (
	"math"
	"testing"
)

func TestFFTForwardImpulse(t *testing.T) {
	fft := NewFFT(4)

	input := []float64{1, 0, 0, 0}

	got := fft.Forward(input)

	want := []complex128{1, 1, 1, 1}

	const epsilon = 1e-9

	for i := range got {
		if math.Abs(real(got[i])-real(want[i])) > epsilon ||
			math.Abs(imag(got[i])-imag(want[i])) > epsilon {
			t.Errorf(
				"bin %d: got %v, want %v",
				i,
				got[i],
				want[i],
			)
		}
	}
}

func TestFFTForwardSineWave(t *testing.T) {
	const size = 8

	input := make([]float64, size)
	for i := range input {
		input[i] = math.Sin(2 * math.Pi * float64(i) / float64(size))
	}

	fft := NewFFT(size)

	got := fft.Forward(input)

	const epsilon = 1e-9

	for i, c := range got {
		mag := math.Hypot(real(c), imag(c))
		if i == 1 || i == 7 {
			if math.Abs(mag-4) > epsilon {
				t.Errorf("bin %d: got magnitude %v, want %v", i, mag, 4.0)
			}
			continue
		}
		if mag > epsilon {
			t.Errorf("bin %d: got magnitude %v, want ~0", i, mag)
		}
	}
}
