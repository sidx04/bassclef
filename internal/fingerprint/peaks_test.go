package fingerprint

import (
	"reflect"
	"testing"
)

func gridOf(frames, bins int, fill float64) [][]float64 {
	grid := make([][]float64, frames)
	for t := range grid {
		grid[t] = make([]float64, bins)
		for f := range grid[t] {
			grid[t][f] = fill
		}
	}
	return grid
}

func TestFindPeaksSinglePeak(t *testing.T) {
	spectrogram := gridOf(7, 7, 1)
	spectrogram[3][3] = 10

	cfg := PeakConfig{FreqWindow: 3, TimeWindow: 3, MinMagnitude: 5}

	got, err := FindPeaks(spectrogram, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []Peak{{Time: 3, Frequency: 3, Magnitude: 10}}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestFindPeaksTwoFarApart(t *testing.T) {
	spectrogram := gridOf(14, 7, 1)
	spectrogram[3][3] = 10
	spectrogram[10][3] = 10

	cfg := PeakConfig{FreqWindow: 3, TimeWindow: 3, MinMagnitude: 5}

	got, err := FindPeaks(spectrogram, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []Peak{
		{Time: 3, Frequency: 3, Magnitude: 10},
		{Time: 10, Frequency: 3, Magnitude: 10},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestFindPeaksBelowThreshold(t *testing.T) {
	spectrogram := gridOf(7, 7, 1)
	spectrogram[3][3] = 3

	cfg := PeakConfig{FreqWindow: 3, TimeWindow: 3, MinMagnitude: 5}

	got, err := FindPeaks(spectrogram, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 0 {
		t.Errorf("got %#v, want no peaks", got)
	}
}

func TestFindPeaksEdgeOfGrid(t *testing.T) {
	spectrogram := gridOf(4, 4, 1)
	spectrogram[0][0] = 10

	cfg := PeakConfig{FreqWindow: 3, TimeWindow: 3, MinMagnitude: 5}

	got, err := FindPeaks(spectrogram, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []Peak{{Time: 0, Frequency: 0, Magnitude: 10}}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestFindPeaksInvalidConfig(t *testing.T) {
	spectrogram := gridOf(7, 7, 1)

	tests := []struct {
		name string
		cfg  PeakConfig
	}{
		{"zero freq window", PeakConfig{FreqWindow: 0, TimeWindow: 3, MinMagnitude: 5}},
		{"negative freq window", PeakConfig{FreqWindow: -1, TimeWindow: 3, MinMagnitude: 5}},
		{"zero time window", PeakConfig{FreqWindow: 3, TimeWindow: 0, MinMagnitude: 5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FindPeaks(spectrogram, tt.cfg)
			if err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestFindPeaksEmptySpectrogram(t *testing.T) {
	cfg := PeakConfig{FreqWindow: 3, TimeWindow: 3, MinMagnitude: 5}

	_, err := FindPeaks(nil, cfg)
	if err == nil {
		t.Fatal("expected an error")
	}
}
