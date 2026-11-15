package dsp

import (
	"math"
	"testing"
)

func TestFFTForwardImpulse(t *testing.T) {
	fft := NewFFT(4)

	input := []float64{1, 0, 0, 0}

	got := fft.Forward(nil, input)

	// A size-4 real input has 4/2+1 = 3 non-redundant bins.
	want := []complex128{1, 1, 1}

	if len(got) != len(want) {
		t.Fatalf("got %d bins, want %d", len(got), len(want))
	}

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

	got := fft.Forward(nil, input)

	// Bins are in index 0 to size/2 only
	// the mirror of bin 1 (bin 7) is not returned
	if len(got) != size/2+1 {
		t.Fatalf("got %d bins, want %d", len(got), size/2+1)
	}

	const epsilon = 1e-9

	for i, c := range got {
		mag := math.Hypot(real(c), imag(c))
		if i == 1 {
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

func TestFFTForwardReusesDst(t *testing.T) {
	const size = 8

	fft := NewFFT(size)
	dst := make([]complex128, size/2+1)

	got := fft.Forward(dst, make([]float64, size))

	if &got[0] != &dst[0] {
		t.Error("Forward allocated instead of writing into dst")
	}
}
