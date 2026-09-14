package fingerprint

import "testing"

func TestDefaultPeakConfig(t *testing.T) {
	got := DefaultPeakConfig()
	want := PeakConfig{FreqWindow: 3, TimeWindow: 3, MinMagnitude: 0}

	if got != want {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestDefaultLandmarkConfig(t *testing.T) {
	got := DefaultLandmarkConfig()
	want := LandmarkConfig{MinDeltaFrames: 1, MaxDeltaFrames: 50, Fanout: 10}

	if got != want {
		t.Errorf("got %#v, want %#v", got, want)
	}
}
