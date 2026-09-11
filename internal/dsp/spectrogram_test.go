package dsp

import (
	"math"
	"testing"
)

type stubFFT struct{}

func (stubFFT) Forward(input []float64) []complex128 {
	out := make([]complex128, len(input))
	for i, v := range input {
		out[i] = complex(v, 0)
	}
	return out
}

func TestSpectrogram(t *testing.T) {
	frames := [][]float64{
		{1, 2, 3},
		{4, 5, 6},
	}

	window := []float64{1, 1, 1}

	got, err := Spectrogram(frames, window, stubFFT{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := [][]float64{
		{1, 2, 3},
		{4, 5, 6},
	}

	const epsilon = 1e-9

	for i := range want {
		for j := range want[i] {
			if math.Abs(got[i][j]-want[i][j]) > epsilon {
				t.Errorf("frame %d bin %d: got %v, want %v", i, j, got[i][j], want[i][j])
			}
		}
	}
}

func TestSpectrogramWindowMismatch(t *testing.T) {
	frames := [][]float64{
		{1, 2, 3},
	}

	window := []float64{1, 1}

	_, err := Spectrogram(frames, window, stubFFT{})
	if err == nil {
		t.Fatal("expected an error")
	}
}
