package dsp

import (
	"math"
	"testing"
)

func TestHannWindow(t *testing.T) {
	window, err := HannWindow(5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []float64{
		0,
		0.5,
		1,
		0.5,
		0,
	}

	const epsilon = 1e-9

	for i := range window {
		if math.Abs(window[i]-want[i]) > epsilon {
			t.Errorf(
				"index %d: got %f, want %f",
				i,
				window[i],
				want[i],
			)
		}
	}
}

func TestApplyWindow(t *testing.T) {
	samples := []float64{
		1, 2, 3,
	}

	window := []float64{
		0.5, 1, 0.5,
	}

	err := ApplyWindow(samples, window)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []float64{
		0.5,
		2,
		1.5,
	}

	for i := range samples {
		if samples[i] != want[i] {
			t.Errorf(
				"index %d: got %f, want %f",
				i,
				samples[i],
				want[i],
			)
		}
	}
}
