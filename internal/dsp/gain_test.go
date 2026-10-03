package dsp

import (
	"math"
	"testing"
)

func TestNormalizeGain(t *testing.T) {
	samples := []float64{0.1, -0.05, 0.02, -0.025}

	NormalizeGain(samples)

	want := []float64{1.0, -0.5, 0.2, -0.25}

	const epsilon = 1e-9

	for i := range samples {
		if math.Abs(samples[i]-want[i]) > epsilon {
			t.Errorf("index %d: got %v, want %v", i, samples[i], want[i])
		}
	}
}

func TestNormalizeGainAlreadyFullScale(t *testing.T) {
	samples := []float64{1.0, -0.5, 0.25}

	NormalizeGain(samples)

	want := []float64{1.0, -0.5, 0.25}

	const epsilon = 1e-9

	for i := range samples {
		if math.Abs(samples[i]-want[i]) > epsilon {
			t.Errorf("index %d: got %v, want %v", i, samples[i], want[i])
		}
	}
}

func TestNormalizeGainSilence(t *testing.T) {
	samples := []float64{0, 0, 0}

	NormalizeGain(samples)

	want := []float64{0, 0, 0}

	for i := range samples {
		if samples[i] != want[i] {
			t.Errorf("index %d: got %v, want %v", i, samples[i], want[i])
		}
	}
}

func TestNormalizeGainEmpty(t *testing.T) {
	samples := []float64{}

	NormalizeGain(samples)

	if len(samples) != 0 {
		t.Errorf("got %v, want empty", samples)
	}
}
